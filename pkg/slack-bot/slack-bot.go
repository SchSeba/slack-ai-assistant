// Package slackbot provides Slack API integration using Socket Mode for real-time WebSocket communication.
package slackbot

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

// Interface defines the contract for Slack bot operations
type Interface interface {
	// Start begins the bot's event processing loop
	Start(ctx context.Context)

	// PostMessage posts a message to a channel
	PostMessage(channel, threadTS, message string) error

	// GetConversationReplies gets replies in a conversation thread
	GetConversationReplies(params *slack.GetConversationRepliesParameters) ([]slack.Message, error)

	// GetBotUser returns the bot user information
	GetBotUser() *slack.AuthTestResponse
}

type SlackBot struct {
	api                 *slack.Client
	socketMode          *socketmode.Client
	botUser             *slack.AuthTestResponse
	appMentionChannel   chan *slackevents.AppMentionEvent
	slashCommandChannel chan *slack.SlashCommand
}

func NewSlackBot(slackBotToken, slackAppToken string,
	appMentionChannel chan *slackevents.AppMentionEvent,
	slashCommandChannel chan *slack.SlashCommand,
	debug bool) (*SlackBot, error) {
	// Create a new Slack API client
	api := slack.New(
		slackBotToken,
		slack.OptionDebug(debug),
		slack.OptionLog(log.New(os.Stdout, "slack-bot: ", log.Lshortfile|log.LstdFlags)),
		slack.OptionAppLevelToken(slackAppToken),
	)

	// Create a new Socket Mode client
	socketMode := socketmode.New(
		api,
		socketmode.OptionDebug(debug),
		socketmode.OptionLog(log.New(os.Stdout, "socketmode: ", log.Lshortfile|log.LstdFlags)),
	)

	// Test the connection
	authTest, err := api.AuthTest()
	if err != nil {
		log.Fatalf("❌ Failed to authenticate with Slack: %v", err)
	}

	botUser := authTest // Store bot user info
	fmt.Printf("✅ Connected to Slack! Bot User: %s (ID: %s)\n", authTest.User, authTest.UserID)
	return &SlackBot{api: api, socketMode: socketMode, botUser: botUser, appMentionChannel: appMentionChannel, slashCommandChannel: slashCommandChannel}, nil
}

// Start begins the bot's event processing loop
//
//nolint:gocognit // this is a long function, but it is a good place to put the event handling logic
func (b *SlackBot) Start(ctx context.Context) {
	// Handle different types of events
	go func() {
		for envelope := range b.socketMode.Events {
			switch envelope.Type {
			case socketmode.EventTypeConnecting:
				fmt.Println("🔌 Connecting to Slack with Socket Mode...")

			case socketmode.EventTypeConnectionError:
				fmt.Printf("❌ Connection failed: %v\n", envelope.Data)

			case socketmode.EventTypeConnected:
				fmt.Println("✅ Connected to Slack with Socket Mode")
			case socketmode.EventTypeHello:
				fmt.Println("👋 Hello from Slack!")

			case socketmode.EventTypeEventsAPI:
				// Handle Events API events
				eventsAPIEvent, ok := envelope.Data.(slackevents.EventsAPIEvent)
				if !ok {
					fmt.Printf("❌ Unexpected event type: %v\n", envelope.Data)
					continue
				}

				// Acknowledge the event
				// TODO: Maybe we should not ack the event here, but in the handleAppMentionEvent and handleSlashCommand functions
				b.socketMode.Ack(*envelope.Request)
				appMentionEvent, ok := eventsAPIEvent.InnerEvent.Data.(*slackevents.AppMentionEvent)
				if !ok {
					fmt.Printf("❌ Unexpected app mention event type: %v\n", eventsAPIEvent.InnerEvent.Data)
					continue
				}
				b.appMentionChannel <- appMentionEvent

			case socketmode.EventTypeSlashCommand:
				// Handle slash commands
				command, ok := envelope.Data.(*slack.SlashCommand)
				if !ok {
					fmt.Printf("❌ Unexpected slash command type: %v\n", envelope.Data)
					continue
				}
				b.slashCommandChannel <- command

			default:
				fmt.Printf("🔍 Unhandled event type: %s\n", envelope.Type)
			}
		}
	}()

	fmt.Println("🤖 Slack AI Assistant Bot is running...")
	if err := b.socketMode.RunContext(ctx); err != nil {
		fmt.Printf("Error running socket mode: %v\n", err)
	}
}

func (b *SlackBot) PostMessage(channel, threadTS, message string) error {
	_, _, err := b.api.PostMessage(
		channel,
		slack.MsgOptionText(message, false),
		slack.MsgOptionTS(threadTS),
	)

	fmt.Printf("🔍 Posted message to channel %s in thread %s: %s\n", channel, threadTS, message)
	if err != nil {
		fmt.Printf("❌ Failed to post message: %v\n", err)
		return fmt.Errorf("failed to post message: %w", err)
	}
	return nil
}

// GetBotUser returns the bot user information
func (b *SlackBot) GetBotUser() *slack.AuthTestResponse {
	return b.botUser
}

// GetConversationReplies gets replies in a conversation thread
func (b *SlackBot) GetConversationReplies(params *slack.GetConversationRepliesParameters) ([]slack.Message, error) {
	replies, _, _, err := b.api.GetConversationReplies(params)
	return replies, err
}
