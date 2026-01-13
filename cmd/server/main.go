package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/aott33/peril/internal/pubsub"
	"github.com/aott33/peril/internal/routing"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril server...")

	connectionString := "amqp://guest:guest@localhost:5672/"

	connection, err := amqp.Dial(connectionString)
	if err != nil {
		fmt.Println("Couldn't connect to rabbitmq", err)
		return
	}

	defer connection.Close()

	fmt.Println("Connection Successful!")

	connChannel, err := connection.Channel()
	if err != nil {
		fmt.Println("Couldn't create connection channel", err)
		return
	}

	err = pubsub.PublishJSON(connChannel, 
		string(routing.ExchangePerilDirect), 
		string(routing.PauseKey),
		routing.PlayingState{
			IsPaused: true,
		},
	)
	if err != nil {
		fmt.Println("Couldn't publish json", err)
		return
	}

	// wait for ctrl+c
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	s := <-signalChan
	fmt.Println("\nShutting down:", s)
}
