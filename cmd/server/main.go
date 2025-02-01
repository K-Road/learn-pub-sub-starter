package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

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

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("could not open channel: %v", err)
	}
	defer ch.Close()

	playingState := routing.PlayingState{
		IsPaused: true,
	}
	pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, playingState)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	fmt.Printf("Shutting down\n")

}
