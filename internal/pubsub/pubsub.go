package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {

	msg, err := json.Marshal(val)
	if err != nil {
		return fmt.Errorf("failed to marchal val: %w", err)
	}

	pubMsg := amqp.Publishing{
		ContentType: "application/json",
		Body:        msg,
	}

	//ctx = context.Background()
	err = ch.PublishWithContext(context.Background(),
		exchange,
		key,
		false,
		false,
		pubMsg,
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}
	return nil
}

type SimpleQueueType int

const (
	SimpleQueueDurable SimpleQueueType = iota
	SimpleQueueTransient
)

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
) (*amqp.Channel, amqp.Queue, error) {
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("failed to create connection: %w", err)
	}

	q, err := ch.QueueDeclare(
		queueName,
		simpleQueueType == SimpleQueueDurable,
		simpleQueueType != SimpleQueueDurable,
		simpleQueueType != SimpleQueueDurable,
		false,
		nil,
	)
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("failed to create queue: %w", err)
	}

	err = ch.QueueBind(q.Name, key, exchange, false, nil)
	if err != nil {
		return nil, q, fmt.Errorf("failed to bind queue: %w", err)
	}
	return ch, q, nil

}
