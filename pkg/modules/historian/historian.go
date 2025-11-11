package historian

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"server-system/pkg/types"
	"sync/atomic"
	"time"
)

// HistorianInit - параметры инициализации
type HistorianInit struct {
	Ctx            context.Context
	ChanSystemMess chan<- types.Message
	ChanStatus     chan<- types.Message
	ChanInputDbs   <-chan types.Message
	ChanOutputVue  chan<- types.Message
	ConfigFile     string
}

type Historian struct {
	ctx        context.Context
	сonfigFile string
	config     HistorianConfig

	// каналы
	chanSystemMess chan<- types.Message
	chanStatus     chan<- types.Message
	chanInputDbs   <-chan types.Message
	chanOutputVue  chan<- types.Message

	// обработчики
	alarmProcessor *AlarmProcessor

	// статистика
	cntMsgGet atomic.Int64
	cntMsgSet atomic.Int64
	cntErr    atomic.Int64
	startTime time.Time
}

func NewModule(init HistorianInit) *Historian {
	hist := &Historian{
		ctx:            init.Ctx,
		chanSystemMess: init.ChanSystemMess,
		chanStatus:     init.ChanStatus,
		chanInputDbs:   init.ChanInputDbs,
		chanOutputVue:  init.ChanOutputVue,
		сonfigFile:     init.ConfigFile,
		startTime:      time.Now(),
	}

	return hist
}

func (hist *Historian) Start() {
	defer func() {
		if r := recover(); r != nil {
			hist.sendMessError("historian panic: %v", r)
		}
	}()

	// Загружаем конфиг
	if err := hist.loadConfig(hist.сonfigFile); err != nil {
		hist.sendMessError("failed to load config: %v", err)
		return
	}

	// Инициализируем обработчики
	if err := hist.initializeProcessors(); err != nil {
		hist.sendMessError("failed to initialize processors: %v", err)
		return
	}

	// Запускаем обработчики в отдельных горутинах
	go hist.processStatus()
	go hist.processMessages()

	hist.sendMessStatus("<%v> module started", hist.config.ID)
}

// initializeProcessors инициализирует обработчики данных
func (hist *Historian) initializeProcessors() error {
	// Инициализируем обработчик алармов
	if hist.config.Alarm.Enable {
		processor, err := NewAlarmProcessor(&hist.config.Alarm)
		if err != nil {
			return fmt.Errorf("failed to initialize alarm processor: %w", err)
		}
		hist.alarmProcessor = processor
		log.Printf("alarm processor initialized")
	}

	defer hist.Stop()

	// TODO: Инициализировать обработчик трендов

	return nil
}

// 1.1 получаем упакованные данные, отправляем на распаковку в отдельной рутине -> processMessage
func (hist *Historian) processMessages() {
	defer func() {
		if r := recover(); r != nil {
			hist.sendMessError("processMessages panic: %v", r)
		}
	}()

	for {
		select {
		case <-hist.ctx.Done():
			return
		case msg, ok := <-hist.chanInputDbs:
			if !ok {
				hist.sendMessError("historian: chanInputDbs channel closed")
				return
			}

			hist.cntMsgGet.Add(1)

			go func(m types.Message) {
				taskCtx, cancel := context.WithTimeout(hist.ctx, time.Duration(hist.config.LimitTimeMs)*time.Millisecond)
				defer cancel()

				hist.processMessage(taskCtx, m)
			}(msg)
		}
	}
}

// 1.2. распаковываем данные, отправляем каждый тэг на обработку -> processTagValue
func (hist *Historian) processMessage(ctx context.Context, msg types.Message) {
	defer func() {
		if r := recover(); r != nil {
			hist.cntErr.Add(1)
			hist.sendMessError("processMessage panic: %v", r)
		}
	}()

	switch msg.Type {
	case "alarms_batch":
		log.Println("alarms_batch received: ", msg.UpdateDT)
		if hist.alarmProcessor != nil {
			if err := hist.alarmProcessor.ProcessBatch(ctx, msg.Data); err != nil {
				hist.sendMessError("failed to process alarm batch: %v", err)
			}
		}

	case "alarms_get_data":
		log.Println("alarms_get_data received: ", msg.UpdateDT)
		if hist.alarmProcessor != nil {
			if err := hist.alarmProcessor.ProcessGetData(ctx, msg.Data, hist.chanOutputVue, hist.config.ID); err != nil {
				hist.sendMessError("failed to process alarm get data: %v", err)
			}
		}

	case "trends_batch":
		log.Println("trends_batch: ", msg.UpdateDT)
		// TODO: добавить функцию сохранения данных в бд

	case "trends_get_data":
		log.Println("trends_get_data: ", msg.UpdateDT)
		// TODO: запрос к бд на извлечение данных
	}
}

// Stop останавливает модуль
func (hist *Historian) Stop() {
	if hist.alarmProcessor != nil {
		hist.alarmProcessor.Close()
	}
}

// === STATUS ==========================================================

func (hist *Historian) processStatus() {
	ticker := time.NewTicker(time.Duration(hist.config.StatusTimeS) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-hist.ctx.Done():
			return
		case <-ticker.C:
			hist.sendStatus()
		}
	}
}

func (hist *Historian) sendStatus() {
	msgGet := hist.cntMsgGet.Load()
	msgSet := hist.cntMsgSet.Load()
	errCount := hist.cntErr.Load()

	status := types.ServiceStatus{
		ModuleID:     hist.config.ID,
		Status:       "running",
		LastUpdate:   time.Now(),
		MessagesSent: int(msgSet),
		MessagesRecv: int(msgGet),
		ErrorsCount:  int(errCount),
	}

	if hist.chanStatus != nil {
		data, _ := json.Marshal(status)
		msg := types.Message{
			Type:   "status",
			Data:   data,
			Source: hist.config.ID,
		}

		select {
		case hist.chanStatus <- msg:
			hist.cntMsgGet.Store(0)
			hist.cntMsgSet.Store(0)
			hist.cntErr.Store(0)
		default:
			log.Printf("Error channel full")
		}
	}
}

// === MESSAGE ==========================================================

func (hist *Historian) sendMessage(msgType string, format string, args ...interface{}) {
	content := fmt.Sprintf(format, args...)

	if hist.chanSystemMess != nil {
		messageData := types.MessageData{
			Message: content,
			Time:    time.Now().Format(time.RFC3339),
			Source:  hist.config.ID,
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
			Source:   hist.config.ID,
		}

		select {
		case hist.chanSystemMess <- msg:
		case <-time.After(100 * time.Millisecond):
			log.Printf("WARNING: Message channel timeout for type: %s", msgType)
		case <-hist.ctx.Done():
		}
	}
}

func (hist *Historian) sendMessError(format string, args ...interface{}) {
	hist.sendMessage(types.MessageTypeError, format, args...)
}

func (hist *Historian) sendMessAlarm(format string, args ...interface{}) {
	hist.sendMessage(types.MessageTypeAlarm, format, args...)
}

func (hist *Historian) sendMessWarning(format string, args ...interface{}) {
	hist.sendMessage(types.MessageTypeWarning, format, args...)
}

func (hist *Historian) sendMessInfo(format string, args ...interface{}) {
	hist.sendMessage(types.MessageTypeInfo, format, args...)
}

func (hist *Historian) sendMessDebug(format string, args ...interface{}) {
	hist.sendMessage(types.MessageTypeDebug, format, args...)
}

func (hist *Historian) sendMessStatus(format string, args ...interface{}) {
	hist.sendMessage(types.MessageTypeStatus, format, args...)
}
