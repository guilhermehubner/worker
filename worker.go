package worker

import (
	"golang.org/x/net/context"

	"github.com/streadway/amqp"
)

type getJobHandle func() (*amqp.Delivery, *JobType)

type worker struct {
	getJob getJobHandle
	cancel chan bool
}

func newWorker(getJob getJobHandle, cancel chan bool) *worker {
	return &worker{
		getJob: getJob,
		cancel: cancel,
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

	for i := retries; i > 0; i-- {
		ctx, cancelFn := context.WithCancel(context.Background())

		err := job.Handle(ctx, message.Body)
		if err == nil {
			break
		}

		cancelFn()
	}
}
