package main

import (
	"fmt"

	"github.com/rabbit/consumer"
)

func main() {
	consumer, err := consumer.NewConsumer("URI", "ctag")

	// running until Consumer is done
	<-consumer.done

	if err := consumer.CleanUp(); err != nil {
		fmt.Errorf("Error during clean up: %s", err)
	}
}
