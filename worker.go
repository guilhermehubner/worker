package worker

import (
	"golang.org/x/net/context"

	"github.com/streadway/amqp"
)

type getJobHandle func() (*amqp.Delivery, *JobType)

type worker struct {
	middlewares []Middleware
	getJob      getJobHandle
	cancel      chan bool
}

func newWorker(middlewares []Middleware, getJob getJobHandle, cancel chan bool) *worker {
	return &worker{
		middlewares: middlewares,
		getJob:      getJob,
		cancel:      cancel,
	}
}

func (w *worker) start() {
	go func() {
		for {
			select {
			case <-w.cancel:
				return
			default:
				w.executeJob()
			}
		}
	}()
}

func (w *worker) executeJob() {
	message, job := w.getJob()
	if message == nil || job == nil {
		return
	}

	retries := 1
	if job.Retry > 0 {
		retries = int(job.Retry)
	}

	wrappedHandle := func(ctx context.Context) error {
		return job.Handle(ctx, message.Body)
	}

	for i := len(w.middlewares) - 1; i >= 0; i-- {
		index := i
		oldWrapped := wrappedHandle

		wrappedHandle = func(ctx context.Context) error {
			return w.middlewares[index](ctx, oldWrapped)
		}
	}

	for i := retries; i > 0; i-- {
		ctx, cancelFn := context.WithCancel(context.Background())

		err := wrappedHandle(ctx)
		if err == nil {
			break
		}

		cancelFn()
	}
}
