# Worker [![Build Status](https://travis-ci.org/guilhermehubner/worker.svg?branch=master)](https://travis-ci.org/guilhermehubner/worker) [![Go Report Card](https://goreportcard.com/badge/github.com/guilhermehubner/worker)](https://goreportcard.com/report/github.com/guilhermehubner/worker) [![GoDoc](https://godoc.org/github.com/guilhermehubner/worker?status.svg)](https://godoc.org/github.com/guilhermehubner/worker)

Worker is a distributed system for enqueuing and processing jobs asynchronously in Go.

* Works on top of RabbitMQ message broker (AMQP)
* Uses protobuf to encode jobs
* Accepts middlewares

## Getting Started

Download the worker package:

```
go get github.com/guilhermehubner/worker
```

Start a RabbitMQ container:
```
docker run -p 5672:5672 -p 15672:15672 -d --hostname my-rabbit --name some-rabbit rabbitmq:3-management
```

Run the example by launching the enqueuer followed by consumer:
```
go run examples/enqueuer/main.go && go run examples/consumer/main.go
```

The example presents a worker pool with two consumer types: a mailing service and a calculator. The enqueuer sends 10 jobs of each that are then consumed according to priorities and concurrency length. Two middlewares are executed synchronously and in order before the actual job computation.    

You can also access the management plugin offered by RabbitMQ on `http://localhost:15672/` to see the queues created, using both username and password as `guest` .

## Enqueuer: send

The enqueuer is responsible for sending job/tasks to the queue. Each job is represented by a message and the queue it is intended to be delivered to. When declaring a enqueuer it is necessary to inform the address the messages will be sent to and that's it! Being the one responsible for sending the jobs, it is also called **producer**.

```golang
enqueuer := worker.NewEnqueuer("amqp://guest:guest@localhost:5672/")
enqueuer.Enqueue("mailing", &payload.Email{
        From:    "John",
        To:      "Mary",
        Subject: fmt.Sprintf("Photos from last night %d/10", i),
        Body:    "Attachment",
    })
```

The example above illustrates the definition of an enqueuer as well as the process of sending a message to the queue named "mailing".

## Worker Pool: receive

A worker pool is a group of workers that will receive the messages and process them. They also can be referred as **consumers**. When creating a worker pool you must inform an URL for connection and its length - how many workers will be asynchronously handling jobs - via the `concurrency` parameter. You might also define middlewares and all types of jobs your consumers will be able to handle. By doing so you are expected to inform its name, priority and the function responsible for handling the message received and processing the job itself.

```golang
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
```

In the example above the pool comprehends of 4 concurrent workers capable of processing 2 kinds of jobs (a calculator and a mailing system) and also has 2 middlewares attached to it.

## Payload

Keep in mind that both producer and consumer must know the structure of the messages so both can encode and decode them.