package producer

import (
	"context"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func Publish(done chan bool, amqpURI, body string) error {
	// Create New RabbitMQ Connection (go <-> rabbitMQ)
	config := amqp.Config{Properties: amqp.NewConnectionProperties()}
	config.Properties.SetClientConnectionName("go-judger-producer")
	connection, err := amqp.DialConfig(amqpURI, config)
	if err != nil {
		return fmt.Errorf("Dial: %s", err)
	}
	defer connection.Close()

	// Open a channel
	channel, err := connection.Channel()
	if err != nil {
		return fmt.Errorf("Channel: %s", err)
	}

	// Declare(Create) Exchange
	if err := channel.ExchangeDeclare(
		"result-exchange", // name
		"direct",          // type
		true,              // durable
		false,             // auto-deleted
		false,             // internal
		false,             // noWait
		nil,               // arguments
	); err != nil {
		return fmt.Errorf("Exchange Declare: %s", err)
	}

	var publishes chan uint64 = nil
	var confirms chan amqp.Confirmation = nil

	// put this channel into confirm mode
	// client can ensure all messages successfully received by server
	if err := channel.Confirm(false); err != nil {
		return fmt.Errorf("Channel could not be put into confirm mode: %s", err)
	}
	// add listner for confirmation
	publishes = make(chan uint64, 8)
	confirms = channel.NotifyPublish(make(chan amqp.Confirmation, 1))

	go confirmHandler(done, publishes, confirms)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Publish Messages
	for {
		seqNo := channel.GetNextPublishSeqNo()
		log.Printf("publishing %dB body (%q)", len(body), body)

		if err := channel.PublishWithContext(ctx,
			"result-exchange", // publish to an exchange
			"result",          // routing to 0 or more queues
			false,             // mandatory
			false,             // immediate
			amqp.Publishing{
				Headers:         amqp.Table{},
				ContentType:     "text/plain",
				ContentEncoding: "",
				Body:            []byte(body),
				DeliveryMode:    amqp.Transient, // 1=non-persistent, 2=persistent
				Priority:        0,              // 0-9
				// a bunch of application/implementation-specific fields
			},
		); err != nil {
			return fmt.Errorf("Exchange Publish: %s", err)
		}

		log.Printf("published %dB OK", len(body))
		publishes <- seqNo

	}

	// return nil
}

func confirmHandler(done chan bool, publishes chan uint64, confirms chan amqp.Confirmation) {
	m := make(map[uint64]bool)
	for {
		select {
		case <-done:
			log.Println("confirmHandler is stopping")
			return
		case publishSeqNo := <-publishes:
			log.Printf("waiting for confirmation of %d", publishSeqNo)
			m[publishSeqNo] = false
		case confirmed := <-confirms:
			if confirmed.DeliveryTag > 0 {
				if confirmed.Ack {
					log.Printf("confirmed delivery with delivery tag: %d", confirmed.DeliveryTag)
				} else {
					fmt.Errorf("failed delivery of delivery tag: %d", confirmed.DeliveryTag)
				}
				delete(m, confirmed.DeliveryTag)
			}
		}
		if len(m) > 1 {
			log.Printf("outstanding confirmations: %d", len(m))
		}
	}
}
