package main

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
	/*
		done := make(chan bool)

		if err := producer.Publish(done, "URI", "submission result"); err != nil {
			fmt.Errorf("%s", err)
		}
	*/
}
