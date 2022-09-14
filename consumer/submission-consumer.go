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

func NewConsumer(amqpURI, ctag string) (*Consumer, error) {
	c := &Consumer{
		connection: nil,
		channel:    nil,
		tag:        ctag,
		Done:       make(chan error),
	}

	var err error

	// Create New RabbitMQ Connection (go <-> RabbitMQ)
	config := amqp.Config{Properties: amqp.NewConnectionProperties()}
	config.Properties.SetClientConnectionName(constants.CONSUMER_CONNECTION)
	c.connection, err = amqp.DialConfig(amqpURI, config)
	if err != nil {
		return nil, fmt.Errorf("Dial: %s", err)
	}

	// Open a channel
	c.channel, err = c.connection.Channel()
	if err != nil {
		return nil, fmt.Errorf("Channel: %s", err)
	}

	// Set prefetchCount for consume channel
	if err = c.channel.Qos(
		1,     // prefetchCount
		0,     // prefetchSize
		false, // global
	); err != nil {
		return nil, fmt.Errorf("Qos Set: %s", err)
	}

	// Declare(Create) Exchange
	if err = c.channel.ExchangeDeclare(
		constants.SUBMISSION_EXCHANGE, // name of the exchange
		constants.DIRECT_TYPE,         // type
		true,                          // durable
		false,                         // delete when complete
		false,                         // internal(deprecated)
		false,                         // noWait
		nil,                           // arguments
	); err != nil {
		return nil, fmt.Errorf("Exchange Declare: %s", err)
	}

	// Declare(Create) Queue
	queue, err := c.channel.QueueDeclare(
		constants.SUBMISSION_QUEUE, // name of the queue
		true,                       // durable
		false,                      // delete when unused
		false,                      // exclusive
		false,                      // noWait
		nil,                        // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("Queue Declare: %s", err)
	}

	// Bind Queue to Exchange
	if err = c.channel.QueueBind(
		constants.SUBMISSION_QUEUE,    // name of the queue
		constants.SUBMISSION_KEY,      // bindingKey
		constants.SUBMISSION_EXCHANGE, // sourceExchange
		false,                         // noWait
		nil,                           // arguments
	); err != nil {
		return nil, fmt.Errorf("Queue Bind: %s", err)
	}

	// Subscribe queue for consume messages
	// Return `<- chan Delivery`
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
		return nil, fmt.Errorf("Queue Consume: %s", err)
	}

	// Receive Messages
	go handleMessages(messages, c.Done)

	return c, nil
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
