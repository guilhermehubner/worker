package worker

import (
	"github.com/streadway/amqp"
	"golang.org/x/net/context"
)

type getJobHandle func() (*amqp.Delivery, jobHandle)

type worker struct {
	ctx    context.Context
	getJob getJobHandle
	cancel chan bool
}

func newWorker(ctx context.Context, getJob getJobHandle,
	cancel chan bool) *worker {
	return &worker{
		ctx:    ctx,
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
	message, handle := w.getJob()
	if message == nil || handle == nil {
		return
	}

	err := handle(w.ctx, message.Body)
	if err == nil {
		message.Ack(false)
	}
}
