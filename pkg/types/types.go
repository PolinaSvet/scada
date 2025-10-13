package types

import (
	"encoding/json"
	"time"
)

type MessageData struct {
	Message string `json:"message"`
	Time    string `json:"time"`
	Source  string `json:"source"`
}

// Константы для типов сообщений
const (
	MessageTypeError   = "ERROR"
	MessageTypeAlarm   = "ALARM"
	MessageTypeWarning = "WARNING"
	MessageTypeInfo    = "INFO"
	MessageTypeDebug   = "DEBUG"
	MessageTypeStatus  = "STATUS"
	MessageTypeMetric  = "METRIC"
)

// Message - основная структура сообщения
type Message struct {
	ID       string          `json:"id"`
	Type     string          `json:"type"`
	Data     json.RawMessage `json:"data"`
	InitDT   time.Time       `json:"init_dt"`
	UpdateDT time.Time       `json:"update_dt"`
	Source   string          `json:"source"`
}

// TagValue - значение тега
type TagValue struct {
	Tag       string      `json:"tag"`
	Alias     string      `json:"alias"`
	Value     interface{} `json:"value"`
	Quality   QualityCode `json:"quality"`
	Timestamp time.Time   `json:"timestamp"`
	DataType  DataType    `json:"data_type"`
}

type QualityCode int

const (
	QualityGood QualityCode = iota
	QualityBad
	QualityUncertain
)

type DataType string

const (
	DataTypeBOOL    DataType = "BOOL"
	DataTypeBYTE    DataType = "BYTE"
	DataTypeINT16   DataType = "INT16"
	DataTypeUINT16  DataType = "UINT16"
	DataTypeINT32   DataType = "INT32"
	DataTypeUINT32  DataType = "UINT32"
	DataTypeINT64   DataType = "INT64"
	DataTypeUINT64  DataType = "UINT64"
	DataTypeFLOAT32 DataType = "FLOAT32"
	DataTypeFLOAT64 DataType = "FLOAT64"
)

// ServiceStatus - статус сервиса
type ServiceStatus struct {
	ModuleID         string             `json:"module_id"`
	Status           string             `json:"status"`
	LastUpdate       time.Time          `json:"last_update"`
	MessagesSent     int                `json:"messages_sent"`
	MessagesRecv     int                `json:"messages_received"`
	ErrorsCount      int                `json:"errors_count"`
	StatusOutChannel StatusChannelAlive `json:"status_channel_alive"`
}

type StatusChannelAlive struct {
	ChannelAlive         bool    `json:"channel_alive"`
	ConsecutiveTimeouts  int     `json:"consecutive_timeouts"`
	TotalSent            int     `json:"total_sent"`
	TotalDropped         int     `json:"total_dropped"`
	SecondsSinceLastSend float64 `json:"seconds_since_last_send"`
}

// DatabaseTag - структура для хранения тега в базе данных
type DatabaseTag struct {
	Enable         bool        `json:"enable"`
	Tag            string      `json:"tag"`
	DataType       DataType    `json:"data_type"`
	Targets        string      `json:"targets"`
	Data           TagValue    `json:"data"`
	ValueOld       interface{} `json:"value_old"`
	TimeLastUpdate time.Time   `json:"time_last_update"`
	TimeLastChange time.Time   `json:"time_last_change"`
}

type TagsConfig struct {
	Tags map[string]DatabaseTag `json:"tags"`
}

type ObjectsConfigFile struct {
	Objects map[string]ObjectConfig `json:"objects"`
}
