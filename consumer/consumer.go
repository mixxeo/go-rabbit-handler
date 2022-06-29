package consumer

import (
	"log"

	"github.com/streadway/amqp"
)

/*
	consume submission message
*/
func main() {
	// rabbitMQ server URL
	amqpServerURL := "AMQP_SERVER_URL"

	// create new rabbitMQ connection
	connectRabbitMQ, err := amqp.Dial(amqpServerURL)
	if err != nil {
		panic(err)
	}
	defer connectRabbitMQ.Close()

	// open a channel
	channelRabbitMQ, err := connectRabbitMQ.Channel()
	if err != nil {
		panic(err)
	}
	defer channelRabbitMQ.Close()

	// subscribe queue for getting messages
	messages, err := channelRabbitMQ.Consume(
		"submission_queue", // queue name
		"",                 // consumer
		true,               // auto-ack
		false,              // exclusive
		false,              // no local
		false,              // no wait
		nil,                // arguments
	)
	if err != nil {
		log.Println(err)
	}

	log.Println("Successfully connected to RabbitMQ")

	// receive messages into infinite loop
	forever := make(chan bool)

	go func() {
		for message := range messages {
			log.Printf("Message: %s\n", message.Body)
		}
	}()
	<-forever
}
