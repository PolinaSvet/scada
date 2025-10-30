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

		//log.Println("XXXX", messageType, data)

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

func (rwc *RealWebSocketConnection) processMessage(message map[string]interface{}, messageChan chan<- VueCommand) {
	// Базовый парсинг VueCommand
	var vueCmd VueCommand

	// Преобразуем map в структуру (упрощенный способ)
	if cmd, ok := message["command"].(string); ok {
		vueCmd.Command = cmd
	}
	if objID, ok := message["objectId"].(string); ok {
		vueCmd.ObjectID = objID
	}
	if userID, ok := message["userId"].(string); ok {
		vueCmd.UserID = userID
	}
	if data, ok := message["data"].(map[string]interface{}); ok {
		vueCmd.Data = data
	}
	if timeStr, ok := message["time"].(string); ok {
		vueCmd.Time = timeStr
	}

	// Обрабатываем команду sendCommand
	if vueCmd.Command == "sendCommand" {
		rwc.processSendCommand(vueCmd)
	}

	// Отправляем в канал для дальнейшей обработки
	//messageChan <- vueCmd
	select {
	case messageChan <- vueCmd:
		log.Printf("Command received from %s: %s", rwc.id, vueCmd)
	default:
		log.Printf("Command channel full, dropping command from %s", rwc.id)
	}
}

func (rwc *RealWebSocketConnection) processSendCommand(vueCmd VueCommand) {
	// Логируем все полученные данные для отладки
	log.Printf("Raw VueCommand: %+v", vueCmd)

	data := vueCmd.Data

	// Прямое извлечение значений из data
	if cmdValue, ok := data["cmdValue"]; ok {
		var intCmdValue int64

		// Обрабатываем разные возможные типы числа
		switch v := cmdValue.(type) {
		case int64:
			intCmdValue = v
		case int:
			intCmdValue = int64(v)
		case float64:
			intCmdValue = int64(v)
		case uint16:
			intCmdValue = int64(v)
		default:
			log.Printf("Unknown type for cmdValue: %T", cmdValue)
			return
		}

		codeCmd := (int(intCmdValue) >> 12) & 0xF
		idObj := int(intCmdValue) & 0xFFF

		objId, _ := data["objId"].(string)
		objType, _ := data["objType"].(string)
		cmdMess, _ := data["cmdMess"].(string)
		cmdMessQuestion, _ := data["cmdMessQuestion"].(string)
		cmdTagData := rwc.parseCmdTag(data["cmdTag"])

		log.Printf("Parsed command: codeCmd=%d, idObj=%d, objId=%s, objType=%s",
			codeCmd, idObj, objId, objType)
		log.Printf("Message: %s, Question: %s", cmdMess, cmdMessQuestion)
		log.Printf("Command tag: %+v", cmdTagData)

	} else {
		log.Printf("cmdValue not found in data")
	}
}

// Функция для парсинга cmdTag
func (rwc *RealWebSocketConnection) parseCmdTag(cmdTag interface{}) map[string]string {
	result := make(map[string]string)

	if cmdTag == nil {
		return result
	}

	switch tag := cmdTag.(type) {

	case map[string]interface{}:
		for key, value := range tag {
			if strValue, ok := value.(string); ok {
				result[key] = strValue
			} else {
				result[key] = fmt.Sprintf("%v", value)
			}
		}

	default:
		log.Printf("Unknown cmdTag type: %T", cmdTag)
	}

	return result
}
