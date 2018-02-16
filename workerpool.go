package worker

import (
	"fmt"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"

	"github.com/guilhermehubner/worker/broker"
	"golang.org/x/net/context"
)

var gracefulStop = make(chan os.Signal)

type WorkerPool struct {
	broker      *broker.AMQPBroker
	workers     []*worker
	jobTypes    jobTypes
	middlewares []Middleware
	stop        bool
}

type Middleware func(context.Context, func(context.Context) error) error

func (wp *WorkerPool) GetPoolStatus() ([]broker.Status, error) {
	stats := make([]broker.Status, 0, len(wp.jobTypes))

	for _, jobType := range wp.jobTypes {
		s, err := wp.broker.GetJobStatus(jobType.Name)
		if err != nil {
			return nil, err
		}

		stats = append(stats, s)
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

	getJob := func() ([]byte, string, *JobType) {
		for _, jobType := range wp.jobTypes {
			if wp.stop {
				return nil, "", nil
			}

			msg, messageID := wp.broker.GetMessage(jobType.Name)
			if msg == nil {
				continue
			}

			return msg, messageID, &jobType
		}

		return nil, "", nil
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
	err := wp.broker.RegisterJob(job.Name)
	if err != nil {
		// TODO
	}

	wp.jobTypes = append(wp.jobTypes, job)
}

/*
NewWorkerPool creates a new worker pool.

URL is a string connection in the AMQP URI format.

Concurrency specifies how many workers to spin up - each worker can process jobs concurrently.
*/
func NewWorkerPool(url string, concurrency uint, middlewares ...Middleware) *WorkerPool {
	if strings.TrimSpace(url) == "" {
		panic("worker workerpool: needs a non-empty url")
	}

	wp := &WorkerPool{
		broker:      broker.NewBroker(url),
		middlewares: middlewares,
		workers:     make([]*worker, concurrency),
	}

	return wp
}
