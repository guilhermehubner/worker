package main

import (
	"fmt"
	"time"

	"github.com/guilhermehubner/worker"
	"github.com/streadway/amqp"
)

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		fmt.Printf("Failed to connect to RabbitMQ: %s", err)
		return
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		fmt.Printf("Failed to open a channel: %s", err)
		return
	}
	defer ch.Close()

	for {
		time.Sleep(500 * time.Millisecond)

		err = worker.NewEnqueuer(ch).Enqueue("fila1", []byte(fmt.Sprintf("%s: Olá fila 1", time.Now().String())))
		if err != nil {
			fmt.Printf("Failed to enqueue 1: %s", err)
			return
		}
		err = worker.NewEnqueuer(ch).Enqueue("fila2", []byte(fmt.Sprintf("%s: Olá fila 2", time.Now().String())))
		if err != nil {
			fmt.Printf("Failed to enqueue 2: %s", err)
			return
		}
	}
}
