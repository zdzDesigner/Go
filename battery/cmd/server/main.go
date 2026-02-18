// Package main implements the entry point for the MQTT broker server.
// This command-line application creates and manages an MQTT broker instance,
// handling configuration, startup, and graceful shutdown procedures.
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
	// address specifies the network address where the MQTT broker will listen for connections
	// Default value is ":1883" which binds to all interfaces on port 1883
	address = flag.String("address", ":1883", "Address to bind the MQTT broker")
	// maxConnections sets the maximum number of concurrent client connections allowed
	// Default value is 100,000 to support high-scale deployments
	maxConnections = flag.Int("max-connections", 100000, "Maximum number of concurrent connections")
)

// main is the entry point for the MQTT broker server application.
// It handles command-line argument parsing, broker initialization,
// signal handling for graceful shutdown, and broker lifecycle management.
func main() {
	// Parse command-line flags to configure the broker
	flag.Parse()

	// Log startup information with configured parameters
	log.Printf("Starting MQTT Broker with max %d connections on %s", *maxConnections, *address)

	// Create a new broker instance with the specified configuration
	b, err := broker.NewBroker(*address, *maxConnections)
	if err != nil {
		// Fatal error if broker creation fails (usually due to network binding issues)
		log.Fatal("Failed to create broker: ", err)
	}

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	// Notify channel on receiving interrupt (Ctrl+C) or termination signals
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start the broker in a separate goroutine to allow for graceful shutdown
	go func() {
		b.Start()
	}()

	// Block until a signal is received
	<-sigChan
	// Log shutdown initiation and stop the broker gracefully
	log.Println("Shutdown signal received, stopping broker...")
	b.Stop()
	log.Println("Broker stopped")
}
