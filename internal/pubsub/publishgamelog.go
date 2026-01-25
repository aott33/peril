package pubsub

import (
	"fmt"

	"github.com/aott33/peril/internal/gamelogic"
	"github.com/aott33/peril/internal/routing"

	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishGameLog(ch *amqp.Channel, gs gamelogic.GameState, gl routing.GameLog) err {
	username := gs.GetUsername()
	gameLogKey := fmt.Sprintf("%s.%s", routing.GameLogSlug, username)

	err := PublishGob(
		ch,
		string(routing.ExchangePerilTopic),
		gameLogKey,
		gl,
	)
	if err != nil {
		ch.Nack()
	}

}
