package vueway

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

type ClientManager struct {
	clients sync.Map
	mu      sync.RWMutex
	config  DemoConfig

	// Атомарные счетчики статистики - ДОБАВЛЕНО
	totalClients atomic.Int64
	fullClients  atomic.Int64
	demoClients  atomic.Int64
}

func NewClientManager(config DemoConfig) *ClientManager {
	return &ClientManager{
		config: config,
	}
}

// RegisterClient регистрирует нового клиента
func (cm *ClientManager) RegisterClient(clientID string, clientType ClientType, userID string) (*ClientConfig, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Проверяем существующего клиента
	if existing, exists := cm.clients.Load(clientID); exists {
		config := existing.(*ClientConfig)
		return config, nil
	}

	// Для демо клиента проверяем время переподключения
	if clientType == ClientTypeDemo {
		if !cm.canReconnectDemo(userID) {
			return nil, fmt.Errorf("demo client can reconnect only after %d minutes", cm.config.ReconnectCooldownMin)
		}
	}

	config := &ClientConfig{
		ID:            clientID,
		Type:          clientType,
		UserID:        userID,
		ConnectedAt:   time.Now(),
		MessageCount:  0,          // ДОБАВЛЕНО
		LastResetTime: time.Now(), // ДОБАВЛЕНО
	}

	if clientType == ClientTypeDemo {
		config.LastDemoTime = time.Now()
	}

	cm.clients.Store(clientID, config)
	cm.updateStats(clientType, 1) // ИЗМЕНЕНО

	log.Printf("Client registered: %s, type: %v, user: %s", clientID, clientType, userID)
	return config, nil
}

// canReconnectDemo проверяет возможность переподключения демо клиента
func (cm *ClientManager) canReconnectDemo(userID string) bool {
	var lastDemoTime time.Time
	cm.clients.Range(func(key, value interface{}) bool {
		client := value.(*ClientConfig)
		if client.UserID == userID && client.Type == ClientTypeDemo {
			if client.LastDemoTime.After(lastDemoTime) {
				lastDemoTime = client.LastDemoTime
			}
		}
		return true
	})

	return time.Since(lastDemoTime) >= time.Duration(cm.config.ReconnectCooldownMin)*time.Minute
}

// UnregisterClient удаляет клиента
func (cm *ClientManager) UnregisterClient(clientID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if config, exists := cm.clients.Load(clientID); exists {
		client := config.(*ClientConfig)
		if client.Type == ClientTypeDemo {
			client.LastDemoTime = time.Now()
		}
		cm.clients.Delete(clientID)
		cm.updateStats(client.Type, -1) // ИЗМЕНЕНО
		log.Printf("Client unregistered: %s", clientID)
	}
}

// GetClient возвращает конфигурацию клиента
func (cm *ClientManager) GetClient(clientID string) (*ClientConfig, bool) {
	config, exists := cm.clients.Load(clientID)
	if !exists {
		return nil, false
	}
	return config.(*ClientConfig), true
}

// CheckClientValidity проверяет валидность клиента
func (cm *ClientManager) CheckClientValidity(clientID string) bool {
	config, exists := cm.GetClient(clientID)
	if !exists {
		return false
	}

	// Для демо клиента проверяем время сессии
	if config.Type == ClientTypeDemo {
		sessionDuration := time.Duration(cm.config.SessionDurationMin) * time.Minute
		if time.Since(config.ConnectedAt) > sessionDuration {
			cm.UnregisterClient(clientID)
			return false
		}
	}

	return true
}

// GetClientStats возвращает статистику
func (cm *ClientManager) GetClientStats() ClientStats {
	return ClientStats{
		TotalClients: int(cm.totalClients.Load()), // ИЗМЕНЕНО
		FullClients:  int(cm.fullClients.Load()),  // ИЗМЕНЕНО
		DemoClients:  int(cm.demoClients.Load()),  // ИЗМЕНЕНО
	}
}

// updateStats обновляет статистику атомарно - ДОБАВЛЕН НОВЫЙ МЕТОД
func (cm *ClientManager) updateStats(clientType ClientType, delta int) {
	cm.totalClients.Add(int64(delta))

	if clientType == ClientTypeFull {
		cm.fullClients.Add(int64(delta))
	} else {
		cm.demoClients.Add(int64(delta))
	}
}

// StartValidityChecker запускает проверку валидности клиентов
func (cm *ClientManager) StartValidityChecker(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cm.clients.Range(func(key, value interface{}) bool {
				clientID := key.(string)
				cm.CheckClientValidity(clientID)
				return true
			})
		}
	}
}
