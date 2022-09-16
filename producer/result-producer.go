package producer

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-rabbit-handler/constants"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Producer struct {
	connection *amqp.Connection
	channel    *amqp.Channel
	Done       chan error
	publishes  chan uint64
}

func NewProducer() *Producer {
	return &Producer{
		connection: nil,
		channel:    nil,
		Done:       make(chan error),
		publishes:  make(chan uint64, 8),
	}
}

func (p *Producer) CreateConnection(amqpURI string) error {
	var err error

	config := amqp.Config{Properties: amqp.NewConnectionProperties()}
	config.Properties.SetClientConnectionName(constants.PRODUCER_CONNECTION)
	p.connection, err = amqp.DialConfig(amqpURI, config)
	if err != nil {
		return fmt.Errorf("dial: %s", err)
	}

	return nil
}

func (p *Producer) OpenChannel() error {
	var err error

	if p.channel, err = p.connection.Channel(); err != nil {
		return fmt.Errorf("channel: %s", err)
	}

	// put this channel into confirm mode
	// client can ensure all messages successfully received by server
	if err := p.channel.Confirm(false); err != nil {
		return fmt.Errorf("channel could not be put into confirm mode: %s", err)
	}
	// add listner for confirmation
	confirms := p.channel.NotifyPublish(make(chan amqp.Confirmation, 1))

	go p.confirmHandler(confirms)

	return nil
}

func (p *Producer) confirmHandler(confirms chan amqp.Confirmation) {
	m := make(map[uint64]bool)
	for {
		select {
		case <-p.Done:
			log.Println("confirmHandler is stopping")
			return
		case publishSeqNo := <-p.publishes:
			log.Printf("waiting for confirmation of %d", publishSeqNo)
			m[publishSeqNo] = false
		case confirmed := <-confirms:
			if confirmed.DeliveryTag > 0 {
				if confirmed.Ack {
					log.Printf("confirmed delivery with delivery tag: %d", confirmed.DeliveryTag)
				} else {
					fmt.Errorf("failed delivery of delivery tag: %d", confirmed.DeliveryTag)
				}
				delete(m, confirmed.DeliveryTag)
			}
		}
		if len(m) > 1 {
			log.Printf("outstanding confirmations: %d", len(m))
		}
	}
}

func (p *Producer) PublishMessage(body []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	seqNo := p.channel.GetNextPublishSeqNo()
	log.Printf("publishing %dB body (%q)", len(body), body)

	if err := p.channel.PublishWithContext(ctx,
		constants.RESULT_EXCHANGE, // publish to an exchange
		constants.RESULT_KEY,      // routing to 0 or more queues
		false,                     // mandatory
		false,                     // immediate
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     "text/plain",
			ContentEncoding: "",
			Body:            []byte(body),
			DeliveryMode:    amqp.Transient, // 1=non-persistent, 2=persistent
			Priority:        0,              // 0-9
		},
	); err != nil {
		return fmt.Errorf("exchange publish: %s", err)
	}

	log.Printf("published %dB OK", len(body))
	p.publishes <- seqNo

	return nil
}

func (p *Producer) CleanUp() error {
	if err := p.channel.Close(); err != nil {
		return fmt.Errorf("channel close failed: %s", err)
	}

	if err := p.connection.Close(); err != nil {
		return fmt.Errorf("connection close error: %s", err)
	}

	return <-p.Done
}
