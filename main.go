package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/go-rabbit-handler/constants"
	"github.com/go-rabbit-handler/consumer"
	"github.com/go-rabbit-handler/producer"
	amqp "github.com/rabbitmq/amqp091-go"
)

func CreateConnection(amqpURI string) (*amqp.Connection, error) {
	config := amqp.Config{Properties: amqp.NewConnectionProperties()}
	config.Properties.SetClientConnectionName(constants.CONNECTION)
	connection, err := amqp.DialConfig(amqpURI, config)
	if err != nil {
		return nil, fmt.Errorf("dial: %s", err)
	}

	return connection, nil
}

func main() {
	/*Open a TCP Connection*/
	amqpURI := os.Getenv("amqpURI")
	var err error
	connection, err := CreateConnection(amqpURI)
	if err != nil {
		fmt.Printf("%s", err)
	}

	/*Create Consumer*/
	c := consumer.NewConsumer(connection, constants.CONSUMER)

	if err = c.OpenChannel(); err != nil {
		fmt.Printf("%s", err)
	}

	if err = c.Subscribe(); err != nil {
		fmt.Printf("%s", err)
	}
	defer c.CleanUp()

	/*Create Producer*/
	p := producer.NewProducer(connection)

	if err = p.OpenChannel(); err != nil {
		fmt.Printf("%s", err)
	}

	submission_result := make(map[string]interface{})
	submission_result["submission_id"] = 999
	submission_result["create_time"] = time.Now()
	body, _ := json.Marshal(submission_result)

	err = p.PublishMessage(body)
	if err != nil {
		fmt.Printf("%s", err)
	}
	defer p.CleanUp()
}
