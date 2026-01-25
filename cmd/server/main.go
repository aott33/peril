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

	routingKey := fmt.Sprintf("%s.*", routing.GameLogSlug)
	
	_, _, err = pubsub.DeclareAndBind(
		connection,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		routingKey,
		pubsub.Durable,
	)
	if err != nil {
		fmt.Println("Declare and bind queue error", err)
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

	gamelogic.PrintServerHelp()

	loop:
	for {
		select {
		case <-done:
			return
		default:
			commands := gamelogic.GetInput()
			if len(commands) == 0 {
				continue
			}
			command := commands[0]
		
			switch command {
			case "pause":
				fmt.Println("Sending pause message")
				err = pubsub.PublishJSON(
					connChannel, 
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
			case "resume":
				fmt.Println("Sending resume message")
				err = pubsub.PublishJSON(
					connChannel, 
					string(routing.ExchangePerilDirect), 
					string(routing.PauseKey),
					routing.PlayingState{
							IsPaused: false,
						},
					)
				if err != nil {
					fmt.Println("Couldn't publish json", err)
					return
				}
			case "quit":
            	gamelogic.PrintQuit()
            	select {
           		case <-done:
            	default:
                	close(done)
            	}
            	break loop
			default:
				fmt.Println("Don't understand command")
			}
		}
	}

	fmt.Println("Have a good day")
}
