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
    handler func(T)(AckType),
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
			achtype := handler(body)
			
			switch achtype {
			case Ack:
				msg.Ack(false)
				fmt.Println("Ack Occured")
			case NackRequeue:
				msg.Nack(false, true)
				fmt.Println("Nack Requeue Occured")
			case NackDiscard:
				msg.Nack(true, false)
				fmt.Println("Nack Discard Occured")
			}

			fmt.Print("> ")

		}
	}()

	return nil
}
