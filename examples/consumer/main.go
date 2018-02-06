package main

import (
	"fmt"

	"golang.org/x/net/context"

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

	wp := worker.NewWorkerPool(5, ch)

	wp.RegisterJob(worker.JobType{
		Name: "fila1",
		Handle: func(_ context.Context, params ...interface{}) error {
			for position, value := range params {
				fmt.Printf("Job: fila 1, param %d, msg: %s\n", position, value)
			}
			return nil
		},
		Priority: 10,
	})

	wp.RegisterJob(worker.JobType{
		Name: "fila2",
		Handle: func(_ context.Context, params ...interface{}) error {
			for position, value := range params {
				fmt.Printf("Job: fila 2, param %d, msg: %s\n", position, value)
			}
			return nil
		},
		Priority: 15,
	})

	wp.Start()
}
