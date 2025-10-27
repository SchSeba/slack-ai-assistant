package agent

import (
	"context"
	"fmt"
	"sync"

	"github.com/slack-go/slack/slackevents"
)

// WorkItem represents a unit of work that can be processed by the worker pool
type WorkItem interface {
	Process(agent *Agent) error
	String() string
}

// AppMentionWorkItem wraps an app mention event for processing
type AppMentionWorkItem struct {
	Event *slackevents.AppMentionEvent
}

func (w AppMentionWorkItem) Process(agent *Agent) error {
	return agent.handleAppMentionEvent(w.Event)
}

func (w AppMentionWorkItem) String() string {
	return fmt.Sprintf("AppMention{User: %s, Channel: %s}", w.Event.User, w.Event.Channel)
}

// WorkerPool manages a pool of workers that process work items
type WorkerPool struct {
	workerCount int
	workQueue   chan WorkItem
	workers     []Worker
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
}

// Worker represents a single worker in the pool
type Worker struct {
	id        int
	workQueue chan WorkItem
	agent     *Agent
	ctx       context.Context
}

// NewWorkerPool creates a new worker pool with the specified number of workers
func NewWorkerPool(workerCount, queueSize int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	return &WorkerPool{
		workerCount: workerCount,
		workQueue:   make(chan WorkItem, queueSize),
		workers:     make([]Worker, workerCount),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Start initializes and starts all workers in the pool
func (wp *WorkerPool) Start(agent *Agent) {
	fmt.Printf("🏭 Starting worker pool with %d workers\n", wp.workerCount)

	for i := 0; i < wp.workerCount; i++ {
		worker := Worker{
			id:        i + 1,
			workQueue: wp.workQueue,
			agent:     agent,
			ctx:       wp.ctx,
		}
		wp.workers[i] = worker

		wp.wg.Add(1)
		go worker.start(&wp.wg)
	}
}

// Submit adds a work item to the queue for processing
func (wp *WorkerPool) Submit(workItem WorkItem) {
	select {
	case wp.workQueue <- workItem:
		// Work item successfully queued
	case <-wp.ctx.Done():
		fmt.Printf("❌ Worker pool is shutting down, cannot submit work: %s\n", workItem.String())
	default:
		fmt.Printf("⚠️ Work queue is full, dropping work item: %s\n", workItem.String())
	}
}

// Stop gracefully shuts down the worker pool
func (wp *WorkerPool) Stop() {
	fmt.Println("🛑 Stopping worker pool...")
	wp.cancel()
	close(wp.workQueue)
	wp.wg.Wait()
	fmt.Println("✅ Worker pool stopped")
}

// start begins the worker's processing loop
func (w *Worker) start(wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Printf("👷 Worker %d started\n", w.id)

	for {
		select {
		case workItem, ok := <-w.workQueue:
			if !ok {
				fmt.Printf("👷 Worker %d shutting down (queue closed)\n", w.id)
				return
			}
			w.processWorkItem(workItem)
		case <-w.ctx.Done():
			fmt.Printf("👷 Worker %d shutting down (context canceled)\n", w.id)
			return
		}
	}
}

// processWorkItem handles a single work item
func (w *Worker) processWorkItem(workItem WorkItem) {
	fmt.Printf("👷 Worker %d processing: %s\n", w.id, workItem.String())

	if err := workItem.Process(w.agent); err != nil {
		fmt.Printf("❌ Worker %d failed to process %s: %v\n", w.id, workItem.String(), err)
	} else {
		fmt.Printf("✅ Worker %d completed: %s\n", w.id, workItem.String())
	}
}
