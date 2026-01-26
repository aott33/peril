package pubsub

import (
	"fmt"

	"github.com/aott33/peril/internal/routing"

	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishGameLog(ch *amqp.Channel, gl routing.GameLog) error {
	username := gl.Username
	gameLogKey := fmt.Sprintf("%s.%s", routing.GameLogSlug, username)

	err := PublishGob(
		ch,
		routing.ExchangePerilTopic,
		gameLogKey,
		gl,
	)
	if err != nil {
		return err
	}

	return nil
}
