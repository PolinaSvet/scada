package vueway

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"server-system/pkg/types"
)

type VueWayConfig struct {
	ID           string            `json:"id"`
	LimitTimeMs  int               `json:"limit_time_ms"`
	StatusTimeS  int               `json:"status_time_s"`
	MaxClients   int               `json:"max_clients"`
	BatchWriting types.BatchConfig `json:"batch_writing"`
	WebSocket    WebSocketConfig   `json:"websocket"`
	DemoMode     DemoConfig        `json:"demo_mode"`
}

type WebSocketConfig struct {
	DataPort        int `json:"data_port"`
	ControlPort     int `json:"control_port"`
	TestPort        int `json:"test_port"`
	ReadBufferSize  int `json:"read_buffer_size"`
	WriteBufferSize int `json:"write_buffer_size"`
	MaxMsgPerMinute int `json:"max_msg_per_minute"` // Защита от спама
}

type DemoConfig struct {
	SessionDurationMin   int `json:"session_duration_min"`
	ReconnectCooldownMin int `json:"reconnect_cooldown_min"`
}

// основная конфигурация модуля
func (vw *VueWay) loadMainConfig(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("не удалось прочитать файл: %w", err)
	}

	var config VueWayConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		return fmt.Errorf("ошибка парсинга JSON: %w", err)
	}
	vw.config = config

	if err := vw.validateMainConfig(); err != nil {
		return fmt.Errorf("ошибка валидации конфига: %w", err)
	}

	return nil
}

func (vw *VueWay) validateMainConfig() error {
	if vw.config.WebSocket.DataPort == 0 {
		return errors.New("поле 'DataPort' обязательно")
	}
	if vw.config.WebSocket.ControlPort == 0 {
		return errors.New("поле 'ControlPort' обязательно")
	}
	if vw.config.WebSocket.TestPort == 0 {
		return errors.New("поле 'TestPort' обязательно")
	}
	if vw.config.WebSocket.MaxMsgPerMinute == 0 {
		vw.config.WebSocket.MaxMsgPerMinute = 60 // По умолчанию 60 сообщений в минуту
	}
	return nil
}
