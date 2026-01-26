package main

import (
	"fmt"
	"time"

	"github.com/aott33/peril/internal/gamelogic"
	"github.com/aott33/peril/internal/pubsub"
	"github.com/aott33/peril/internal/routing"
	
	amqp "github.com/rabbitmq/amqp091-go"
)

func handleWarMessages(gs *gamelogic.GameState, ch *amqp.Channel) func(gamelogic.RecognitionOfWar) pubsub.AckType {
	return func(rw gamelogic.RecognitionOfWar) pubsub.AckType {
		defer fmt.Print("> ")
		warOutcome, winner, loser := gs.HandleWar(rw)

		var msg string
		logWorthy := false
		
		switch warOutcome {
			case gamelogic.WarOutcomeNotInvolved:
				return pubsub.NackRequeue

			case gamelogic.WarOutcomeNoUnits:
				return pubsub.NackDiscard

			case gamelogic.WarOutcomeOpponentWon, gamelogic.WarOutcomeYouWon:
				msg = fmt.Sprintf("%s won a war against %s", winner, loser)
				logWorthy = true

			case gamelogic.WarOutcomeDraw:
				msg = fmt.Sprintf("A war between %s and %s resulted in a draw", winner, loser)
				logWorthy = true

			default:
				fmt.Println("Unknown war outcome")
				return pubsub.NackDiscard
		}

		if logWorthy {
			log := routing.GameLog{
				CurrentTime: time.Now(),
				Username: gs.GetUsername(),
				Message: msg,
			}
			err := pubsub.PublishGameLog(ch, log)
			if err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		}
		return pubsub.NackDiscard
	}
}

