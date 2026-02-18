// Package router implements an MQTT topic routing system that manages topic subscriptions
// and efficiently matches topics with wildcards (+ and #) to corresponding client IDs.
package router

import (
	"strings"
	"sync"
)

// TopicMatcher represents the core topic routing structure that maintains
// the mapping between MQTT topics and client IDs for message routing.
// This structure uses a concurrent-safe map with RWMutex to support
// high-concurrency scenarios with many simultaneous subscriptions/unsubscriptions.
type TopicMatcher struct {
	// mu provides thread-safe access to the clients map
	mu sync.RWMutex
	// clients stores the mapping of topic patterns to client IDs
	// key: topic pattern (e.g. "sensor/+/temperature" or "devices/#")
	// value: list of client IDs that have subscribed to this topic
	clients map[string][]string
}

// NewTopicMatcher creates and initializes a new TopicMatcher instance
// with an empty subscription map. This is the constructor function
// for the TopicMatcher type.
//
// Returns:
//   - A pointer to the newly created TopicMatcher instance
func NewTopicMatcher() *TopicMatcher {
	return &TopicMatcher{
		clients: make(map[string][]string),
	}
}

// Subscribe adds a client to the subscription list for a specific topic.
// This method handles the registration of a client's interest in receiving
// messages published to the specified topic. The client ID will be added
// only if it's not already in the list for this topic.
//
// Parameters:
//   - topic: The MQTT topic pattern to subscribe to (supports wildcards + and #)
//   - clientID: The unique identifier of the client subscribing to the topic
//
// Thread Safety: This method uses a mutex to ensure thread-safe access to
// the shared clients map during the subscription process.
func (tm *TopicMatcher) Subscribe(topic, clientID string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	clientList := tm.clients[topic]
	for _, id := range clientList {
		if id == clientID {
			return
		}
	}

	tm.clients[topic] = append(tm.clients[topic], clientID)
}

// Unsubscribe removes a client from the subscription list for a specific topic.
func (tm *TopicMatcher) Unsubscribe(topic, clientID string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	clientList := tm.clients[topic]
	newList := []string{}

	for _, id := range clientList {
		if id != clientID {
			newList = append(newList, id)
		}
	}

	if len(newList) == 0 {
		delete(tm.clients, topic)
	} else {
		tm.clients[topic] = newList
	}
}

// GetClientsForTopic retrieves all client IDs that match the specified topic.
// This includes clients subscribed to the exact topic as well as clients
// subscribed to wildcard patterns that match the topic (e.g., client subscribed
// to "sensor/+" would receive messages sent to "sensor/temperature").
//
// Parameters:
//   - topic: The topic to match against existing subscriptions
//
// Returns:
//   - A slice of client IDs that should receive messages for the given topic
func (tm *TopicMatcher) GetClientsForTopic(topic string) []string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	var result []string
	seen := make(map[string]bool)

	if clients, exists := tm.clients[topic]; exists {
		for _, clientID := range clients {
			if !seen[clientID] {
				result = append(result, clientID)
				seen[clientID] = true
			}
		}
	}

	for storedTopic, clients := range tm.clients {
		if storedTopic != topic && matchTopic(storedTopic, topic) {
			for _, clientID := range clients {
				if !seen[clientID] {
					result = append(result, clientID)
					seen[clientID] = true
				}
			}
		}
	}

	return result
}

// matchTopic implements MQTT topic pattern matching with wildcards.
// Supports two types of wildcards:
// - '+' matches a single topic level (e.g. "sport/+/player" matches "sport/tennis/player")
// - '#' matches multiple topic levels (e.g. "sport/#" matches "sport/tennis/stats/players")
//
// Parameters:
//   - pattern: The topic pattern containing wildcards
//   - topic: The actual topic to match against the pattern
//
// Returns:
//   - true if the topic matches the pattern, false otherwise
func matchTopic(pattern, topic string) bool {
	patternParts := strings.Split(pattern, "/")
	topicParts := strings.Split(topic, "/")

	pIdx := 0
	tIdx := 0

	for pIdx < len(patternParts) && tIdx < len(topicParts) {
		p := patternParts[pIdx]

		if p == "#" {
			return true
		} else if p == "+" {
			tIdx++
		} else if p == topicParts[tIdx] {
			tIdx++
		} else {
			return false
		}

		pIdx++
	}

	if pIdx < len(patternParts) && patternParts[pIdx] == "#" {
		return true
	}

	return pIdx >= len(patternParts) && tIdx >= len(topicParts)
}

func (tm *TopicMatcher) GetMatchingTopics(pattern string) []string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	var matches []string

	for topic := range tm.clients {
		if matchTopic(pattern, topic) {
			matches = append(matches, topic)
		}
	}

	return matches
}

func (tm *TopicMatcher) GetAllSubscriptions(clientID string) []string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	var topics []string

	for topic, clients := range tm.clients {
		for _, id := range clients {
			if id == clientID {
				topics = append(topics, topic)
				break
			}
		}
	}

	return topics
}
