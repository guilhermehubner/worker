package worker

import (
	"time"

	"github.com/streadway/amqp"
)

type Enqueuer struct {
	channel *amqp.Channel
}

// NewEnqueuer creates a new enqueuer with the specified RabbitMQ channel.
func NewEnqueuer(channel *amqp.Channel) *Enqueuer {
	if channel == nil {
		panic("worker equeuer: needs a non-nil *amqp.Channel")
	}

	return &Enqueuer{
		channel: channel,
	}
}

// Enqueue will enqueue the specified job name and arguments. The args param can be nil if no args ar needed.
func (e *Enqueuer) Enqueue(jobName string, params ...interface{}) error {
	queue, err := e.channel.QueueDeclare(
		jobName, // name
		true,    // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	if err != nil {
		// TODO
		return err
	}

	body, err := encode(params)
	if err != nil {
		return err
	}

	err = e.channel.Publish(
		"",
		queue.Name,
		false,
		false,
		amqp.Publishing{
			MessageId: makeIdentifier(),
			Timestamp: time.Now(),
			Body:      body,
		},
	)

	if err != nil {
		// TODO
		return err
	}

	return nil
}
