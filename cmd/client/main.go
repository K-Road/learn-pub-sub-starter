package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril client...")
	const rabbit_connection_string = "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(rabbit_connection_string)
	if err != nil {
		log.Fatalf("could not connect to RabbitMQ: %v", err)
	}
	defer conn.Close()
	fmt.Println("Connection was successful")

	user, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("failed to get user: %v", err)
	}

	ch, q, err := pubsub.DeclareAndBind(conn, routing.ExchangePerilDirect, fmt.Sprintf("%s.%s", routing.PauseKey, user), routing.PauseKey, 0)
	if err != nil {
		log.Fatalf("failing to subscribe to pause: %v", err)
	}
	defer ch.Close()
	fmt.Printf("Queue %v declared and bound!\n", q.Name)

	fmt.Println(q)

	gameState := gamelogic.NewGameState(user)

	for {
		input := gamelogic.GetInput()
		if len(input) == 0 {
			continue
		}

		switch input[0] {
		case "spawn":
			fmt.Println("spawn")
		case "move":
			fmt.Println("move")
		case "status":
			gameState.CommandStatus()
		case "help":
			fmt.Println("help")
		case "spam":
			fmt.Println("spam")
		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			fmt.Printf("Invalid input: %s", input[0])
		}
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	fmt.Printf("Shutting down\n")

}
