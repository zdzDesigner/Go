// Package router 实现了一个MQTT主题路由系统，管理主题订阅
// 并高效地将带有通配符（+ 和 #）的主题与相应的客户端ID进行匹配。
package router

import (
	"strings"
	"sync"
)

// TopicMatcher 表示核心主题路由结构，维护MQTT主题和客户端ID之间的映射
// 用于消息路由。此结构使用带RWMutex的并发安全映射来支持
// 高并发场景下的大量同时订阅/取消订阅。
type TopicMatcher struct {
	// mu 提供对客户端映射的安全访问
	mu sync.RWMutex
	// clients 存储主题模式到客户端ID的映射
	// key: 主题模式 (例如 "sensor/+/temperature" 或 "devices/#")
	// value: 已订阅此主题的客户端ID列表
	clients map[string][]string
}

// NewTopicMatcher 创建并初始化一个新的 TopicMatcher 实例
// 使用一个空的订阅映射。这是 TopicMatcher 类型的构造函数。
//
// 返回值：
//   - 指向新创建的 TopicMatcher 实例的指针
func NewTopicMatcher() *TopicMatcher {
	return &TopicMatcher{
		clients: make(map[string][]string),
	}
}

// Subscribe 将客户端添加到特定主题的订阅列表中。
// 此方法处理客户端对接收发布到指定主题的消息的兴趣注册。
// 仅当客户端ID尚未在此主题的列表中时才会添加。
//
// 参数：
//   - topic: 要订阅的MQTT主题模式（支持通配符 + 和 #）
//   - clientID: 订阅该主题的客户端的唯一标识符
//
// 线程安全：此方法使用互斥锁确保在订阅过程中对共享客户端映射的线程安全访问。
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

// GetClientsForTopic 获取所有与指定主题匹配的客户端ID。
// 这包括订阅了确切主题的客户端以及订阅了与主题匹配的通配符模式的客户端（例如，
// 订阅了"sensor/+"的客户端将接收到发送到"sensor/temperature"的消息）。
//
// 参数：
//   - topic: 要与现有订阅匹配的主题
//
// 返回值：
//   - 应该接收给定主题消息的客户端ID切片
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

// matchTopic 实现带有通配符的MQTT主题模式匹配。
// 支持两种类型的通配符：
// - '+' 匹配单个主题级别（例如 "sport/+/player" 匹配 "sport/tennis/player"）
// - '#' 匹配多个主题级别（例如 "sport/#" 匹配 "sport/tennis/stats/players"）
//
// 参数：
//   - pattern: 包含通配符的主题模式
//   - topic: 与模式匹配的实际主题
//
// 返回值：
//   - 如果主题与模式匹配则返回true，否则返回false
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
