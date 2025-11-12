package agent_test

import (
	"context"
	"errors"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"go.uber.org/mock/gomock"

	"github.com/SchSeba/slack-ai-assistant/pkg/agent"
	databaseMock "github.com/SchSeba/slack-ai-assistant/pkg/mocks/database"
	llmMock "github.com/SchSeba/slack-ai-assistant/pkg/mocks/llm"
	slackbotMock "github.com/SchSeba/slack-ai-assistant/pkg/mocks/slack-bot"
)

var _ = Describe("Agent", func() {
	var (
		ctrl                *gomock.Controller
		mockDB              *databaseMock.MockInterface
		mockSlackBot        *slackbotMock.MockInterface
		mockLLM             *llmMock.MockInterface
		appMentionChannel   chan *slackevents.AppMentionEvent
		slashCommandChannel chan *slack.SlashCommand
		testAgent           *agent.Agent
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockDB = databaseMock.NewMockInterface(ctrl)
		mockSlackBot = slackbotMock.NewMockInterface(ctrl)
		mockLLM = llmMock.NewMockInterface(ctrl)

		appMentionChannel = make(chan *slackevents.AppMentionEvent, 10)
		slashCommandChannel = make(chan *slack.SlashCommand, 10)

		testAgent = agent.NewAgent(mockDB, mockSlackBot, mockLLM, appMentionChannel, slashCommandChannel, 1)
	})

	AfterEach(func() {
		close(appMentionChannel)
		close(slashCommandChannel)
		ctrl.Finish()
	})

	Describe("NewAgent", func() {
		It("should create a new agent with proper dependencies", func() {
			Expect(testAgent).NotTo(BeNil())
		})
	})

	Describe("AnswerQuestion", func() {
		var (
			channel  = "C1234567890"
			threadTS = "1234567890.123456"
			project  = "sriov"
			version  = "4.16"
		)

		Context("when thread does not exist in database", func() {
			It("should create new thread and answer question", func() {
				// Mock expectations
				mockSlackBot.EXPECT().PostMessage(channel, threadTS, "Searching for answer...").Return(nil)
				mockSlackBot.EXPECT().GetConversationReplies(gomock.Any()).Return([]slack.Message{
					{Msg: slack.Msg{Text: "User message 1"}},
					{Msg: slack.Msg{Text: "Bot response"}},
					{Msg: slack.Msg{Text: "User question"}},
				}, nil)
				mockDB.EXPECT().GetSlugForThread(threadTS).Return("", false, nil)
				mockLLM.EXPECT().CreateThread(project, version).Return("test-thread-slug", nil)
				mockDB.EXPECT().CreateSlackThreadWithSlug(threadTS, "test-thread-slug").Return(nil)
				mockLLM.EXPECT().SendMessageToChat(project, version, "test-thread-slug", gomock.Any()).Return("AI response", nil)
				mockSlackBot.EXPECT().PostMessage(channel, threadTS, gomock.Any()).Return(nil)

				err := testAgent.AnswerQuestion(channel, threadTS, project, version, false)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when thread exists in database", func() {
			It("should use existing thread slug", func() {
				existingSlug := "existing-thread-slug"

				mockSlackBot.EXPECT().PostMessage(channel, threadTS, "Searching for answer...").Return(nil)
				mockSlackBot.EXPECT().GetConversationReplies(gomock.Any()).Return([]slack.Message{
					{Msg: slack.Msg{Text: "User message 1"}},
					{Msg: slack.Msg{Text: "Bot response"}},
					{Msg: slack.Msg{Text: "User question"}},
				}, nil)
				mockDB.EXPECT().GetSlugForThread(threadTS).Return(existingSlug, true, nil)
				mockLLM.EXPECT().SendMessageToChat(project, version, existingSlug, gomock.Any()).Return("AI response", nil)
				mockSlackBot.EXPECT().PostMessage(channel, threadTS, gomock.Any()).Return(nil)

				err := testAgent.AnswerQuestion(channel, threadTS, project, version, false)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when database operation fails", func() {
			It("should return error when getting slug fails", func() {
				mockSlackBot.EXPECT().PostMessage(channel, threadTS, "Searching for answer...").Return(nil)
				mockSlackBot.EXPECT().GetConversationReplies(gomock.Any()).Return([]slack.Message{
					{Msg: slack.Msg{Text: "User message 1"}},
					{Msg: slack.Msg{Text: "Bot response"}},
					{Msg: slack.Msg{Text: "User question"}},
				}, nil)
				mockDB.EXPECT().GetSlugForThread(threadTS).Return("", false, errors.New("database error"))

				err := testAgent.AnswerQuestion(channel, threadTS, project, version, false)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to get slug for thread from database"))
			})

			It("should return error when creating thread fails", func() {
				mockSlackBot.EXPECT().PostMessage(channel, threadTS, "Searching for answer...").Return(nil)
				mockSlackBot.EXPECT().GetConversationReplies(gomock.Any()).Return([]slack.Message{
					{Msg: slack.Msg{Text: "User message 1"}},
					{Msg: slack.Msg{Text: "Bot response"}},
					{Msg: slack.Msg{Text: "User question"}},
				}, nil)
				mockDB.EXPECT().GetSlugForThread(threadTS).Return("", false, nil)
				mockLLM.EXPECT().CreateThread(project, version).Return("", errors.New("LLM error"))

				err := testAgent.AnswerQuestion(channel, threadTS, project, version, false)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to create thread"))
			})

			It("should return error when SendMessageToChat fails", func() {
				mockSlackBot.EXPECT().PostMessage(channel, threadTS, "Searching for answer...").Return(nil)
				mockSlackBot.EXPECT().GetConversationReplies(gomock.Any()).Return([]slack.Message{
					{Msg: slack.Msg{Text: "User message 1"}},
					{Msg: slack.Msg{Text: "Bot response"}},
					{Msg: slack.Msg{Text: "User question"}},
				}, nil)
				mockDB.EXPECT().GetSlugForThread(threadTS).Return("existing-slug", true, nil)
				mockLLM.EXPECT().SendMessageToChat(project, version, "existing-slug", gomock.Any()).Return("", errors.New("no index found"))
				mockSlackBot.EXPECT().PostMessage(channel, threadTS, "❌ Error: no index found").Return(nil)

				err := testAgent.AnswerQuestion(channel, threadTS, project, version, false)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to generate response"))
			})
		})
	})

	Describe("Elaborate", func() {
		var (
			channel  = "C1234567890"
			threadTS = "1234567890.123456"
		)

		It("should elaborate on the last message", func() {
			mockSlackBot.EXPECT().PostMessage(channel, threadTS, "Elaborating...").Return(nil)
			mockSlackBot.EXPECT().GetConversationReplies(gomock.Any()).Return([]slack.Message{
				{Msg: slack.Msg{Text: "User message 1"}},
				{Msg: slack.Msg{Text: "Bot response"}},
				{Msg: slack.Msg{Text: "User question"}},
			}, nil)
			mockLLM.EXPECT().CreateThread("elaborate", "").Return("elaborate-thread-slug", nil)
			mockLLM.EXPECT().Elaborate("elaborate-thread-slug", gomock.Any()).Return("Elaborated response", nil)
			mockSlackBot.EXPECT().PostMessage(channel, threadTS, "Elaborated response").Return(nil)

			err := testAgent.Elaborate(channel, threadTS)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should handle LLM create thread failure", func() {
			mockSlackBot.EXPECT().PostMessage(channel, threadTS, "Elaborating...").Return(nil)
			mockSlackBot.EXPECT().GetConversationReplies(gomock.Any()).Return([]slack.Message{
				{Msg: slack.Msg{Text: "User message 1"}},
				{Msg: slack.Msg{Text: "Bot response"}},
				{Msg: slack.Msg{Text: "User question"}},
			}, nil)
			mockLLM.EXPECT().CreateThread("elaborate", "").Return("", errors.New("LLM error"))

			err := testAgent.Elaborate(channel, threadTS)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to create thread"))
		})

		It("should handle LLM elaborate failure", func() {
			mockSlackBot.EXPECT().PostMessage(channel, threadTS, "Elaborating...").Return(nil)
			mockSlackBot.EXPECT().GetConversationReplies(gomock.Any()).Return([]slack.Message{
				{Msg: slack.Msg{Text: "User message 1"}},
				{Msg: slack.Msg{Text: "Bot response"}},
				{Msg: slack.Msg{Text: "User question"}},
			}, nil)
			mockLLM.EXPECT().CreateThread("elaborate", "").Return("elaborate-thread-slug", nil)
			mockLLM.EXPECT().Elaborate("elaborate-thread-slug", gomock.Any()).Return("", errors.New("elaboration failed"))
			mockSlackBot.EXPECT().PostMessage(channel, threadTS, "❌ Error: elaboration failed").Return(nil)

			err := testAgent.Elaborate(channel, threadTS)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to generate response"))
		})
	})

	Describe("Inject", func() {
		var (
			channel  = "C1234567890"
			threadTS = "1234567890.123456"
			project  = "sriov"
			version  = "4.16"
		)

		It("should inject messages successfully", func() {
			mockSlackBot.EXPECT().GetConversationReplies(gomock.Any()).Return([]slack.Message{
				{Msg: slack.Msg{Text: "User message 1", User: "U123"}},
				{Msg: slack.Msg{Text: "Bot response", User: "BOT123"}},
				{Msg: slack.Msg{Text: "User question", User: "U123"}},
			}, nil)
			mockLLM.EXPECT().Inject(project, version, gomock.Any()).Return(nil)
			mockSlackBot.EXPECT().PostMessage(channel, threadTS, "Document injected for project sriov on version 4.16").Return(nil)

			err := testAgent.Inject(channel, threadTS, project, version)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should handle injection failure", func() {
			mockSlackBot.EXPECT().GetConversationReplies(gomock.Any()).Return([]slack.Message{
				{Msg: slack.Msg{Text: "User message 1", User: "U123"}},
				{Msg: slack.Msg{Text: "Bot response", User: "BOT123"}},
				{Msg: slack.Msg{Text: "User question", User: "U123"}},
			}, nil)
			mockLLM.EXPECT().Inject(project, version, gomock.Any()).Return(errors.New("injection failed"))
			mockSlackBot.EXPECT().PostMessage(channel, threadTS, "❌ Error: injection failed").Return(nil)

			err := testAgent.Inject(channel, threadTS, project, version)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to inject messages"))
		})
	})

	Describe("Start", func() {
		It("should start the agent and handle app mention events", func() {
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			// Create a test event
			testEvent := &slackevents.AppMentionEvent{
				Type:      "app_mention",
				User:      "U123456",
				Text:      "<@BOT123> invalid command",
				Channel:   "C1234567890",
				TimeStamp: "1234567890.123456",
			}

			// Set up mock expectations for the bot user
			botUser := &slack.AuthTestResponse{
				User:   "slack-ai-assistant",
				UserID: "BOT123",
			}
			mockSlackBot.EXPECT().GetBotUser().Return(botUser).AnyTimes()
			mockSlackBot.EXPECT().PostMessage(gomock.Any(), gomock.Any(), "Please use one of the following commands (answer,elaborate,inject)").Return(nil).AnyTimes()

			// Mock the Start method to not block
			mockSlackBot.EXPECT().Start(gomock.Any()).Do(func(ctx context.Context) {
				<-ctx.Done()
			})

			// Start the agent in a goroutine
			go testAgent.Start(ctx)

			// Send an event
			appMentionChannel <- testEvent

			// Wait for context to complete
			<-ctx.Done()
		})
	})
})
