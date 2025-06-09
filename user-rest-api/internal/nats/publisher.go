package nats

import (
	"github.com/nats-io/nats.go"
	"log"
)

var nc *nats.Conn

func InitNATS(url string) error {
	var err error
	nc, err = nats.Connect(url)
	if err != nil {
		return err
	}
	return nil
}

func Publish(subject string, data []byte) {
	if nc == nil {
		log.Println("NATS connection not initialized")
		return
	}
	if err := nc.Publish(subject, data); err != nil {
		log.Printf("Failed to publish to NATS subject %s: %v\n", subject, err)
	}
}
