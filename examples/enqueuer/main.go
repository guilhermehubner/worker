package main

import (
	"fmt"

	"github.com/guilhermehubner/worker"
	"github.com/guilhermehubner/worker/examples/payload"
)

func main() {
	enqueuer := worker.NewEnqueuer("amqp://guest:guest@localhost:5672/")

	for i := 1; i <= 10; i++ {
		err := enqueuer.Enqueue("mailing", &payload.Email{
			From:    "John",
			To:      "Mary",
			Subject: fmt.Sprintf("Photos from last night %d/10", i),
			Body:    "Attachment",
		})
		if err != nil {
			fmt.Printf("Failed enqueueing a mailing job: %s", err)
			return
		}
		fmt.Printf("Sent mailing job %d\n", i)

		err = enqueuer.Enqueue("calculator", &payload.Expression{
			Operand1:  10,
			Operand2:  int64(i),
			Operation: payload.Expression_ADD,
		})
		if err != nil {
			fmt.Printf("Failed enqueueing calculation job: %s", err)
			return
		}
		fmt.Printf("Sent calculation job %d\n", i)
	}
}
