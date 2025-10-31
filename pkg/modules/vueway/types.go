package vueway

import (
	"server-system/pkg/types"
	"time"
)

// ClientType - тип клиента
type ClientType int

const (
	ClientTypeFull ClientType = iota // Полный доступ
	ClientTypeDemo                   // Демо режим
)

// ClientConfig - конфигурация клиента
type ClientConfig struct {
	ID                  string
	Type                ClientType
	UserID              string
	ConnectedAt         time.Time
	LastDemoTime        time.Time
	DataWSConnection    WebSocketConnection
	ControlWSConnection WebSocketConnection
	TestWSConnection    WebSocketConnection
	MessageCount        int       // Счетчик сообщений для защиты от спама
	LastResetTime       time.Time // Время последнего сброса счетчика
}

// WebSocketConnection - интерфейс для WebSocket соединения
type WebSocketConnection interface {
	//Send(messageType string, data []byte) error
	Send(messageType string, message types.Message) error

	Close() error
	GetID() string
}

// ClientStats - статистика по клиентам
type ClientStats struct {
	TotalClients     int `json:"totalClients"`
	FullClients      int `json:"fullClients"`
	DemoClients      int `json:"demoClients"`
	MessagesSent     int `json:"messagesSent"`
	MessagesReceived int `json:"messagesReceived"`
}
