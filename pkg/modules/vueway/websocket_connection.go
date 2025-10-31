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

// Listen слушает входящие сообщения
func (rwc *RealWebSocketConnection) Listen(messageChan chan<- types.Message, maxMsgPerMinute int, onSpamDetected func()) {
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
			rwc.decodeMessagePack(data, messageChan)
		} else {
			log.Printf("Failed type message: %v", messageType)
		}
	}
}

// Основной метод декодирования MessagePack
func (rwc *RealWebSocketConnection) decodeMessagePack(data []byte, messageChan chan<- types.Message) {
	// Декодируем как generic map
	var rawMap map[string]interface{}
	if err := msgpack.Unmarshal(data, &rawMap); err != nil {
		log.Printf("Failed to unmarshal msgpack message: %v", err)
		return
	}

	// Преобразуем map в структуру Message
	message := rwc.mapToMessage(rawMap)

	// Обрабатываем только команды
	if message.Type == "command" {

		// Отправляем в канал для дальнейшей обработки
		select {
		case messageChan <- message:
			log.Printf("Command sent to channel: %s", message.Type)
		default:
			log.Printf("Command channel full, dropping command from %s", rwc.id)
		}
	} else {
		log.Printf("Unknown message type: %s", message.Type)
	}

	//rwc.processMessage(message, messageChan)
}

// Преобразование map в структуру Message
func (rwc *RealWebSocketConnection) mapToMessage(rawMap map[string]interface{}) types.Message {
	var message types.Message

	if id, ok := rawMap["id"].(string); ok {
		message.ID = id
	}
	if msgType, ok := rawMap["type"].(string); ok {
		message.Type = msgType
	}
	if source, ok := rawMap["source"].(string); ok {
		message.Source = source
	}
	if clientID, ok := rawMap["clientId"].(string); ok {
		message.ClientID = clientID
	}

	// Обрабатываем data поле - конвертируем в json.RawMessage
	if dataField, exists := rawMap["data"]; exists {
		if dataBytes, err := json.Marshal(dataField); err == nil {
			message.Data = json.RawMessage(dataBytes)
		}
	}

	// Обрабатываем временные метки
	if initDT, ok := rawMap["init_dt"].(string); ok {
		if t, err := time.Parse(time.RFC3339, initDT); err == nil {
			message.InitDT = t
		}
	}
	if updateDT, ok := rawMap["update_dt"].(string); ok {
		if t, err := time.Parse(time.RFC3339, updateDT); err == nil {
			message.UpdateDT = t
		}
	}

	return message
}
