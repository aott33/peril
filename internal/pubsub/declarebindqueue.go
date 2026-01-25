package pubsub

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {
	var channel *amqp.Channel
	var queue amqp.Queue

	channel, err := conn.Channel()
	if err != nil {
		return nil, queue, err
	}

	isDurable := false
	if queueType == Durable {
		isDurable = true
	}

	isTransient := false 
	if queueType == Transient {
		isTransient = true
	}

	queue, err = channel.QueueDeclare(
		queueName,
		isDurable,
		isTransient,
		isTransient,
		false,
		amqp.Table{
			"x-dead-letter-exchange": "peril_dlx",
		},
	)
	if err != nil {
		return nil, queue, err
	}
	
	err = channel.QueueBind(queueName, key, exchange, false, nil)
	if err != nil {
		return nil, queue, err
	}

	return channel, queue, nil	
}
