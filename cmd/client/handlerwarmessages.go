package main

import (
	"fmt"

	"github.com/aott33/peril/internal/gamelogic"
	"github.com/aott33/peril/internal/pubsub"
	
)

func handleWarMessages(gs *gamelogic.GameState) func(gamelogic.RecognitionOfWar) pubsub.AckType {
	return func(rw gamelogic.RecognitionOfWar) pubsub.AckType {
		defer fmt.Print("> ")
		warOutcome, _, _ := gs.HandleWar(rw)
		
		switch warOutcome {
			case gamelogic.WarOutcomeNotInvolved:
				return pubsub.NackRequeue

			case gamelogic.WarOutcomeNoUnits:
				return pubsub.NackDiscard

			case gamelogic.WarOutcomeOpponentWon, gamelogic.WarOutcomeYouWon, gamelogic.WarOutcomeDraw:
				return pubsub.Ack
			default:
				fmt.Println("Unknown war outcome")
				return pubsub.NackDiscard
		}
	}
}

