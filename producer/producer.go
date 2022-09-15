package producer

import (
	"fmt"

	"github.com/go-rabbit-handler/constants"
	amqp "github.com/rabbitmq/amqp091-go"
)

func CreateConnection() (*amqp.Connection, error) {
	// Create New RabbitMQ Connection (go <-> rabbitMQ)
	config := amqp.Config{Properties: amqp.NewConnectionProperties()}
	config.Properties.SetClientConnectionName(constants.PRODUCER_CONNECTION)
	connection, err := amqp.DialConfig("amqpURI", config)
	if err != nil {
		return nil, fmt.Errorf("Dial: %s", err)
	}
	//defer connection.Close()

	return connection, nil
}

func CreateChannel(connection *amqp.Connection) (*amqp.Channel, error) {
	channel, err := connection.Channel()
	if err != nil {
		return nil, fmt.Errorf("Channel: %s", err)
	}

	return channel, nil
}
