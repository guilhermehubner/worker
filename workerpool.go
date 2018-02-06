package worker

import (
	"sort"

	"github.com/streadway/amqp"
	"golang.org/x/net/context"
)

type WorkerPool struct {
	channel     *amqp.Channel
	workers     []*worker
	jobTypes    jobTypes
	cancel      chan bool
	middlewares []Middleware
}

type Middleware func(context.Context, func(context.Context) error) error

// Start starts the workers and associated processes.
func (wp *WorkerPool) Start() {
	sort.Sort(wp.jobTypes)

	getJob := func() (*amqp.Delivery, *JobType) {
		for _, jobType := range wp.jobTypes {
			msg, ok, _ := wp.channel.Get(jobType.Name, true)
			if !ok {
				continue
			}

			return &msg, &jobType
		}

		return nil, nil
	}

	for _, w := range wp.workers {
		w = newWorker(wp.middlewares, getJob, wp.cancel)
		w.start()
	}

	<-wp.cancel
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
