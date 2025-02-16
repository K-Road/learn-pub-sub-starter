package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril server...")
	const rabbit_connection_string = "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(rabbit_connection_string)
	if err != nil {
		log.Fatalf("could not connect to RabbitMQ: %v", err)
	}
	defer conn.Close()
	fmt.Println("Connection was successful")

	publishCh, err := conn.Channel()
	if err != nil {
		log.Fatalf("could not open channel: %v", err)
	}
	//defer publishCh.Close()

	// _, q, err := pubsub.DeclareAndBind(
	// 	conn,
	// 	routing.ExchangePerilTopic,
	// 	routing.GameLogSlug,
	// 	routing.GameLogSlug+".*",
	// 	pubsub.SimpleQueueDurable)
	// if err != nil {
	// 	log.Fatalf("failing to subscribe to pause: %v", err)
	// }

	//fmt.Printf("Queue %v declared and bound!\n", q.Name)

	err = pubsub.SubscribeGob(
		conn,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		routing.GameLogSlug+".*",
		pubsub.SimpleQueueDurable,
		handlerLogs(),
	)
	if err != nil {
		log.Fatalf("could not start consuming logs: %v", err)
	}

	gamelogic.PrintServerHelp()
	for {
		input := gamelogic.GetInput()
		if len(input) == 0 {
			continue
		}

		switch input[0] {
		case "pause":
			fmt.Println("Publishing pause game state")
			err = pubsub.PublishJSON(
				publishCh,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{
					IsPaused: true,
				},
			)
			if err != nil {
				log.Printf("could not pause: %v", err)
			}
		case "resume":
			fmt.Println("Publishing resume game state")
			err = pubsub.PublishJSON(
				publishCh,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{
					IsPaused: false,
				},
			)
			if err != nil {
				log.Printf("could not resume: %v", err)
			}
		case "quit":
			log.Printf("logging %s\n", input[0])
			return
		default:
			fmt.Printf("invalid input: %s\n", input[0])

		}

		// signalChan := make(chan os.Signal, 1)
		// signal.Notify(signalChan, os.Interrupt)
		// <-signalChan
		//fmt.Printf("Shutting down\n")
	}
}
