package worker

import (
	"fmt"
	"time"

	"golang.org/x/net/context"

	"github.com/golang/protobuf/proto"
	"github.com/guilhermehubner/worker/broker"
	"github.com/guilhermehubner/worker/log"
	"github.com/jpillora/backoff"
)

const (
	minBackoffTime = 100 * time.Millisecond
	maxBackoffTime = 10 * time.Second
)

type getJobHandle func() (*broker.Message, *JobType, error)

type worker struct {
	middlewares       []Middleware
	getJob            getJobHandle
	cancel            chan struct{}
	ended             chan struct{}
	backoff           *backoff.Backoff
	getMessageBackoff *backoff.Backoff
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
		getMessageBackoff: &backoff.Backoff{
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
	message, job, err := w.getJob()
	if err != nil {
		log.Get().Error(fmt.Sprintf("worker: fail to get job: %v", err))
		time.Sleep(w.getMessageBackoff.Duration())
		return
	}

	w.getMessageBackoff.Reset()

	if message == nil {
		return
	}

	gen := func(msg proto.Message) error {
		return proto.Unmarshal(message.Body(), msg)
	}

	wrappedHandle := func(ctx context.Context) error {
		return job.Handle(ctx, gen)
	}

	for i := len(w.middlewares) - 1; i >= 0; i-- {
		index := i
		oldWrapped := wrappedHandle

		wrappedHandle = func(ctx context.Context) error {
			return w.middlewares[index](injectJobInfo(ctx, *job, message),
				oldWrapped)
		}
	}

	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()

	if err := wrappedHandle(ctx); err != nil {
		if job != nil && message.Retries() < job.Retry {
			for err := message.Requeue(); err != nil; {
				time.Sleep(w.backoff.Duration())
			}

			w.backoff.Reset()
		}
	}
}
