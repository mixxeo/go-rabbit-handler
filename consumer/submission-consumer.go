package consumer

import (
	"fmt"
	"log"

	"github.com/go-rabbit-handler/constants"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	connection *amqp.Connection
	channel    *amqp.Channel
	tag        string
	Done       chan error
}

func NewConsumer(ctag string) *Consumer {
	return &Consumer{
		connection: nil,
		channel:    nil,
		tag:        ctag,
		Done:       make(chan error),
	}
}

func (c *Consumer) CreateConnection(amqpURI string) error {
	var err error

	config := amqp.Config{Properties: amqp.NewConnectionProperties()}
	config.Properties.SetClientConnectionName(constants.CONSUMER_CONNECTION)
	c.connection, err = amqp.DialConfig(amqpURI, config)
	if err != nil {
		return fmt.Errorf("dial: %s", err)
	}

	return nil
}

func (c *Consumer) OpenChannel() error {
	var err error

	if c.channel, err = c.connection.Channel(); err != nil {
		return fmt.Errorf("channel: %s", err)
	}

	// Set prefetchCount for consume channel
	if err = c.channel.Qos(
		1,     // prefetchCount
		0,     // prefetchSize
		false, // global
	); err != nil {
		return fmt.Errorf("qos set: %s", err)
	}

	return nil

}

func (c *Consumer) Subscribe() error {
	messages, err := c.channel.Consume(
		constants.SUBMISSION_QUEUE, // queue name
		c.tag,                      // consumer
		false,                      // autoAck
		false,                      // exclusive
		false,                      // noLocal
		false,                      // noWait
		nil,                        // arguments
	)
	if err != nil {
		return fmt.Errorf("queue consume: %s", err)
	}

	// Receive Messages
	go handleMessages(messages, c.Done)

	return nil
}

func handleMessages(messages <-chan amqp.Delivery, done chan error) {
	clean := func() {
		done <- nil
	}
	defer clean()

	for m := range messages {
		log.Printf("Message: %s\n", m.Body)

		m.Ack(false)
	}
}

func (c *Consumer) CleanUp() error {
	// Close channel
	if err := c.channel.Cancel(c.tag, true); err != nil {
		return fmt.Errorf("Consumer cancel failed: %s", err)
	}

	// Close Connection
	if err := c.connection.Close(); err != nil {
		return fmt.Errorf("AMQP connection close error: %s", err)
	}
	defer log.Print("RabbitMQ connection clear done")

	// wait for handle() to exit
	return <-c.Done
}
