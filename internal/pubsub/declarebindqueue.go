package pubsub

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int

const (
	durable SimpleQueueType = iota
	transient
)

func Durable() SimpleQueueType {
	return durable
}

func Transient() SimpleQueueType {
	return transient
}

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
	if queueType == durable {
		isDurable = true
	}

	isTransient := false 
	if queueType == transient {
		isTransient = true
	}

	queue, err = channel.QueueDeclare(
		queueName,
		isDurable,
		isTransient,
		isTransient,
		false,
		nil,
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
