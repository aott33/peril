package main

import (
	"fmt"
	"time"

	"github.com/aott33/peril/internal/gamelogic"
	"github.com/aott33/peril/internal/pubsub"
	"github.com/aott33/peril/internal/routing"
)

func handleWarMessages(gs *gamelogic.GameState) func(gamelogic.RecognitionOfWar) pubsub.AckType {
	return func(rw gamelogic.RecognitionOfWar) pubsub.AckType {
		defer fmt.Print("> ")
		warOutcome, winner, loser := gs.HandleWar(rw)
		
		switch warOutcome {
			case gamelogic.WarOutcomeNotInvolved:
				return pubsub.NackRequeue

			case gamelogic.WarOutcomeNoUnits:
				return pubsub.NackDiscard

			case gamelogic.WarOutcomeOpponentWon, gamelogic.WarOutcomeYouWon:
				msg := fmt.Sprintf("%s won a war agains %s", winner, loser)
				log := routing.GameLog{
					CurrentTime: time.Now(),
					Username: gs.GetUsername(),
					Message: msg,
				}
				return pubsub.Ack

			case gamelogic.WarOutcomeDraw:
				msg := fmt.Sprintf("A war beetween %s and %s resulted in a draw", winner, loser)
				log := routing.GameLog{
					CurrentTime: time.Now(),
					Username: gs.GetUsername(),
					Message: msg,
				}
				return pubsub.Ack

			default:
				fmt.Println("Unknown war outcome")
				return pubsub.NackDiscard
		}
	}
}

