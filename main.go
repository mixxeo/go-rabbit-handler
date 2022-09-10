package main

import (
	"fmt"

	"github.com/go-rabbit-handler/consumer"
)

func main() {
	c, err := consumer.NewConsumer("URI", "ctag")
	if err != nil {
		fmt.Errorf("%s", err)
	}

	// running until Consumer is done
	<-c.Done

	if err := c.CleanUp(); err != nil {
		fmt.Errorf("Error during clean up: %s", err)
	}
}
