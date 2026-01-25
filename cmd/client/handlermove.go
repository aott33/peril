package main

import (
	"fmt"

	"github.com/aott33/peril/internal/gamelogic"
	"github.com/aott33/peril/internal/pubsub"
	"github.com/aott33/peril/internal/routing"
	
	amqp "github.com/rabbitmq/amqp091-go"
)

func handlerMove(gs *gamelogic.GameState, ch *amqp.Channel) func(gamelogic.ArmyMove) pubsub.AckType {
	return func(move gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Print("> ")
		moveOutcome := gs.HandleMove(move)
		
		switch moveOutcome {
			case gamelogic.MoveOutComeSafe:
				return pubsub.Ack
			case gamelogic.MoveOutcomeMakeWar:
				fmt.Println("Sending war recognition message")
				
				warRecKey := fmt.Sprintf("%s.%s", routing.WarRecognitionsPrefix, gs.GetUsername())
				warRecStruct := gamelogic.RecognitionOfWar{
					Attacker: move.Player,
					Defender: gs.GetPlayerSnap(),
				}
				
				err := pubsub.PublishJSON(
					ch, 
					string(routing.ExchangePerilTopic), 
					warRecKey,
					warRecStruct,
				)
				if err != nil {
					fmt.Println("Couldn't publish war recognition", err)
					return pubsub.NackRequeue
				}
				
				return pubsub.Ack

			default:
				return pubsub.NackDiscard
		}
	}
}
