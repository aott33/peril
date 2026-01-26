package pubsub

import (
	"bytes"
	"encoding/gob"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func SubscribeGob[T any](
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

	err = ch.Qos(10, 0, false)
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
			reader := bytes.NewReader(msg.Body)
			decoder := gob.NewDecoder(reader)

			err = decoder.Decode(&body)
			if err != nil {
				fmt.Println(err)
				continue
			}
			acktype := handler(body)
			
			switch acktype {
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

