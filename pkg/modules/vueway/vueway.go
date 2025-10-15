package vueway

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"server-system/pkg/batch"
	"server-system/pkg/types"
	"sync/atomic"
	"time"
)

type VueWayInit struct {
	Ctx            context.Context
	ChanSystemMess chan<- types.Message
	ChanStatus     chan<- types.Message
	ChanOutputDbs  chan<- types.Message
	ChanInputDbs   <-chan types.Message
	ConfigFile     string
}

type VueWay struct {
	ctx        context.Context
	сonfigFile string
	config     VueWayConfig

	// каналы
	chanSystemMess chan<- types.Message
	chanStatus     chan<- types.Message
	сhanOutputDbs  chan<- types.Message
	сhanInputDbs   <-chan types.Message

	// пакетная обработка данных
	batchProcessor *batch.BatchProcessor

	// статистика
	cntMsgGet atomic.Int64
	cntMsgSet atomic.Int64
	cntErr    atomic.Int64
	startTime time.Time

	clientManager    *ClientManager
	websocketManager *WebSocketManager
}

func NewModule(init VueWayInit) *VueWay {

	vw := &VueWay{ctx: init.Ctx,
		chanSystemMess: init.ChanSystemMess,
		chanStatus:     init.ChanStatus,
		сhanOutputDbs:  init.ChanOutputDbs,
		сhanInputDbs:   init.ChanInputDbs,
		сonfigFile:     init.ConfigFile,
		startTime:      time.Now(), // ДОБАВЛЕНО
	}

	return vw
}

func (vw *VueWay) Start() {
	defer func() {
		if r := recover(); r != nil {
			vw.sendMessError("VueWay panic: %v", r)
		}
	}()

	// Загружаем конфиг
	if err := vw.loadMainConfig(vw.сonfigFile); err != nil {
		vw.sendMessError("Failed to load config: %v", err)
		return
	}

	// Инициализируем clientManager
	clientManager := NewClientManager(vw.config.DemoMode, vw.config.MaxClients)
	vw.clientManager = clientManager

	// Инициализируем websocketManager
	vw.websocketManager = NewWebSocketManager(
		clientManager,
		vw.сhanOutputDbs,
		vw.сhanInputDbs,
		vw.config.WebSocket,
	)

	// Инициализируем пакетный процессор
	vw.batchProcessor = batch.NewBatchProcessor(
		vw.сhanOutputDbs,
		vw.chanSystemMess,
		vw.config.BatchWriting,
		vw.config.ID,
	)

	// Запускаем отправку статусов
	go vw.processStatus()

	// Запускаем менеджер клиентов
	go vw.clientManager.StartValidityChecker(vw.ctx)

	// Запускаем WebSocket серверы
	go vw.websocketManager.StartWebSocketServers(vw.ctx)

	// Запускаем обработку обновлений от database
	go vw.websocketManager.ProcessDatabaseUpdates(vw.ctx)

	// Запускаем обработку команд от клиентов
	go vw.processCommands()

	log.Printf("VueWay started, max clients: %d", vw.config.MaxClients)
	vw.sendMessStatus("<%v> module started", vw.config.ID)
}

// processCommands обрабатывает команды от клиентов
func (vw *VueWay) processCommands() {
	for {
		select {
		case <-vw.ctx.Done():
			return
		case command := <-vw.websocketManager.GetCommandChan():
			vw.handleCommand(command)
		}
	}
}

// handleCommand обрабатывает отдельную команду
func (vw *VueWay) handleCommand(command VueCommand) {
	defer func() {
		if r := recover(); r != nil {
			vw.sendMessError("handleCommand panic: %v", r)
		}
	}()

	vw.cntMsgGet.Add(1)

	// Отправляем команду в database
	if err := vw.websocketManager.SendCommandToDatabase(command); err != nil {
		vw.sendMessError("Failed to send command to database: %v", err)
		return
	}

	log.Printf("Command processed - Client: %s, User: %s, Object: %s, Command: %s",
		command.ClientID, command.UserID, command.ObjectID, command.Command)
}

// === STATUS ==========================================================

func (vw *VueWay) processStatus() {
	ticker := time.NewTicker(time.Duration(vw.config.StatusTimeS) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-vw.ctx.Done():
			return
		case <-ticker.C:
			vw.sendStatus()
		}
	}
}

func (vw *VueWay) sendStatus() {
	msgGet := vw.cntMsgGet.Load()
	msgSet := vw.cntMsgSet.Load()
	errCount := vw.cntErr.Load()

	clientStats := vw.clientManager.GetClientStats()

	status := types.ServiceStatus{
		ModuleID:         vw.config.ID,
		Status:           "running",
		LastUpdate:       time.Now(),
		MessagesSent:     int(msgSet),
		MessagesRecv:     int(msgGet),
		ErrorsCount:      int(errCount),
		StatusOutChannel: vw.batchProcessor.GetChannelStats(),
		ExtraData: map[string]interface{}{
			"total_clients": clientStats.TotalClients,
			"full_clients":  clientStats.FullClients,
			"demo_clients":  clientStats.DemoClients,
			"uptime":        time.Since(vw.startTime).String(),
			"max_clients":   vw.config.MaxClients,
		},
	}

	if vw.chanStatus != nil {
		data, _ := json.Marshal(status)
		msg := types.Message{
			Type:   "status",
			Data:   data,
			Source: vw.config.ID,
		}

		select {
		case vw.chanStatus <- msg:
			vw.cntMsgGet.Store(0)
			vw.cntMsgSet.Store(0)
			vw.cntErr.Store(0)
		default:
			log.Printf("Error channel full")
		}
	}
}

// === MESSAGE ==========================================================

func (vw *VueWay) sendMessage(msgType string, format string, args ...interface{}) {
	content := fmt.Sprintf(format, args...)

	// Если канал не nil, отправляем сообщение
	if vw.chanSystemMess != nil {
		messageData := types.MessageData{
			Message: content,
			Time:    time.Now().Format(time.RFC3339),
			Source:  vw.config.ID,
		}

		data, err := json.Marshal(messageData)
		if err != nil {
			log.Printf("ERROR: Failed to marshal message: %v", err)
			return
		}

		msg := types.Message{
			Type:     msgType,
			Data:     data,
			InitDT:   time.Now(),
			UpdateDT: time.Now(),
			Source:   vw.config.ID,
		}

		// Неблокирующая отправка
		select {
		case vw.chanSystemMess <- msg:
			// Сообщение отправлено
		case <-time.After(100 * time.Millisecond):
			log.Printf("WARNING: Message channel timeout for type: %s", msgType)
		case <-vw.ctx.Done():
			// Контекст отменен
		}
	}
}

// Специализированные методы для разных типов сообщений
func (vw *VueWay) sendMessError(format string, args ...interface{}) {
	vw.sendMessage(types.MessageTypeError, format, args...)
}

func (vw *VueWay) sendMessAlarm(format string, args ...interface{}) {
	vw.sendMessage(types.MessageTypeAlarm, format, args...)
}

func (vw *VueWay) sendMessWarning(format string, args ...interface{}) {
	vw.sendMessage(types.MessageTypeWarning, format, args...)
}

func (vw *VueWay) sendMessInfo(format string, args ...interface{}) {
	vw.sendMessage(types.MessageTypeInfo, format, args...)
}

func (vw *VueWay) sendMessDebug(format string, args ...interface{}) {
	vw.sendMessage(types.MessageTypeDebug, format, args...)
}

func (vw *VueWay) sendMessStatus(format string, args ...interface{}) {
	vw.sendMessage(types.MessageTypeStatus, format, args...)
}
