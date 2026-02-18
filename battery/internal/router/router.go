package router

import (
	"strings"
	"sync"
)

type TopicMatcher struct {
	mu      sync.RWMutex
	clients map[string][]string
}

func NewTopicMatcher() *TopicMatcher {
	return &TopicMatcher{
		clients: make(map[string][]string),
	}
}

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
