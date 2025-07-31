package agent_test

import (
	"errors"
	"fmt"
	"sync"
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

// TestWorkItem implements the WorkItem interface for testing
type TestWorkItem struct {
	ID          string
	ShouldError bool
	ProcessFunc func(*agent.Agent) error
}

func (t TestWorkItem) Process(agentProcess *agent.Agent) error {
	if t.ProcessFunc != nil {
		return t.ProcessFunc(agentProcess)
	}
	if t.ShouldError {
		return errors.New("test error")
	}
	return nil
}

func (t TestWorkItem) String() string {
	return "TestWorkItem{ID: " + t.ID + "}"
}

var _ = Describe("WorkerPool", func() {
	var (
		ctrl         *gomock.Controller
		mockDB       *databaseMock.MockInterface
		mockSlackBot *slackbotMock.MockInterface
		mockLLM      *llmMock.MockInterface
		testAgent    *agent.Agent
		workerPool   *agent.WorkerPool
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockDB = databaseMock.NewMockInterface(ctrl)
		mockSlackBot = slackbotMock.NewMockInterface(ctrl)
		mockLLM = llmMock.NewMockInterface(ctrl)

		appMentionChannel := make(chan *slackevents.AppMentionEvent, 10)
		slashCommandChannel := make(chan *slack.SlashCommand, 10)

		testAgent = agent.NewAgent(mockDB, mockSlackBot, mockLLM, appMentionChannel, slashCommandChannel, 2)
		workerPool = agent.NewWorkerPool(2, 10)
	})

	AfterEach(func() {
		if workerPool != nil {
			workerPool.Stop()
		}
		ctrl.Finish()
	})

	Describe("NewWorkerPool", func() {
		It("should create a worker pool with specified workers and queue size", func() {
			wp := agent.NewWorkerPool(3, 20)
			Expect(wp).NotTo(BeNil())
			wp.Stop()
		})
	})

	Describe("Start and Stop", func() {
		It("should start workers and stop gracefully", func() {
			workerPool.Start(testAgent)

			// Give workers time to start
			time.Sleep(10 * time.Millisecond)

			// Stop should complete without hanging
			done := make(chan bool)
			go func() {
				workerPool.Stop()
				done <- true
			}()

			select {
			case <-done:
				// Success
			case <-time.After(1 * time.Second):
				Fail("Worker pool stop timed out")
			}

			workerPool = nil // Prevent double stop in AfterEach
		})
	})

	Describe("Submit", func() {
		BeforeEach(func() {
			workerPool.Start(testAgent)
		})

		It("should process work items successfully", func() {
			processed := make(chan bool, 1)
			workItem := TestWorkItem{
				ID: "test1",
				ProcessFunc: func(agent *agent.Agent) error {
					processed <- true
					return nil
				},
			}

			workerPool.Submit(workItem)

			select {
			case <-processed:
				// Success
			case <-time.After(500 * time.Millisecond):
				Fail("Work item was not processed")
			}
		})

		It("should handle work item errors gracefully", func() {
			processed := make(chan bool, 1)
			workItem := TestWorkItem{
				ID: "test2",
				ProcessFunc: func(agent *agent.Agent) error {
					processed <- true
					return errors.New("intentional test error")
				},
			}

			workerPool.Submit(workItem)

			select {
			case <-processed:
				// Success - error was handled
			case <-time.After(500 * time.Millisecond):
				Fail("Work item was not processed")
			}
		})

		It("should process multiple work items concurrently", func() {
			const numItems = 5
			processed := make(chan string, numItems)

			var wg sync.WaitGroup
			wg.Add(numItems)

			for i := 0; i < numItems; i++ {
				workItem := TestWorkItem{
					ID: fmt.Sprintf("test%d", i),
					ProcessFunc: func(agent *agent.Agent) error {
						defer wg.Done()
						processed <- "done"
						return nil
					},
				}
				workerPool.Submit(workItem)
			}

			// Wait for all items to be processed or timeout
			done := make(chan bool)
			go func() {
				wg.Wait()
				done <- true
			}()

			select {
			case <-done:
				Expect(len(processed)).To(Equal(numItems))
			case <-time.After(1 * time.Second):
				Fail("Not all work items were processed")
			}
		})
	})

	Describe("AppMentionWorkItem", func() {
		var (
			testEvent *slackevents.AppMentionEvent
			workItem  agent.AppMentionWorkItem
		)

		BeforeEach(func() {
			testEvent = &slackevents.AppMentionEvent{
				Type:      "app_mention",
				User:      "U123456",
				Text:      "<@BOT123> invalid",
				Channel:   "C1234567890",
				TimeStamp: "1234567890.123456",
			}
			workItem = agent.AppMentionWorkItem{Event: testEvent}
		})

		It("should implement WorkItem interface correctly", func() {
			// Test String method
			str := workItem.String()
			Expect(str).To(ContainSubstring("AppMention"))
			Expect(str).To(ContainSubstring("U123456"))
			Expect(str).To(ContainSubstring("C1234567890"))
		})

		It("should process app mention events with error handling", func() {
			// Mock the bot user for the event processing
			botUser := &slack.AuthTestResponse{
				User:   "slack-ai-assistant",
				UserID: "BOT123",
			}

			// Set up mock expectations
			mockSlackBot.EXPECT().GetBotUser().Return(botUser).AnyTimes()
			mockSlackBot.EXPECT().GetConversationReplies(gomock.Any()).Return(nil, nil).AnyTimes() // Return nil to simulate API unavailable
			mockSlackBot.EXPECT().PostMessage(gomock.Any(), gomock.Any(), "Please use one of the following commands (answer,elaborate,inject)").Return(nil)

			err := workItem.Process(testAgent)
			Expect(err).NotTo(HaveOccurred()) // The error is handled internally and a help message is posted
		})
	})

	Describe("Queue Management", func() {
		It("should handle queue overflow gracefully", func() {
			// Create a small queue
			smallPool := agent.NewWorkerPool(1, 2)
			defer smallPool.Stop()

			smallPool.Start(testAgent)

			// Create a blocking work item
			blockingItem := TestWorkItem{
				ID: "blocking",
				ProcessFunc: func(agent *agent.Agent) error {
					time.Sleep(100 * time.Millisecond)
					return nil
				},
			}

			// Fill the queue beyond capacity
			smallPool.Submit(blockingItem)
			smallPool.Submit(blockingItem)

			// This should not block - the item should be dropped
			done := make(chan bool)
			go func() {
				smallPool.Submit(TestWorkItem{ID: "overflow"})
				done <- true
			}()

			select {
			case <-done:
				// Success - submit didn't block
			case <-time.After(200 * time.Millisecond):
				Fail("Submit blocked when queue was full")
			}
		})
	})
})
