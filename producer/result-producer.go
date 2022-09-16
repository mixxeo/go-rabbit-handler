package producer

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-rabbit-handler/constants"
	amqp "github.com/rabbitmq/amqp091-go"
)

func CreateConnection() (*amqp.Connection, error) {
	// Create New RabbitMQ Connection (go <-> rabbitMQ)
	config := amqp.Config{Properties: amqp.NewConnectionProperties()}
	config.Properties.SetClientConnectionName(constants.PRODUCER_CONNECTION)
	connection, err := amqp.DialConfig("amqURI", config)
	if err != nil {
		return nil, fmt.Errorf("dial: %s", err)
	}

	return connection, nil
}

func OpenChannel(connection *amqp.Connection, done chan bool, publishes chan uint64) (*amqp.Channel, error) {
	channel, err := connection.Channel()
	if err != nil {
		return nil, fmt.Errorf("channel: %s", err)
	}

	// put this channel into confirm mode
	// client can ensure all messages successfully received by server
	if err := channel.Confirm(false); err != nil {
		return nil, fmt.Errorf("channel could not be put into confirm mode: %s", err)
	}
	// add listner for confirmation
	confirms := channel.NotifyPublish(make(chan amqp.Confirmation, 1))

	go confirmHandler(done, publishes, confirms)

	return channel, nil
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

func PublishMessage(channel *amqp.Channel, publishes chan uint64, body []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	seqNo := channel.GetNextPublishSeqNo()
	log.Printf("publishing %dB body (%q)", len(body), body)

	if err := channel.PublishWithContext(ctx,
		constants.RESULT_EXCHANGE, // publish to an exchange
		constants.RESULT_KEY,      // routing to 0 or more queues
		false,                     // mandatory
		false,                     // immediate
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     "text/plain",
			ContentEncoding: "",
			Body:            []byte(body),
			DeliveryMode:    amqp.Transient, // 1=non-persistent, 2=persistent
			Priority:        0,              // 0-9
		},
	); err != nil {
		return fmt.Errorf("Exchange Publish: %s", err)
	}

	log.Printf("published %dB OK", len(body))
	publishes <- seqNo

	return nil
}
