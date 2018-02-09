package worker

import (
	"time"

	"golang.org/x/net/context"

	"github.com/golang/protobuf/proto"
	"github.com/jpillora/backoff"
	"github.com/streadway/amqp"
)

const (
	minBackoffTime = 100 * time.Millisecond
	maxBackoffTime = 10 * time.Second
)

type getJobHandle func() (*amqp.Delivery, *JobType)

type worker struct {
	middlewares []Middleware
	getJob      getJobHandle
	cancel      chan struct{}
	ended       chan struct{}
	backoff     *backoff.Backoff
}

func newWorker(middlewares []Middleware, getJob getJobHandle) *worker {
	return &worker{
		middlewares: middlewares,
		getJob:      getJob,
		cancel:      make(chan struct{}),
		ended:       make(chan struct{}),
		backoff: &backoff.Backoff{
			Min: minBackoffTime,
			Max: maxBackoffTime,
		},
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
	rawMessage, job := w.getJob()
	if rawMessage == nil || job == nil {
		return
	}

	gen := func(message proto.Message) error {
		return proto.Unmarshal(rawMessage.Body, message)
	}

	retries := 1
	if job.Retry > 0 {
		retries = int(job.Retry)
	}

	wrappedHandle := func(ctx context.Context) error {
		return job.Handle(ctx, gen)
	}

	for i := len(w.middlewares) - 1; i >= 0; i-- {
		index := i
		oldWrapped := wrappedHandle

		wrappedHandle = func(ctx context.Context) error {
			return w.middlewares[index](injectJobInfo(ctx, *job, rawMessage.MessageId),
				oldWrapped)
		}
	}

	for i := retries; i > 0; i-- {
		ctx, cancelFn := context.WithCancel(context.Background())

		err := wrappedHandle(ctx)
		if err == nil {
			w.backoff.Reset()
			break
		}

		cancelFn()
		time.Sleep(w.backoff.Duration())
	}
}
