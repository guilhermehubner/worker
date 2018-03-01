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
	channel    *amqp.Channel
}

type Status struct {
	JobName  string
	Messages int64
	Error    error `json:",omitempty"`
}

type Message struct {
	message amqp.Delivery
	broker  *AMQPBroker
}

func (m *Message) Body() []byte {
	return m.message.Body
}

func (m *Message) ID() string {
	return m.message.MessageId
}

func (m *Message) Retries() uint8 {
	retry, _ := m.message.Headers["Retry"].(uint8)
	return retry
}

func (m *Message) Requeue() error {
	ch, err := m.broker.connection.Channel()
	if err != nil {
		return err
	}

	return ch.Publish(
		"",
		m.message.RoutingKey,
		false,
		false,
		amqp.Publishing{
			MessageId:    m.message.MessageId,
			Timestamp:    time.Now(),
			Body:         m.message.Body,
			DeliveryMode: 2,
			Headers: amqp.Table{
				"Retry": m.Retries() + 1,
			},
		},
	)
}

func (b *AMQPBroker) RegisterJob(jobName string) error {
	_, err := b.channel.QueueDeclare(
		jobName, // name
		true,    // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)

	return err
}

func (b *AMQPBroker) GetQueueStatus(name string) (Status, error) {
	q, err := b.channel.QueueInspect(name)

	return Status{
		JobName:  q.Name,
		Messages: int64(q.Messages),
		Error:    err,
	}, nil
}

func (b *AMQPBroker) GetMessage(jobName string) *Message {
	msg, ok, err := b.channel.Get(jobName, true)
	if err != nil {
		log.Get().Error(fmt.Sprintf("broker/amqp: fail to get message: %v", err))
		return nil
	}
	if !ok {
		return nil
	}

	return &Message{
		message: msg,
		broker:  b,
	}
}

func (b *AMQPBroker) Enqueue(name, messageID string, message proto.Message) error {
	queue, err := b.channel.QueueDeclare(
		name,  // queue name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
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

	err = b.channel.Publish(
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

func (b *AMQPBroker) EnqueueIn(name, messageID string, message proto.Message,
	secondsFromNow int64) (string, error) {
	_, err := b.channel.QueueDeclare(
		name,  // queue name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		// TODO
		return "", err
	}

	queue, err := b.channel.QueueDeclare(
		fmt.Sprintf(NoConsumerQueue, name), // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		amqp.Table{
			// Exchange where to send messages after TTL expiration.
			"x-dead-letter-exchange": "",
			// Routing key which use when resending expired messages.
			"x-dead-letter-routing-key": name,
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

	err = b.channel.Publish(
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

func (b *AMQPBroker) connect() {
	log.Get().Info("CONNECTING...")

	for {
		connection, err := amqp.Dial(b.url)
		if err != nil {
			log.Get().Error(fmt.Sprintf("broker/amqp: fail to connect: %v", err))
			time.Sleep(b.backoff.Duration())
			continue
		}

		channel, err := connection.Channel()
		if err != nil {
			log.Get().Error(fmt.Sprintf("broker/amqp: fail to get connection channel: %v", err))
			continue
		}

		b.connection = connection
		b.channel = channel
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
