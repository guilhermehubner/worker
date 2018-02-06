package worker

import (
	"fmt"
	"os"
	"os/signal"
	"sort"
	"syscall"

	"github.com/streadway/amqp"
	"golang.org/x/net/context"
)

var gracefulStop = make(chan os.Signal)

type WorkerPool struct {
	channel     *amqp.Channel
	workers     []*worker
	jobTypes    jobTypes
	middlewares []Middleware
}

type Middleware func(context.Context, func(context.Context) error) error

// Start starts the workers and associated processes.
func (wp *WorkerPool) Start() {
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)
	go func() {
		<-gracefulStop
		for _, w := range wp.workers {
			if w != nil {
				w.cancel <- struct{}{}
			}
		}
	}()

	sort.Sort(wp.jobTypes)

	getJob := func() (*amqp.Delivery, *JobType) {
		for _, jobType := range wp.jobTypes {
			msg, ok, err := wp.channel.Get(jobType.Name, true)
			if !ok || err != nil {
				continue
			}

			return &msg, &jobType
		}

		return nil, nil
	}

	workersEnded := make([]chan struct{}, 0, len(wp.workers))

	for i, _ := range wp.workers {
		wp.workers[i] = newWorker(wp.middlewares, getJob)
		workersEnded = append(workersEnded, wp.workers[i].start())
	}

	for i, w := range wp.workers {
		<-w.ended
		fmt.Printf("Finish worker %d\n", i+1)
	}
}

// RegisterJob adds a job with handler for 'name' queue and allows you to specify options such as a job's priority and it's retry count.
func (wp *WorkerPool) RegisterJob(job JobType) {
	wp.jobTypes = append(wp.jobTypes, job)
}

// NewWorkerPool creates a new worker pool. ctx will be used for middleware and handlers. concurrency specifies how many workers to spin up - each worker can process jobs concurrently.
func NewWorkerPool(concurrency uint, channel *amqp.Channel,
	middlewares ...Middleware) *WorkerPool {
	if channel == nil {
		panic("worker equeuer: needs a non-nil *amqp.Channel")
	}

	wp := &WorkerPool{
		middlewares: middlewares,
		channel:     channel,
		workers:     make([]*worker, concurrency),
	}

	return wp
}
