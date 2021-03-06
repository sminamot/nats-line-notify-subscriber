package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"

	"github.com/nats-io/nats.go"
	line "github.com/sminamot/nats-line-notify"
)

// environment variable
var (
	natsServer      string
	natsChannel     string
	natsQueueGroup  string
	maxRequeueCount int
)

func init() {
	natsServer = os.Getenv("NATS_SERVER")
	natsChannel = os.Getenv("NATS_CHANNEL")
	natsQueueGroup = os.Getenv("NATS_QUEUE_GROUP")
	var err error
	maxRequeueCount, err = strconv.Atoi(os.Getenv("MAX_REQUEUE_COUNT"))
	if err != nil {
		log.Fatalln("MAX_REQUEUE_COUNT is must be a numeric value")
	}

	switch "" {
	case natsServer, natsChannel, natsQueueGroup:
		log.Fatalln("specify environment variable")
	}
}

func main() {
	nc, err := nats.Connect(natsServer,
		nats.ErrorHandler(func(nc *nats.Conn, s *nats.Subscription, err error) {
			if s != nil {
				log.Printf("Async error in %q/%q: %v", s.Subject, s.Queue, err)
			} else {
				log.Printf("Async error outside subscription: %v", err)
			}
		}))
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()
	ec, err := nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	if err != nil {
		log.Fatal(err)
	}
	defer ec.Close()

	subscribeFunc := func(s *line.Line) {
		isRetry := s.RetryCount != 0
		if isRetry {
			fmt.Printf("Received a retried(%d) message\n", s.RetryCount)
		} else {
			fmt.Println("Received a new message")
		}
		if err := s.Notify(); err != nil && s.RetryCount < maxRequeueCount {
			s.RetryCount++
			if err := ec.Publish(natsChannel, s); err != nil {
				log.Fatal(err)
			}
		}
	}
	// Subscribe
	// Decoding errors will be passed to the function supplied via
	// nats.ErrorHandler above, and the callback supplied here will
	// not be invoked.
	if _, err := ec.QueueSubscribe(natsChannel, natsQueueGroup, subscribeFunc); err != nil {
		log.Fatal(err)
	}

	nc.Flush()

	if err := nc.LastError(); err != nil {
		log.Fatal(err)
	}

	// Setup the interrupt handler to drain so we don't miss
	// requests when scaling down.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	log.Println()
	log.Printf("Draining...")
	nc.Drain()
	log.Fatalf("Exiting")
}
