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
	stop        bool
}

type Status struct {
	JobName  string
	Messages int64
}

type Middleware func(context.Context, func(context.Context) error) error

func (wp *WorkerPool) GetPoolStatus() ([]Status, error) {
	stats := make([]Status, 0, len(wp.jobTypes))

	for _, jobType := range wp.jobTypes {
		q, err := wp.channel.QueueInspect(jobType.Name)
		if err != nil {
			return nil, err
		}

		stats = append(stats, Status{
			JobName:  jobType.Name,
			Messages: int64(q.Messages),
		})
	}

	return stats, nil
}

// Start starts the workers and associated processes.
func (wp *WorkerPool) Start() {
	signal.Notify(gracefulStop, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-gracefulStop
		wp.stop = true
		for _, w := range wp.workers {
			if w != nil {
				w.cancel <- struct{}{}
			}
		}
	}()

	sort.Sort(wp.jobTypes)

	getJob := func() (*amqp.Delivery, *JobType) {
		for _, jobType := range wp.jobTypes {
			if wp.stop {
				return nil, nil
			}

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
	_, err := wp.channel.QueueDeclare(
		job.Name, // name
		true,     // durable
		false,    // delete when unused
		false,    // exclusive
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		// TODO
	}

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
