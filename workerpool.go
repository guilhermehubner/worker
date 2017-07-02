package worker

import (
	"sort"

	"github.com/streadway/amqp"
	"golang.org/x/net/context"
)

type WorkerPool struct {
	ctx      context.Context
	channel  *amqp.Channel
	workers  []*worker
	jobTypes jobTypes
	cancel   chan bool
}

func (wp *WorkerPool) Start() {
	sort.Sort(wp.jobTypes)

	getJob := func() (*amqp.Delivery, jobHandle) {
		for _, jobType := range wp.jobTypes {
			msg, ok, _ := wp.channel.Get(jobType.Name, false)
			if !ok {
				continue
			}

			return &msg, jobType.Handle
		}

		return nil, nil
	}

	for _, w := range wp.workers {
		w = newWorker(wp.ctx, getJob, wp.cancel)
		w.start()
	}

	<-wp.cancel
}

func (wp *WorkerPool) RegisterJob(job JobType) {
	wp.jobTypes = append(wp.jobTypes, job)
}

// NewWorkerPool creates a new worker pool. ctx will be used for middleware and handlers. concurrency specifies how many workers to spin up - each worker can process jobs concurrently.
func NewWorkerPool(ctx context.Context, concurrency uint, channel *amqp.Channel) *WorkerPool {
	if channel == nil {
		panic("worker equeuer: needs a non-nil *amqp.Channel")
	}

	wp := &WorkerPool{
		ctx:     ctx,
		channel: channel,
		workers: make([]*worker, concurrency),
	}

	return wp
}
