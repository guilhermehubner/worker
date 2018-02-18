package main

import (
	"fmt"
	"log"

	"golang.org/x/net/context"

	"github.com/guilhermehubner/worker"
	"github.com/guilhermehubner/worker/examples/payload"
)

func main() {
	wp, err := worker.NewWorkerPool("amqp://guest:guest@localhost:5672/", 5,
		func(ctx context.Context, next func(context.Context) error) error {
			fmt.Print("Enter on Middleware 1 > ")
			return next(ctx)
		},
		func(ctx context.Context, next func(context.Context) error) error {
			fmt.Print("Enter on Middleware 2 > ")
			return next(ctx)
		})
	if err != nil {
		log.Fatal(err)
	}

	err = wp.RegisterJob(worker.JobType{
		Name: "queue1",
		Handle: func(ctx context.Context, gen worker.GenFunc) error {
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
	if err != nil {
		log.Fatal(err)
	}

	err = wp.RegisterJob(worker.JobType{
		Name: "queue2",
		Handle: func(ctx context.Context, gen worker.GenFunc) error {
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
	if err != nil {
		log.Fatal(err)
	}

	wp.Start()
}
