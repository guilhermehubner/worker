package main

import (
	"fmt"
	"time"

	"github.com/guilhermehubner/worker"
	"github.com/guilhermehubner/worker/examples/payload"
)

func main() {
	enqueuer := worker.NewEnqueuer("amqp://guest:guest@localhost:5672/")

	for {
		time.Sleep(500 * time.Millisecond)

		err := enqueuer.Enqueue("queue1", &payload.Payload{
			Text:   "Hello queue 1",
			Number: time.Now().Unix(),
		})
		if err != nil {
			fmt.Printf("Failed to enqueue 1: %s", err)
			return
		}

		err = enqueuer.Enqueue("queue2", &payload.Payload{
			Text:   "Hello queue 2",
			Number: time.Now().Unix(),
		})
		if err != nil {
			fmt.Printf("Failed to enqueue 2: %s", err)
			return
		}
	}
}
