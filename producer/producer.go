package main

import (
	"encoding/json"
	"os"
	"time"

	"github.com/streadway/amqp"
)

/*
	produce submission result message
*/
func main() {
	// rabbitMQ server URL
	amqpServerURL := os.Getenv("AMQP_SERVER_URL")

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
	_, err = channelRabbitMQ.QueueDeclare(
		"result_queue", // queue name
		true,           // durable
		false,          // auto delete
		false,          // exclusive
		false,          // no wait
		nil,            // arguments
	)
	if err != nil {
		panic(err)
	}

	submission_result := make(map[string]interface{})
	submission_result["submission_id"] = 999
	submission_result["create_time"] = time.Now()
	submission_result["result"] = "Accepted"
	submission_result["accepted_number"] = 10
	submission_result["total_score"] = 100
	json, _ := json.Marshal(submission_result)

	message := amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte(json),
	}

	err = channelRabbitMQ.Publish(
		"",             // exchange
		"result_queue", // routing key
		false,          // mandatory
		false,          // immediate
		message,
	)
	if err != nil {
		panic(err)
	}
}
