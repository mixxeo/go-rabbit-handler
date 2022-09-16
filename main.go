package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-rabbit-handler/producer"
)

func main() {
	/*main for consumer*/
	/*
		c, err := consumer.NewConsumer("URI", "ctag")
		if err != nil {
			fmt.Errorf("%s", err)
		}

		// running until Consumer is done
		<-c.Done

		if err := c.CleanUp(); err != nil {
			fmt.Errorf("Error during clean up: %s", err)
		}
	*/

	/*main for producer*/
	var err error
	p := producer.NewProducer()

	err = p.CreateConnection("amqp://skku:1234@localhost:5672/%2f")
	if err != nil {
		fmt.Errorf("%s", err)
	}
	// connection close

	err = p.OpenChannel()
	if err != nil {
		fmt.Errorf("%s", err)
	}
	// channel close

	submission_result := make(map[string]interface{})
	submission_result["submission_id"] = 999
	submission_result["create_time"] = time.Now()
	body, _ := json.Marshal(submission_result)

	err = p.PublishMessage(body)
	if err != nil {
		fmt.Errorf("%s", err)
	}

	p.Done <- nil
}
