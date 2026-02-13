package connection

import (
	"context"
	"sync"
	"time"
)

// ConnectionManager handles connection pooling and resource management for high-scale MQTT
type ConnectionManager struct {
	connections    sync.Map
	maxConnections int
	currentCount   int64
	countMu        sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
	metrics        *Metrics
}

// Metrics holds connection metrics for monitoring
type Metrics struct {
	ConnectionsAccepted int64
	ConnectionsClosed   int64
	MessagesProcessed   int64
	BytesTransferred    int64
}

// NewConnectionManager creates a new connection manager with specified max connections
func NewConnectionManager(maxConns int) *ConnectionManager {
	ctx, cancel := context.WithCancel(context.Background())

	manager := &ConnectionManager{
		maxConnections: maxConns,
		ctx:            ctx,
		cancel:         cancel,
		metrics:        &Metrics{},
	}

	return manager
}

func (cm *ConnectionManager) RegisterConnection(clientID string) error {
	cm.countMu.Lock()
	currentCount := cm.currentCount

	if currentCount >= int64(cm.maxConnections) {
		cm.countMu.Unlock()
		return ErrMaxConnectionsReached
	}

	cm.currentCount++
	cm.countMu.Unlock()

	cm.connections.Store(clientID, true)

	cm.metrics.ConnectionsAccepted++

	return nil
}

func (cm *ConnectionManager) DeregisterConnection(clientID string) {
	cm.connections.Delete(clientID)

	cm.countMu.Lock()
	cm.currentCount--
	cm.countMu.Unlock()

	cm.metrics.ConnectionsClosed++
}

func (cm *ConnectionManager) ConnectionExists(clientID string) bool {
	_, exists := cm.connections.Load(clientID)
	return exists
}

func (cm *ConnectionManager) GetActiveConnections() int64 {
	cm.countMu.RLock()
	defer cm.countMu.RUnlock()
	return cm.currentCount
}

func (cm *ConnectionManager) GetMaxConnections() int {
	return cm.maxConnections
}

func (cm *ConnectionManager) GetMetrics() Metrics {
	return *cm.metrics
}

func (cm *ConnectionManager) Close() {
	cm.cancel()
}

var ErrMaxConnectionsReached = &ConnectionError{"maximum connections reached"}

type ConnectionError struct {
	msg string
}

func (e *ConnectionError) Error() string {
	return e.msg
}

func (cm *ConnectionManager) Monitor(reportInterval time.Duration) {
	ticker := time.NewTicker(reportInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Log metrics periodically
			// active := cm.GetActiveConnections()
			// max := cm.GetMaxConnections()
			// totalAccepted := cm.metrics.ConnectionsAccepted
			// totalClosed := cm.metrics.ConnectionsClosed
			// msgProcessed := cm.metrics.MessagesProcessed

		case <-cm.ctx.Done():
			return
		}
	}
}

// IncrementMessageCounter increments the message counter
func (cm *ConnectionManager) IncrementMessageCounter() {
	cm.metrics.MessagesProcessed++
}

// AddToByteTransferCounter adds bytes transferred to the counter
func (cm *ConnectionManager) AddToByteTransferCounter(bytes int64) {
	cm.metrics.BytesTransferred += bytes
}
