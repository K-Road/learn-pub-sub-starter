package pubsub

import (
	"bytes"
	"context"
	"encoding/gob"
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

func PublishGob[T any](ch *amqp.Channel, exchange, key string, val T) error {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	err := enc.Encode(val)
	if err != nil {
		return fmt.Errorf("failed to encode: %w", err)
	}

	pubMsg := amqp.Publishing{
		ContentType: "application/gob",
		Body:        buf.Bytes(),
	}

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
