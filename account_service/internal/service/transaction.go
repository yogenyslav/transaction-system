package service

import (
	"accountservice/internal/config"
	"accountservice/internal/model"
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type TransactionClient struct {
	conn  *amqp.Connection
	ch    *amqp.Channel
	qName string
	msgs  <-chan amqp.Delivery
}

func MustNewTransactionClient(cfg *config.Config) *TransactionClient {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	url := fmt.Sprintf(
		"amqp://%s:%s@%s:%d/",
		cfg.Rabbit.User,
		cfg.Rabbit.Password,
		cfg.Rabbit.Host,
		cfg.Rabbit.Port,
	)

	for {
		time.Sleep(time.Second)
		if ctx.Err() != nil {
			panic(ctx.Err())
		}
		conn, err := amqp.Dial(url)
		if err != nil {
			slog.Error("rabbit conn", slog.Any("error", err))
			continue
		}

		ch, err := conn.Channel()
		if err != nil {
			slog.Error("channel", slog.Any("error", err))
			continue
		}

		return &TransactionClient{
			conn: conn,
			ch:   ch,
		}
	}
}

func (p *TransactionClient) WithQueue(qName string, durable, autoDelete bool) *TransactionClient {
	q, err := p.ch.QueueDeclare(
		qName,
		durable,
		autoDelete,
		true,  // exclusive
		false, // no wait
		nil,
	)
	if err != nil {
		slog.Error("decalre queue", slog.Any("error", err))
		panic(err)
	}
	p.qName = q.Name
	return p
}

// consumer="" for random consumer name
func (p *TransactionClient) WithConsumer(consumer string) *TransactionClient {
	msgs, err := p.ch.Consume(
		p.qName,
		consumer,
		true,  // auto ack
		false, // exclusive
		false, // no local
		false, // no wait
		nil,
	)
	if err != nil {
		slog.Error("consumer", slog.Any("error", err))
		panic(err)
	}
	p.msgs = msgs
	return p
}

// timeout in seconds
func (p *TransactionClient) ProcessTransaction(ctx context.Context, transactionId uint) (model.Status, error) {
	status := model.Error

	id := fmt.Sprintf("%d", transactionId)
	if err := p.ch.PublishWithContext(ctx,
		"",
		"process_transaction",
		false,
		false,
		amqp.Publishing{
			ContentType:   "text/plain",
			CorrelationId: id,
			ReplyTo:       p.qName,
			Body:          []byte(id),
		},
	); err != nil {
		return status, err
	}

	var (
		result int
		err    error
	)
	for d := range p.msgs {
		if id == d.CorrelationId {
			result, err = strconv.Atoi(string(d.Body))
			break
		}
	}

	return model.Status(result), err
}

func (p *TransactionClient) Close() {
	p.conn.Close()
	p.ch.Close()
}
