package main

import (
	"fmt"
	"time"

	"github.com/guilhermehubner/worker"
	"github.com/guilhermehubner/worker/examples/payload"
	"golang.org/x/net/context"
)

func main() {
	wp := worker.NewWorkerPool("amqp://guest:guest@localhost:5672/", 4, emoji, log)

	wp.RegisterJob(worker.JobType{
		Name:     "calculator",
		Handle:   calculate,
		Priority: 10,
	})

	wp.RegisterJob(worker.JobType{
		Name:     "mailing",
		Handle:   sendEmail,
		Priority: 15,
	})

	wp.Start()
}

func sendEmail(_ context.Context, gen worker.GenFunc) error {
	msg := payload.Email{}
	err := gen(&msg)
	if err != nil {
		fmt.Println("Fail to decode message on mailing service")
		return err
	}

	fmt.Printf("Email sent to \"%s\": \"%s\"\n", msg.To, msg.Subject)

	time.Sleep(3 * time.Second)

	return nil
}

func calculate(_ context.Context, gen worker.GenFunc) error {
	msg := payload.Expression{}
	err := gen(&msg)
	if err != nil {
		fmt.Println("Fail to decode message on calculator")
		return err
	}

	var result int64

	switch msg.Operation {
	case payload.Expression_ADD:
		result = msg.Operand1 + msg.Operand2
	case payload.Expression_SUB:
		result = msg.Operand1 - msg.Operand2
	case payload.Expression_MUL:
		result = msg.Operand1 * msg.Operand2
	case payload.Expression_DIV:
		result = msg.Operand1 / msg.Operand2
	}

	fmt.Printf("%d %s %d = %d\n", msg.Operand1, msg.Operation, msg.Operand2, result)

	time.Sleep(3 * time.Second)

	return nil
}

func emoji(ctx context.Context, next worker.NextMiddleware) error {
	switch worker.JobInfoFromContext(ctx).Name {
	case "mailing":
		fmt.Print("✉️  ")
	case "calculator":
		fmt.Print("📟  ")
	}

	return next(ctx)
}

func log(ctx context.Context, next worker.NextMiddleware) error {
	fmt.Printf("%s: ", worker.JobInfoFromContext(ctx).Name)

	return next(ctx)
}
