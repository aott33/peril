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
	
	connChannel, err := connection.Channel()
	if err != nil {
		fmt.Println("Couldn't create connection channel", err)
		return
	}

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		fmt.Println("Get username error", err)
		return
	}

	gameState := gamelogic.NewGameState(username)
	
	queueName := fmt.Sprintf("%s.%s", routing.PauseKey, username)
	handler := handlerPause(gameState)

	err = pubsub.SubscribeJSON(
		connection,
		routing.ExchangePerilDirect,
		queueName,
		routing.PauseKey,
		pubsub.Transient(),
		handler,
	)
	if err != nil {
		fmt.Println("Subscribe error", err)
		return
	}


	armyMovesQueueName := fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, username)
	armyMovesRoutingKey := fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, "*")
	moveHandler := handlerMove(gameState)

	err = pubsub.SubscribeJSON(
		connection,
		routing.ExchangePerilTopic,
		armyMovesQueueName,
		armyMovesRoutingKey,
		pubsub.Transient(),
		moveHandler,
	)
	if err != nil {
		fmt.Println("Subscribe to move error", err)
		return
	}

	done := make(chan struct{})

	// listen for Ctrl+C
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	go func() {
    	<-signalChan              // wait for Ctrl+C
    	fmt.Println("\nCtrl+C pressed, shutting down...")
    	
		select {
		case <-done:

		default:
			close(done)
		}
	}()

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
			
			move, err := gameState.CommandMove(words)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("Move Successful!")
			}

			err = pubsub.PublishJSON(
				connChannel,
				string(routing.ExchangePerilTopic),
				armyMovesRoutingKey,
				move,
			)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("Publish Move Successful!")
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
			close(done)
			return

		default:
			fmt.Println("Don't understand command")
		}
	}


	fmt.Println("\nShutting down")
}

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) {
	return func(ps routing.PlayingState) {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
	}
}

func handlerMove(gs *gamelogic.GameState) func(gamelogic.ArmyMove) {
	return func(move gamelogic.ArmyMove) {
		defer fmt.Print("> ")
		gs.HandleMove(move)
	}
}
