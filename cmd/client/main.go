package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/aott33/peril/internal/gamelogic"
	"github.com/aott33/peril/internal/pubsub"
	"github.com/aott33/peril/internal/routing"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril client...")

	connectionString := "amqp://guest:guest@localhost:5672/"

	connection, err := amqp.Dial(connectionString)
	if err != nil {
		fmt.Println("Couldn't connect to rabbitmq", err)
		return
	}

	defer connection.Close()

	fmt.Println("Connection Successful!")

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		fmt.Println("Get username error", err)
		return
	}

	queueName := fmt.Sprintf("%s.%s", routing.PauseKey, username)

	_, _, err = pubsub.DeclareAndBind(
		connection,
		routing.ExchangePerilDirect,
		queueName,
		routing.PauseKey,
		pubsub.Transient(),
	)
	if err != nil {
		fmt.Println("Declare and bind queue error", err)
		return
	}

	gameState := gamelogic.NewGameState(username)

	for {
		words := gamelogic.GetInput()
		command := words[0]
		
		switch command {
		case "spawn":
			fmt.Println("Spawn requested...")

			err = gameState.CommandSpawn(words)
			if err != nil {
				fmt.Println(err)
			}

		case "move":
			fmt.Println("Move requested...")
			
			_, err = gameState.CommandMove(words)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("Move Successful!")
			}

		case "status":
			fmt.Println("Status requested...")

			gameState.CommandStatus()

		case "help":
			fmt.Println("Help requested...")

			gamelogic.PrintClientHelp()

		case "spam":
			fmt.Println("Spamming not allowed yet...")

		case "quit":
			gamelogic.PrintQuit()

		default:
			fmt.Println("Don't understand command")
		}

		if command == "quit" {
			break
		}
	}


	// wait for ctrl+c
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	s := <-signalChan
	fmt.Println("\nShutting down:", s)
}

