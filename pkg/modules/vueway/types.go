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

// VueCommand - команда от Vue клиента
/*type VueCommand struct {
	ClientID string          `json:"clientId"`
	UserID   string          `json:"userId"`
	ObjectID string          `json:"objectId"`
	Command  string          `json:"command"`
	Data     json.RawMessage `json:"data"`
	Time     time.Time       `json:"time"`
}*/

type VueCommand struct {
	Command  string                 `json:"command" msgpack:"command"`
	ObjectID string                 `json:"objectId" msgpack:"objectId"`
	UserID   string                 `json:"userId" msgpack:"userId"`
	ClientID string                 `json:"clientId"`
	Data     map[string]interface{} `json:"data" msgpack:"data"`
	Time     string                 `json:"time" msgpack:"time"`
}

type PendingCommand struct {
	CmdTag          map[string]interface{} `json:"cmdTag" msgpack:"cmdTag"`
	CmdValue        int                    `json:"cmdValue" msgpack:"cmdValue"`
	CmdMess         string                 `json:"cmdMess" msgpack:"cmdMess"`
	CmdMessQuestion string                 `json:"cmdMessQuestion" msgpack:"cmdMessQuestion"`
	ObjId           string                 `json:"objId" msgpack:"objId"`
	ObjType         string                 `json:"objType" msgpack:"objType"`
}

// ClientStats - статистика по клиентам
type ClientStats struct {
	TotalClients     int `json:"totalClients"`
	FullClients      int `json:"fullClients"`
	DemoClients      int `json:"demoClients"`
	MessagesSent     int `json:"messagesSent"`
	MessagesReceived int `json:"messagesReceived"`
}
