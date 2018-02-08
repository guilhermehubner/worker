package main

import (
	"fmt"

	"context"

	"github.com/golang/protobuf/proto"
	"github.com/guilhermehubner/worker"
	"github.com/guilhermehubner/worker/examples/payload"
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

	wp := worker.NewWorkerPool(5, ch,
		func(ctx context.Context, next func(context.Context) error) error {
			fmt.Print("Enter on Middleware 1 > ")
			return next(ctx)
		},
		func(ctx context.Context, next func(context.Context) error) error {
			fmt.Print("Enter on Middleware 2 > ")
			return next(ctx)
		})

	wp.RegisterJob(worker.JobType{
		Name: "queue1",
		Handle: func(ctx context.Context, gen func(proto.Message) error) error {
			msg := payload.Payload{}
			err := gen(&msg)
			if err != nil {
				fmt.Println("Fail to decode message on queue 1")
				return nil
			}

			fmt.Printf("Job: queue 1, msg: %s gen-> %d - job: %s\n", msg.Text, msg.Number,
				worker.JobInfoFromContext(ctx).Name)
			return nil
		},
		Priority: 10,
	})

	wp.RegisterJob(worker.JobType{
		Name: "queue2",
		Handle: func(ctx context.Context, gen func(proto.Message) error) error {
			msg := payload.Payload{}
			err := gen(&msg)
			if err != nil {
				fmt.Println("Fail to decode message on queue 2")
				return nil
			}

			fmt.Printf("Job: queue 2, msg: %s gen-> %d - job: %s\n", msg.Text, msg.Number,
				worker.JobInfoFromContext(ctx).Name)
			return nil
		},
		Priority: 15,
	})

	wp.Start()
}
