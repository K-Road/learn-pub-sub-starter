package main

import (
	"fmt"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	"github.com/rabbitmq/amqp091-go"
)

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.Acktype {
	return func(state routing.PlayingState) pubsub.Acktype {
		defer fmt.Print("> ")
		gs.HandlePause(state)
		return pubsub.Ack
	}
}

func handlerMove(gs *gamelogic.GameState, ch *amqp091.Channel) func(gamelogic.ArmyMove) pubsub.Acktype {
	return func(move gamelogic.ArmyMove) pubsub.Acktype {
		defer fmt.Print("> ")
		move_outcome := gs.HandleMove(move)

		switch move_outcome {
		case gamelogic.MoveOutComeSafe:
			return pubsub.Ack
		case gamelogic.MoveOutcomeMakeWar:
			err := pubsub.PublishJSON(
				ch,
				routing.ExchangePerilTopic,
				routing.WarRecognitionsPrefix+"."+gs.GetUsername(),
				gamelogic.RecognitionOfWar{
					Attacker: move.Player,
					Defender: gs.GetPlayerSnap(),
				},
			)
			if err != nil {
				fmt.Printf("Failed to publish war recognition message: %v\n", err)
				return pubsub.NackRequeue // Retry the message
			}
			return pubsub.Ack
		case gamelogic.MoveOutcomeSamePlayer:
			return pubsub.Ack
		}

		fmt.Println("error: unknown move outcome")
		return pubsub.NackDiscard
	}
}

func handlerWar(gs *gamelogic.GameState, ch *amqp091.Channel) func(rw gamelogic.RecognitionOfWar) pubsub.Acktype {
	return func(rw gamelogic.RecognitionOfWar) pubsub.Acktype {

		defer fmt.Print("> ")

		outcome, winner, loser := gs.HandleWar(rw)

		var ack pubsub.Acktype
		var msg string
		switch outcome {
		case gamelogic.WarOutcomeNotInvolved:
			ack = pubsub.NackRequeue
		case gamelogic.WarOutcomeNoUnits:
			ack = pubsub.NackDiscard
		case gamelogic.WarOutcomeOpponentWon:
			msg = fmt.Sprintf("%s won a war against %s", winner, loser)
			ack = pubsub.Ack
		case gamelogic.WarOutcomeYouWon:
			msg = fmt.Sprintf("%s won a war against %s", winner, loser)
			ack = pubsub.Ack
		case gamelogic.WarOutcomeDraw:
			msg = fmt.Sprintf("A war between %s and %s resulted in a draw", winner, loser)
			ack = pubsub.Ack
		default:
			fmt.Println("Failed war")
			return pubsub.NackDiscard
		}

		err := publishGameLog(
			ch,
			gs.GetUsername(),
			msg,
		)
		if err != nil {
			return pubsub.NackRequeue
		}
		return ack
	}
}
