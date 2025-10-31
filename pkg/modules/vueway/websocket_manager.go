package vueway

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"server-system/pkg/types"
	"sync"
	"time"
)

type WebSocketManager struct {
	clientManager        *ClientManager
	dataWSConnections    sync.Map
	controlWSConnections sync.Map
	testWSConnections    sync.Map
	сhanOutputDbs        chan<- types.Message
	сhanInputDbs         <-chan types.Message
	config               WebSocketConfig
}

func NewWebSocketManager(clientManager *ClientManager, сhanOutputDbs chan<- types.Message, сhanInputDbs <-chan types.Message, config WebSocketConfig) *WebSocketManager {
	return &WebSocketManager{
		clientManager: clientManager,
		сhanOutputDbs: сhanOutputDbs,
		сhanInputDbs:  сhanInputDbs,
		config:        config,
	}
}

// HandleDataConnection обрабатывает подключение WebSocket для данных
func (wm *WebSocketManager) HandleDataConnection(clientID string, userID string, clientType ClientType, w http.ResponseWriter, r *http.Request) error {
	conn, err := NewRealWebSocketConnection(w, r, clientID+"_data")
	if err != nil {
		return err
	}

	wm.dataWSConnections.Store(clientID, conn)
	log.Printf("Data WebSocket connected: %s (user: %s, type: %v)", clientID, userID, clientType)
	return nil
}

// HandleControlConnection обрабатывает подключение WebSocket для команд
func (wm *WebSocketManager) HandleControlConnection(clientID string, userID string, clientType ClientType, w http.ResponseWriter, r *http.Request) error {
	// Регистрируем клиента
	config, err := wm.clientManager.RegisterClient(clientID, clientType, userID)
	if err != nil {
		return err
	}

	conn, err := NewRealWebSocketConnection(w, r, clientID+"_control")
	if err != nil {
		wm.clientManager.UnregisterClient(clientID)
		return err
	}

	config.ControlWSConnection = conn
	wm.controlWSConnections.Store(clientID, conn)

	// Запускаем слушатель сообщений с защитой от спама
	go conn.Listen(wm.сhanOutputDbs, wm.config.MaxMsgPerMinute, func() {
		wm.CloseConnection(clientID)
	})

	log.Printf("Control WebSocket connected: %s (user: %s, type: %v)", clientID, userID, clientType)
	return nil
}

// HandleTestConnection обрабатывает подключение тестового WebSocket
func (wm *WebSocketManager) HandleTestConnection(clientID string, userID string, w http.ResponseWriter, r *http.Request) error {
	conn, err := NewRealWebSocketConnection(w, r, clientID+"_test")
	if err != nil {
		return err
	}

	wm.testWSConnections.Store(clientID, conn)

	// Запускаем отправку тестовых данных
	go wm.startTestDataSender(clientID, conn)

	log.Printf("Test WebSocket connected: %s", clientID)
	return nil
}

// startTestDataSender отправляет тестовые данные каждые 10 секунд
func (wm *WebSocketManager) startTestDataSender(clientID string, conn WebSocketConnection) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		msg, _ := json.Marshal("Test data from server")
		data := types.Message{
			Type:     "test_data",
			Data:     msg,
			InitDT:   time.Now(),
			UpdateDT: time.Now(),
			Source:   "VUE_TEST",
		}

		if err := conn.Send("test_data", data); err != nil {
			log.Printf("Failed to send test data to %s: %v", clientID, err)
			return
		}
	}
}

// SendToClient отправляет сообщение клиенту через WebSocket данных
func (wm *WebSocketManager) SendToClient(clientID string, message types.Message) error {
	connInterface, exists := wm.dataWSConnections.Load(clientID)
	if !exists {
		return fmt.Errorf("client data connection not found: %s", clientID)
	}

	conn, ok := connInterface.(WebSocketConnection)
	if !ok {
		return fmt.Errorf("invalid data connection type for client: %s", clientID)
	}

	return conn.Send(message.Type, message)
}

// SendToAllClients отправляет сообщение всем подключенным клиентам
func (wm *WebSocketManager) SendToAllClients(message types.Message) {
	clientCount := 0
	wm.dataWSConnections.Range(func(key, value interface{}) bool {
		clientID := key.(string)
		if wm.clientManager.CheckClientValidity(clientID) {
			if err := wm.SendToClient(clientID, message); err != nil {
				log.Printf("Error sending to client %s: %v", clientID, err)
				wm.CloseConnection(clientID)
			} else {
				clientCount++
			}
		}
		return true
	})
}

// SendToClientByID отправляет сообщение конкретному клиенту
func (wm *WebSocketManager) SendToClientByID(clientID string, message types.Message) error {
	if clientID == "" {
		// Если ClientID не указан, отправляем всем
		wm.SendToAllClients(message)
		return nil
	}

	// Отправляем конкретному клиенту
	return wm.SendToClient(clientID, message)
}

// ProcessDatabaseUpdates обрабатывает обновления от database
func (wm *WebSocketManager) ProcessDatabaseUpdates(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-wm.сhanInputDbs:
			wm.handleDatabaseMessage(msg)
		}
	}
}

// handleDatabaseMessage обрабатывает сообщение от database
func (wm *WebSocketManager) handleDatabaseMessage(msg types.Message) {
	if msg.Type == "data_batch" {
		// Отправляем данные клиентам
		wm.SendToClientByID(msg.ClientID, msg)
	}
}

// StartWebSocketServers запускает WebSocket серверы
func (wm *WebSocketManager) StartWebSocketServers(ctx context.Context) {
	// WebSocket сервер для данных
	go wm.startServer(ctx, wm.config.DataPort, wm.handleDataWebSocket)

	// WebSocket сервер для команд
	go wm.startServer(ctx, wm.config.ControlPort, wm.handleControlWebSocket)

	// Тестовый WebSocket сервер
	go wm.startServer(ctx, wm.config.TestPort, wm.handleTestWebSocket)
}

func (wm *WebSocketManager) startServer(ctx context.Context, port int, handler func(http.ResponseWriter, *http.Request)) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", handler)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	go func() {
		log.Printf("WebSocket server starting on port %d", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("WebSocket server error: %v", err)
		}
	}()

	// Graceful shutdown
	<-ctx.Done()
	server.Shutdown(context.Background())
}

func (wm *WebSocketManager) handleDataWebSocket(w http.ResponseWriter, r *http.Request) {
	clientID := r.URL.Query().Get("clientId")
	userID := r.URL.Query().Get("userId")
	clientType := r.URL.Query().Get("type")

	if clientID == "" || userID == "" {
		http.Error(w, "Missing clientId or userId", http.StatusBadRequest)
		return
	}

	var cType ClientType
	if clientType == "demo" {
		cType = ClientTypeDemo
	} else {
		cType = ClientTypeFull
	}

	if err := wm.HandleDataConnection(clientID, userID, cType, w, r); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (wm *WebSocketManager) handleControlWebSocket(w http.ResponseWriter, r *http.Request) {
	clientID := r.URL.Query().Get("clientId")
	userID := r.URL.Query().Get("userId")
	clientType := r.URL.Query().Get("type")

	if clientID == "" || userID == "" {
		http.Error(w, "Missing clientId or userId", http.StatusBadRequest)
		return
	}

	var cType ClientType
	if clientType == "demo" {
		cType = ClientTypeDemo
	} else {
		cType = ClientTypeFull
	}

	if err := wm.HandleControlConnection(clientID, userID, cType, w, r); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (wm *WebSocketManager) handleTestWebSocket(w http.ResponseWriter, r *http.Request) {
	clientID := r.URL.Query().Get("clientId")
	userID := r.URL.Query().Get("userId")

	if clientID == "" {
		http.Error(w, "Missing clientId", http.StatusBadRequest)
		return
	}

	if err := wm.HandleTestConnection(clientID, userID, w, r); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// CloseConnection закрывает соединение клиента
func (wm *WebSocketManager) CloseConnection(clientID string) {
	// Закрываем все соединения клиента
	connections := []*sync.Map{&wm.dataWSConnections, &wm.controlWSConnections, &wm.testWSConnections}
	for _, connMap := range connections {
		if conn, exists := connMap.Load(clientID); exists {
			if wsConn, ok := conn.(WebSocketConnection); ok {
				wsConn.Close()
			}
			connMap.Delete(clientID)
		}
	}

	wm.clientManager.UnregisterClient(clientID)
	log.Printf("All connections closed for client: %s", clientID)
}
