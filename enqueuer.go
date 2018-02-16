package worker

import (
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/guilhermehubner/worker/broker"
)

type Enqueuer struct {
	broker *broker.AMQPBroker
}

/*
NewEnqueuer creates a new enqueuer.

URL is a string connection in the AMQP URI format.
*/
func NewEnqueuer(url string) *Enqueuer {
	if strings.TrimSpace(url) == "" {
		panic("worker equeuer: needs a non-empty url")
	}

	return &Enqueuer{
		broker: broker.NewBroker(url),
	}
}

// Enqueue will enqueue the specified message for job queue.
func (e *Enqueuer) Enqueue(jobName string, message proto.Message) error {
	return e.broker.Enqueue(jobName, makeIdentifier(), message)
}

// EnqueueIn enqueues a message in the scheduled job queue for execution secondsFromNow seconds.
func (e *Enqueuer) EnqueueIn(jobName string, message proto.Message,
	secondsFromNow int64) (string, error) {
	return e.broker.EnqueueIn(jobName, makeIdentifier(), message, secondsFromNow)
}
