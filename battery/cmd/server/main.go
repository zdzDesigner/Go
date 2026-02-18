package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"battery/internal/broker"
)

var (
	address        = flag.String("address", ":1883", "Address to bind the MQTT broker")
	maxConnections = flag.Int("max-connections", 100000, "Maximum number of concurrent connections")
)

func main() {
	flag.Parse()

	log.Printf("Starting MQTT Broker with max %d connections on %s", *maxConnections, *address)

	b, err := broker.NewBroker(*address, *maxConnections)
	if err != nil {
		log.Fatal("Failed to create broker: ", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		b.Start()
	}()

	<-sigChan
	log.Println("Shutdown signal received, stopping broker...")
	b.Stop()
	log.Println("Broker stopped")
}
