package worker

import (
	"golang.org/x/net/context"

	"github.com/streadway/amqp"
)

type getJobHandle func() (*amqp.Delivery, *JobType)

type worker struct {
	middlewares []Middleware
	getJob      getJobHandle
	cancel      chan struct{}
	ended       chan struct{}
}

func newWorker(middlewares []Middleware, getJob getJobHandle) *worker {
	return &worker{
		middlewares: middlewares,
		getJob:      getJob,
		cancel:      make(chan struct{}),
		ended:       make(chan struct{}),
	}
}

func (w *worker) start() chan struct{} {
	go func() {
		for {
			select {
			case <-w.cancel:
				w.ended <- struct{}{}
				return
			default:
				w.executeJob()
			}
		}
	}()

	return w.ended
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
