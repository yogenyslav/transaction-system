package main

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Status int8

const (
	_ Status = iota
	Success
	Error
	Create
)

func processTransaction(id uint) (Status, error) {
	time.Sleep(10 * time.Second)
	if id%4 == 0 {
		return Error, nil
	}

	return Success, nil
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var (
		conn *amqp.Connection
		err  error
	)
	for {
		time.Sleep(time.Second)
		if ctx.Err() != nil {
			panic(ctx.Err())
		}
		conn, err = amqp.Dial("amqp://guest:guest@rabbit:5672/")
		if err != nil {
			slog.Error("rabbit connection failed", slog.Any("erorr", err))
			continue
		} else {
			break
		}
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"process_transaction", // name
		false,                 // durable
		false,                 // delete when unused
		false,                 // exclusive
		false,                 // no-wait
		nil,                   // arguments
	)
	if err != nil {
		panic(err)
	}

	if err := ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	); err != nil {
		panic(err)
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)

	var forever chan struct{}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		for d := range msgs {
			transactionId, err := strconv.Atoi(string(d.Body))
			if err != nil {
				panic(err)
			}

			slog.Info("processing transaction", slog.Any("id", transactionId))
			response, err := processTransaction(uint(transactionId))

			err = ch.PublishWithContext(ctx,
				"",        // exchange
				d.ReplyTo, // routing key
				false,     // mandatory
				false,     // immediate
				amqp.Publishing{
					ContentType:   "text/plain",
					CorrelationId: d.CorrelationId,
					Body:          []byte(fmt.Sprintf("%d", response)),
				})

			d.Ack(false)
		}
	}()

	slog.Info("listening for rpc requests")
	<-forever
}
