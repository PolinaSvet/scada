package vueway

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"server-system/pkg/types"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/vmihailenco/msgpack/v5"
)

// RealWebSocketConnection - реальная реализация WebSocket соединения
type RealWebSocketConnection struct {
	conn     *websocket.Conn
	id       string
	mu       sync.RWMutex
	isClosed bool
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // В продакшене нужно настроить properly
	},
}

// NewRealWebSocketConnection создает новое WebSocket соединение
func NewRealWebSocketConnection(w http.ResponseWriter, r *http.Request, id string) (*RealWebSocketConnection, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}

	return &RealWebSocketConnection{
		conn: conn,
		id:   id,
	}, nil
}

// Send отправляет сообщение через WebSocket с MessagePack
// func (rwc *RealWebSocketConnection) Send(messageType string, data []byte) error {
func (rwc *RealWebSocketConnection) Send(messageType string, message types.Message) error {

	rwc.mu.RLock()
	defer rwc.mu.RUnlock()

	if rwc.isClosed {
		return fmt.Errorf("websocket connection closed")
	}

	var decodeData interface{}
	_ = json.Unmarshal(message.Data, &decodeData)

	// Создаем структуру сообщения
	msgWrapper := map[string]interface{}{
		"type": messageType,
		"data": decodeData,
		"time": time.Now(),
	}

	// Упаковываем в MessagePack
	msgpackData, err := msgpack.Marshal(msgWrapper)
	if err != nil {
		return fmt.Errorf("failed to marshal message with msgpack: %v", err)
	}

	return rwc.conn.WriteMessage(websocket.BinaryMessage, msgpackData)
}

// SendJSON отправляет JSON сообщение (для обратной совместимости)
func (rwc *RealWebSocketConnection) SendJSON(messageType string, data []byte) error {
	rwc.mu.RLock()
	defer rwc.mu.RUnlock()

	if rwc.isClosed {
		return fmt.Errorf("websocket connection closed")
	}

	message := map[string]interface{}{
		"type": messageType,
		"data": data,
		"time": time.Now(),
	}

	return rwc.conn.WriteJSON(message)
}

// Close закрывает соединение
func (rwc *RealWebSocketConnection) Close() error {
	rwc.mu.Lock()
	defer rwc.mu.Unlock()

	if rwc.isClosed {
		return nil
	}

	rwc.isClosed = true
	return rwc.conn.Close()
}

// GetID возвращает ID соединения
func (rwc *RealWebSocketConnection) GetID() string {
	return rwc.id
}

// Listen слушает входящие сообщения с защитой от спама
func (rwc *RealWebSocketConnection) Listen(messageChan chan<- VueCommand, maxMsgPerMinute int, onSpamDetected func()) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("WebSocket listen panic: %v", r)
		}
	}()

	messageCount := 0
	lastReset := time.Now()

	for {
		rwc.mu.RLock()
		if rwc.isClosed {
			rwc.mu.RUnlock()
			return
		}
		rwc.mu.RUnlock()

		// Читаем сообщение
		messageType, data, err := rwc.conn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			rwc.Close()
			return
		}

		// Проверяем защиту от спама
		if time.Since(lastReset) > time.Minute {
			messageCount = 0
			lastReset = time.Now()
		}
		messageCount++

		if messageCount > maxMsgPerMinute {
			log.Printf("Spam detected from client %s, closing connection", rwc.id)
			if onSpamDetected != nil {
				onSpamDetected()
			}
			rwc.Close()
			return
		}

		// Обрабатываем сообщение
		if messageType == websocket.BinaryMessage {
			// MessagePack сообщение
			var message map[string]interface{}
			if err := msgpack.Unmarshal(data, &message); err != nil {
				log.Printf("Failed to unmarshal msgpack message: %v", err)
				continue
			}
			rwc.processMessage(message, messageChan)
		} else if messageType == websocket.TextMessage {
			// JSON сообщение (для обратной совместимости)
			var message map[string]interface{}
			if err := rwc.conn.ReadJSON(&message); err != nil {
				log.Printf("Failed to unmarshal JSON message: %v", err)
				continue
			}
			rwc.processMessage(message, messageChan)
		}
	}
}

// processMessage обрабатывает распакованное сообщение
func (rwc *RealWebSocketConnection) processMessage(message map[string]interface{}, messageChan chan<- VueCommand) {
	if command, ok := message["command"]; ok {
		vueCommand := VueCommand{
			ClientID: rwc.id,
			UserID:   getString(message, "userId"),
			ObjectID: getString(message, "objectId"),
			Command:  command.(string),
			Time:     time.Now(),
		}

		// Сериализуем data если есть
		if data, exists := message["data"]; exists {
			dataBytes, err := msgpack.Marshal(data)
			if err == nil {
				vueCommand.Data = dataBytes
			}
		}

		select {
		case messageChan <- vueCommand:
			log.Printf("Command received from %s: %s", rwc.id, command)
		default:
			log.Printf("Command channel full, dropping command from %s", rwc.id)
		}
	}
}

func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}
