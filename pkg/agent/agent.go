package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"

	"github.com/SchSeba/slack-ai-assistant/pkg/database"
	"github.com/SchSeba/slack-ai-assistant/pkg/llm"
	slackbot "github.com/SchSeba/slack-ai-assistant/pkg/slack-bot"
)

type Agent struct {
	db                  database.Interface
	appMentionChannel   chan (*slackevents.AppMentionEvent)
	slashCommandChannel chan (*slack.SlashCommand)
	slackBot            slackbot.Interface
	llmClient           llm.Interface
	workerPool          *WorkerPool
}

func NewAgent(db database.Interface, slackBot slackbot.Interface, llmClient llm.Interface, appMentionChannel chan (*slackevents.AppMentionEvent), slashCommandChannel chan (*slack.SlashCommand), workerCount int) *Agent {
	// Create worker pool with configurable size
	// Queue size is set to 200 to handle bursts of events
	workerPool := NewWorkerPool(workerCount, 200)
	
	return &Agent{
		db:                  db,
		slackBot:            slackBot,
		llmClient:           llmClient,
		appMentionChannel:   appMentionChannel,
		slashCommandChannel: slashCommandChannel,
		workerPool:          workerPool,
	}
}

func (a *Agent) Start(ctx context.Context) {
	// Start the worker pool
	a.workerPool.Start(a)
	
	// Start the dispatcher goroutine that reads from channels and submits work
	go func() {
		defer a.workerPool.Stop()
		for {
			select {
			case event := <-a.appMentionChannel:
				workItem := AppMentionWorkItem{Event: event}
				a.workerPool.Submit(workItem)
			case <-ctx.Done():
				fmt.Println("🛑 Agent dispatcher shutting down...")
				return
			}
		}
	}()

	a.slackBot.Start(ctx)
}

// handleAppMentionEvent is the internal implementation called by worker pool
func (a *Agent) handleAppMentionEvent(event *slackevents.AppMentionEvent) error {
	botUser := a.slackBot.GetBotUser()
	fmt.Printf("🏷️ Bot mentioned: %s from user %s in channel %s\n",
		event.Text, event.User, event.Channel)

	// Extract bot's username and ID
	botUsername := "slack-ai-assistant"
	botUserID := ""
	if botUser != nil {
		botUsername = botUser.User
		botUserID = botUser.UserID
	}
	fmt.Printf("🤖 Bot info - Username: %s, ID: %s\n", botUsername, botUserID)

	// Determine the thread timestamp
	var threadTS string
	if event.ThreadTimeStamp != "" {
		// This is already in a thread, use the existing thread timestamp
		threadTS = event.ThreadTimeStamp
		fmt.Printf("📎 Message is in an existing thread: %s\n", threadTS)
	} else {
		// This is a new message, use its timestamp to create a new thread
		threadTS = event.TimeStamp
		fmt.Printf("🆕 Creating new thread with timestamp: %s\n", threadTS)
	}

	// Check if we have parameters in the message
	parameters := strings.Split(event.Text, " ")
	command := ""
	if len(parameters) > 1 {
		fmt.Printf("🔍 Parameters: %v\n", parameters)
		command = parameters[1]
	}
    
	switch command {
	case "answer":
		if len(parameters) > 2 && len(parameters) < 4 {
			return a.slackBot.PostMessage(event.Channel,threadTS, "To answer the question please provide the project name (example: sriov,metallb) and the openshift version (4.16,4.18, etc..)")
		}
		return a.AnswerQuestion(event.Channel,threadTS,parameters[2], parameters[3],false)
	case "answer-all":
		if len(parameters) > 2 && len(parameters) < 4 {
			return a.slackBot.PostMessage(event.Channel,threadTS, "To answer the question please provide the project name (example: sriov,metallb) and the openshift version (4.16,4.18, etc..)")
		}
		return a.AnswerQuestion(event.Channel,threadTS, parameters[2], parameters[3],true)
	case "inject":
		if len(parameters) > 2 && len(parameters) < 4 {
			return a.slackBot.PostMessage(event.Channel,threadTS, "To inject the last message in the thread please provide the project name (example: sriov,metallb) and the openshift version (4.16,4.18, etc..)")
		}
		return a.Inject(event.Channel,threadTS,parameters[2], parameters[3])
	case "elaborate":
		return a.Elaborate(event.Channel,threadTS)
	}


	return a.slackBot.PostMessage(event.Channel,threadTS, "Please use one of the following commands (answer,elaborate,inject)")
}

func (a *Agent) AnswerQuestion(channel, threadTS, project, version string, fullThread bool) error {
	err := a.slackBot.PostMessage(channel,threadTS, "Searching for answer...")
	if err != nil {
		return fmt.Errorf("failed to post initial message: %w", err)
	}
	var messages = ""
	if fullThread {
		// Get all thread messages
		messages, err = a.getThreadMessages(channel, threadTS)
		if err != nil {
			fmt.Printf("❌ Failed to get thread messages: %v\n", err)
			return fmt.Errorf("failed to get thread messages: %w", err)
		}
	} else {
		messages, err = a.getLastMessageInThread(channel, threadTS)
		if err != nil {
			fmt.Printf("❌ Failed to get last message in thread: %v\n", err)
			return fmt.Errorf("failed to get last message in thread: %w", err)
		}
	}

	// Check if a slug in anythingllm already exist
	slug,exist, err := a.db.GetSlugForThread(threadTS)
	if err != nil {
		fmt.Printf("❌ Failed to get slug for thread from database: %v\n", err)
		return fmt.Errorf("failed to get slug for thread from database: %w", err)
	}

	if !exist {
		slug, err = a.llmClient.CreateThread(project, version)
		if err != nil {
			fmt.Printf("❌ Failed to create thread: %v\n", err)
			return fmt.Errorf("failed to create thread: %w", err)
		}
		err = a.db.CreateSlackThreadWithSlug(threadTS, slug)
		if err != nil {
			fmt.Printf("❌ Failed to create slack thread in database: %v\n", err)
			return fmt.Errorf("failed to create slack thread in database: %w", err)
		}
	}

	response, err := a.llmClient.SendMessageToChat(project, version,slug, messages)
	if err != nil {
		fmt.Printf("❌ Failed to generate response: %v\n", err)
		return fmt.Errorf("failed to generate response: %w", err)
	}

	err = a.slackBot.PostMessage(channel,threadTS, fmt.Sprintf("Here is the information I was able to find\n%s", response))
	if err != nil {
		return fmt.Errorf("failed to send response: %w", err)
	}
	return nil
}

func (a *Agent) Elaborate(channel, threadTS string) error {
	err := a.slackBot.PostMessage(channel,threadTS, "Elaborating...")
	if err != nil {
		return fmt.Errorf("failed to post initial message: %w", err)
	}

	lastMessage, err := a.getLastMessageInThread(channel, threadTS)
	if err != nil {
		fmt.Printf("❌ Failed to get last message in thread: %v\n", err)
		return fmt.Errorf("failed to get last message in thread: %w", err)
	}
	slug, err := a.llmClient.CreateThread("elaborate","")
	if err != nil {
		fmt.Printf("❌ Failed to create thread: %v\n", err)
		return fmt.Errorf("failed to create thread: %w", err)
	}

	response, err := a.llmClient.Elaborate(slug, lastMessage)
	if err != nil {
		fmt.Printf("❌ Failed to generate response: %v\n", err)
		return fmt.Errorf("failed to generate response: %w", err)
	}
	err = a.slackBot.PostMessage(channel,threadTS, fmt.Sprintf("%s", response))
	if err != nil {
		return fmt.Errorf("failed to send response: %w", err)
	}
	
	return nil
}

func (a *Agent) Inject(channel, threadTS, project, version string) error {
	messages, err := a.getLastMessagesFromTheSameUser(channel, threadTS)
		if err != nil {
			fmt.Printf("❌ Failed to get thread messages: %v\n", err)
			return fmt.Errorf("failed to get thread messages: %w", err)
		}
	
	err = a.llmClient.Inject(project, version, messages)
	if err != nil {
		fmt.Printf("❌ Failed to inject messages: %v\n", err)
		return fmt.Errorf("failed to inject messages: %w", err)
	}

	err = a.slackBot.PostMessage(channel,threadTS, fmt.Sprintf("Document injected for project %s on version %s", project, version))
	if err != nil {
		return fmt.Errorf("failed to send response: %w", err)
	}
	return nil
}

// getThreadMessages retrieves and returns all messages in a thread
func (a *Agent) getThreadMessages(channel, threadTS string) (string, error) {
	fmt.Printf("🧵 Retrieving thread messages for thread: %s\n", threadTS)

	// Get conversation replies (thread messages)
	replies, _, _, err := a.slackBot.GetAPI().GetConversationReplies(&slack.GetConversationRepliesParameters{
		ChannelID: channel,

		Timestamp: threadTS,
		Inclusive: true, // Include the parent message
	})

	if err != nil {
		fmt.Printf("❌ Failed to retrieve thread messages: %v\n", err)
		return "", err
	}

	fmt.Printf("📋 Thread contains %d message(s):\n", len(replies))
	messages := ""
	for _, msg := range replies {
		messages += fmt.Sprintf("%s\n", msg.Text)
	}
	fmt.Printf("📋 messages in thread:\n%s", messages)
	return messages, nil
}

func (a *Agent) getLastMessageInThread(channel, threadTS string) (string, error) {
	// Get conversation replies (thread messages)
	replies, _, _, err := a.slackBot.GetAPI().GetConversationReplies(&slack.GetConversationRepliesParameters{
		ChannelID: channel,

		Timestamp: threadTS,
		Inclusive: true, // Include the parent message
	})

	if err != nil {
		fmt.Printf("❌ Failed to retrieve thread messages: %v\n", err)
		return "", err
	}
	if len(replies) < 3 {
		return "", fmt.Errorf("unexpected number of messages in thread")
	}
	return replies[len(replies)-3].Text, nil
}

func (a *Agent) getLastMessagesFromTheSameUser(channel, threadTS string) (string, error) {
	replies, _, _, err := a.slackBot.GetAPI().GetConversationReplies(&slack.GetConversationRepliesParameters{
		ChannelID: channel,
		Timestamp: threadTS,
		Inclusive: true, // Include the parent message
	})

	if err != nil {
		fmt.Printf("❌ Failed to retrieve thread messages: %v\n", err)
		return "", err
	}

	lastMessageUser := replies[len(replies)-2].User
	messages := ""
	for index := len(replies)-2; index > 0; index-- {
		if replies[index].User != lastMessageUser {
			break
		}

		messages = fmt.Sprintf("%s%s",replies[index].Text,messages)
	}
	messages = strings.TrimPrefix(messages,"Elaborating...")
	return messages, nil
}