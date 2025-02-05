package pubsub

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	msg, err := json.Marshal(val)
	if err != nil {
		return fmt.Errorf("failed to marshal val: %w", err)
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
