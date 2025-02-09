package pubsub

import (
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Acktype int
type SimpleQueueType int

const (
	Ack Acktype = iota
	NackDiscard
	NackRequeue
)

const (
	SimpleQueueDurable SimpleQueueType = iota
	SimpleQueueTransient
)

func SubscribeJSON[T any](conn *amqp.Connection, exchange, queueName, key string, simpleQueueType SimpleQueueType, handler func(T) Acktype) error {
	ch, _, err := DeclareAndBind(conn, exchange, queueName, key, simpleQueueType)
	if err != nil {
		return fmt.Errorf("failed to bind: %w", err)
	}

	//delivery channel
	delch, err := ch.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("failed to create del channel: %w", err)
	}

	unmarshaller := func(data []byte) (T, error) {
		var target T
		err := json.Unmarshal(data, &target)
		return target, err
	}

	go func() {
		defer ch.Close()
		for msg := range delch {
			target, err := unmarshaller(msg.Body)
			if err != nil {
				fmt.Printf("failed to unmarshal msg: %v", err)
				continue
			}
			switch handler(target) {
			case Ack:
				msg.Ack(false)
				//fmt.Println("Ack")
			case NackDiscard:
				msg.Nack(false, false)
				//fmt.Println("NackDiscard")
			case NackRequeue:
				msg.Nack(false, true)
				//fmt.Println("NackRequeue")
			}
		}
	}()
	return nil
}

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
) (*amqp.Channel, amqp.Queue, error) {
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("failed to create connection: %v", err)
	}

	args := amqp.Table{
		"x-dead-letter-exchange": "peril_dlx",
	}

	q, err := ch.QueueDeclare(
		queueName,
		simpleQueueType == SimpleQueueDurable,
		simpleQueueType != SimpleQueueDurable,
		simpleQueueType != SimpleQueueDurable,
		false,
		args,
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
