package worker

import (
	"fmt"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/streadway/amqp"
)

const NoConsumerQueue = "_%s_worker_delayed_5f345b3c-cab6-498a-9bc5-4de9537f8a5b"

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
func (e *Enqueuer) Enqueue(jobName string, message proto.Message) error {
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

	body, err := proto.Marshal(message)
	if err != nil {
		// TODO
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

// EnqueueIn enqueues a job in the scheduled job queue for execution secondsFromNow seconds.
func (e *Enqueuer) EnqueueIn(jobName string, message proto.Message, secondsFromNow int64) (string, error) {
	_, err := e.channel.QueueDeclare(
		jobName, // name
		true,    // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	if err != nil {
		// TODO
		return "", err
	}

	queue, err := e.channel.QueueDeclare(
		fmt.Sprintf(NoConsumerQueue, jobName), // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		amqp.Table{
			// Exchange where to send messages after TTL expiration.
			"x-dead-letter-exchange": "",
			// Routing key which use when resending expired messages.
			"x-dead-letter-routing-key": jobName,
		},
	)
	if err != nil {
		// TODO
		return "", err
	}

	body, err := proto.Marshal(message)
	if err != nil {
		// TODO
		return "", err
	}

	messageID := makeIdentifier()

	err = e.channel.Publish(
		"",
		queue.Name,
		false,
		false,
		amqp.Publishing{
			MessageId:  messageID,
			Timestamp:  time.Now(),
			Body:       body,
			Expiration: fmt.Sprintf("%d", secondsFromNow*1000),
		},
	)

	if err != nil {
		// TODO
		return "", err
	}

	return messageID, nil
}
