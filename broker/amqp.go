package broker

import (
	"fmt"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/guilhermehubner/worker/log"
	"github.com/jpillora/backoff"
	"github.com/streadway/amqp"
)

const (
	minBackoffTime = 100 * time.Millisecond
	maxBackoffTime = 5 * time.Second

	NoConsumerQueue = "_%s_worker_delayed_5f345b3c-cab6-498a-9bc5-4de9537f8a5b"
)

type AMQPBroker struct {
	url        string
	connection *amqp.Connection
	closed     chan *amqp.Error
	backoff    *backoff.Backoff
	connected  bool
}

type Status struct {
	JobName  string
	Messages int64
	Error    error `json:",omitempty"`
}

func (b *AMQPBroker) RegisterJob(jobName string) error {
	channel, err := b.getChannel()
	if err != nil {
		return err
	}

	_, err = channel.QueueDeclare(
		jobName, // name
		true,    // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)

	return err
}

func (b *AMQPBroker) GetJobStatus(jobName string) (Status, error) {
	channel, err := b.getChannel()
	if err != nil {
		return Status{}, err
	}

	q, err := channel.QueueInspect(jobName)

	return Status{
		JobName:  jobName,
		Messages: int64(q.Messages),
		Error:    err,
	}, nil
}

func (b *AMQPBroker) GetMessage(jobName string) ([]byte, string) {
	channel, err := b.getChannel()
	if err != nil {
		return nil, ""
	}

	msg, ok, err := channel.Get(jobName, true)
	if err != nil {
		log.Get().Error(fmt.Sprintf("broker/amqp: fail to get message: %v", err))
		return nil, ""
	}
	if !ok {
		return nil, ""
	}

	return msg.Body, msg.MessageId
}

func (b *AMQPBroker) Enqueue(jobName, messageID string, message proto.Message) error {
	channel, err := b.getChannel()
	if err != nil {
		// TODO
		return err
	}

	queue, err := channel.QueueDeclare(
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

	err = channel.Publish(
		"",
		queue.Name,
		false,
		false,
		amqp.Publishing{
			MessageId:    messageID,
			Timestamp:    time.Now(),
			Body:         body,
			DeliveryMode: 2,
		},
	)

	if err != nil {
		// TODO
	}

	return err
}

func (b *AMQPBroker) EnqueueIn(jobName, messageID string, message proto.Message,
	secondsFromNow int64) (string, error) {
	channel, err := b.getChannel()
	if err != nil {
		// TODO
		return "", err
	}

	_, err = channel.QueueDeclare(
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

	queue, err := channel.QueueDeclare(
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

	err = channel.Publish(
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

func (b *AMQPBroker) getChannel() (*amqp.Channel, error) {
	var channel *amqp.Channel
	var err error

	for i := 0; i < 20; i++ {
		channel, err = b.connection.Channel()
		if err != nil {
			log.Get().Error(fmt.Sprintf("broker/amqp: fail to get channel: %v", err))
			time.Sleep(5 * time.Millisecond)
			continue
		}

		return channel, nil
	}

	return channel, err
}

func (b *AMQPBroker) connect() {
	log.Get().Info("CONNECTING...")

	for {
		connection, err := amqp.Dial(b.url)
		if err != nil {
			log.Get().Error(fmt.Sprintf("broker/amqp: fail to connect: %v", err))
			time.Sleep(b.backoff.Duration())
			continue
		}

		b.connection = connection
		b.closed = b.connection.NotifyClose(make(chan *amqp.Error))
		b.backoff.Reset()
		b.connected = true
		log.Get().Info("\x1b[1;32mCONNECTED\x1b[0m")
		break
	}
}

func (b *AMQPBroker) reconnect() {
	for e := range b.closed {
		log.Get().Info("\x1b[1;31mCONNECTION CLOSED\x1b[0m")
		log.Get().Error(e)

		b.connected = false
		b.connect()
	}
}

func NewBroker(url string) *AMQPBroker {
	broker := &AMQPBroker{
		url: url,
		backoff: &backoff.Backoff{
			Min: minBackoffTime,
			Max: maxBackoffTime,
		},
	}
	broker.connect()
	go broker.reconnect()

	return broker
}
