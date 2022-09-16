package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-rabbit-handler/producer"
)

// import (
// 	"fmt"

// 	"github.com/go-rabbit-handler/consumer"
// 	"github.com/go-rabbit-handler/producer"
// )

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
	done := make(chan bool)
	publishes := make(chan uint64, 8)

	producerConnection, err := producer.CreateConnection()
	if err != nil {
		fmt.Errorf("%s", err)
	}
	defer producerConnection.Close()

	producerChannel, err := producer.OpenChannel(producerConnection, done, publishes)
	if err != nil {
		fmt.Errorf("%s", err)
	}
	defer producerChannel.Close()

	submission_result := make(map[string]interface{})
	submission_result["submission_id"] = 999
	submission_result["create_time"] = time.Now()
	body, _ := json.Marshal(submission_result)

	err = producer.PublishMessage(producerChannel, publishes, body)
	if err != nil {
		fmt.Errorf("%s", err)
	}

	done <- true
}
