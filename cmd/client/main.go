package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"time"

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
		pubsub.Transient,
		handler,
	)
	if err != nil {
		fmt.Println("Subscribe error", err)
		return
	}

	armyMovesQueueName := fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, username)
	armyMovesRoutingKey := fmt.Sprintf("%s.*", routing.ArmyMovesPrefix)
	moveHandler := handlerMove(gameState, connChannel)

	err = pubsub.SubscribeJSON(
		connection,
		routing.ExchangePerilTopic,
		armyMovesQueueName,
		armyMovesRoutingKey,
		pubsub.Transient,
		moveHandler,
	)
	if err != nil {
		fmt.Println("Subscribe to move error", err)
		return
	}

	warMessagesQueueName := "war"
	warMessagesRoutingKey := fmt.Sprintf("%s.*", routing.WarRecognitionsPrefix)
	warMessagesHandler := handleWarMessages(gameState, connChannel)

	err = pubsub.SubscribeJSON(
		connection,
		routing.ExchangePerilTopic,
		warMessagesQueueName,
		warMessagesRoutingKey,
		pubsub.Durable,
		warMessagesHandler,
	)
	if err != nil {
		fmt.Println("Subscribe to war messages error", err)
		return
	}

	done := make(chan struct{})

	// listen for Ctrl+C
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	go func() {
    	<-signalChan
		fmt.Println("\nCtrl+C pressed, shutting down...")
		os.Exit(0)
	}()

	loop:
	for {
		words := gamelogic.GetInput()
		command := words[0]

		select {
    	case <-done:
        	break loop
    	default:
    	}
		
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
			
			armyMovePublishKey := fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, username)
			err = pubsub.PublishJSON(
				connChannel,
				string(routing.ExchangePerilTopic),
				armyMovePublishKey,
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
			if len(words) < 2 {
				fmt.Println("Specify number. For example... spam 10")
				continue
			}
			n := words[1]
			intN, err := strconv.Atoi(n)
			if err != nil {
				fmt.Println(err)
				continue
			}
			for range intN {
				msg := gamelogic.GetMaliciousLog()
				gl := routing.GameLog{
					Username: username,
					CurrentTime: time.Now(),
					Message: msg,
				}
				err = pubsub.PublishGameLog(connChannel, gl)
				if err != nil {
					fmt.Println(err)
					continue
				}
			}

		case "quit":
			gamelogic.PrintQuit()
			break loop
		default:
			fmt.Println("Don't understand command")
		}
	}


	fmt.Println("Have a good day")
}
