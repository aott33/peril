package pubsub

import (
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func SubscribeJSON[T any](
    conn *amqp.Connection,
    exchange,
    queueName,
    key string,
    queueType SimpleQueueType, // an enum to represent "durable" or "transient"
    handler func(T),
) error {
	ch, queue, err:= DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err
	}

	deliveryCh, err := ch.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	msgs := deliveryCh
	go func() {
		for msg := range msgs {
			var body T
			unmarshalErr := json.Unmarshal(msg.Body, &body)
			if unmarshalErr != nil {
				fmt.Println(unmarshalErr)
				continue
			}
			handler(body)
			msg.Ack(false)
		}
	}()

	return nil
}
