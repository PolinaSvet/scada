package database

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"server-system/pkg/batch"
	"server-system/pkg/objects"
	"server-system/pkg/types"
	"sync"
	"time"
)

// DatabaseInit - параметры инициализации
type DatabaseInit struct {
	Ctx            context.Context
	ChanSystemMess chan<- types.Message
	ChanStatus     chan<- types.Message
	ChanInputGen   <-chan types.Message
	ChanOutputVue  chan<- types.Message
	ChanInputVue   <-chan types.Message
	ChanOutputDbsA chan<- types.Message
	ChanOutputDbsT chan<- types.Message
	ConfigFile     string
}

type Database struct {
	ctx        context.Context
	сonfigFile string
	config     types.DatabaseMainConfig

	// каналы
	chanSystemMess chan<- types.Message
	chanStatus     chan<- types.Message
	chanInputGen   <-chan types.Message
	chanOutputVue  chan<- types.Message
	chanInputVue   <-chan types.Message
	chanOutputDbsA chan<- types.Message
	chanOutputDbsT chan<- types.Message

	// пакетная обработка данных
	batchProcessor      *batch.BatchProcessor
	batchProcessorMess  *batch.BatchProcessor
	batchProcessorHistA *batch.BatchProcessor
	batchProcessorHistT *batch.BatchProcessor

	// статистика
	cntMsgGet int
	cntMsgSet int
	cntErr    int
	statsMu   sync.Mutex

	// база тэгов - используем sync.Map для безопасного доступа из горутин
	dbTags sync.Map // ключ: string (alias), значение: types.DatabaseTag

	// Быстрый доступ по алиасам - sync.Map
	aliasIndex sync.Map // ключ: string (alias), значение: []types.ObjectReference

	// база тэгов - используем sync.Map для безопасного доступа из горутин
	dbTrend sync.Map // ключ: string (alias), значение: types.TrendTagInfo

	// Хранилища конфигов
	objSensorsConfig sync.Map // ключ: string, значение: types.ObjectConfig
	objDiConfig      sync.Map // ключ: string, значение: types.ObjectConfig

	// Хранилища состояний (конкретные типы)
	objSensorsState sync.Map // ключ: string, значение: *types.VueObjectSensorsState
	objDiState      sync.Map // ключ: string, значение: *types.VueObjectDiState

	objectsManager *ObjectsManager
}

func NewModule(init DatabaseInit) *Database {

	db := &Database{ctx: init.Ctx,
		chanSystemMess: init.ChanSystemMess,
		chanStatus:     init.ChanStatus,
		chanInputGen:   init.ChanInputGen,
		chanOutputVue:  init.ChanOutputVue,
		chanInputVue:   init.ChanInputVue,
		chanOutputDbsA: init.ChanOutputDbsA,
		chanOutputDbsT: init.ChanOutputDbsT,
		сonfigFile:     init.ConfigFile,
	}

	// Инициализируем менеджер объектов
	db.objectsManager = NewObjectsManager()

	// Регистрируем хранилища для каждого типа
	db.objectsManager.RegisterStorage(
		string(objects.TypeSensor),
		&db.objSensorsConfig,
		&db.objSensorsState,
		func() interface{} { return &objects.VueObjectSensorsState{} },
	)
	db.objectsManager.RegisterStorage(
		string(objects.TypeDi),
		&db.objDiConfig,
		&db.objDiState,
		func() interface{} { return &objects.VueObjectDiState{} },
	)

	return db
}

func (db *Database) Start() {

	defer func() {
		if r := recover(); r != nil {
			db.sendMessError("Database panic: %v", r)
		}
	}()

	// Загружаем конфиг
	if err := db.loadMainConfig(db.сonfigFile); err != nil {
		db.sendMessError("Failed to load config: %v", err)
		return
	}

	// Загружаем тэги
	if err := db.loadTagsConfig(db.config.DatabaseConfigPath); err != nil {
		db.sendMessError("Failed to load database data: %v", err)
		return
	}

	// Загружаем trend
	if err := db.loadTrendConfig(db.config.TrendConfig.ConfigPath); err != nil {
		db.sendMessError("Failed to load trend data: %v", err)
		return
	}

	// Загружаем объекты
	if err := db.loadObjectsConfig(); err != nil {
		db.sendMessError("Failed to load objects data: %v", err)
		return
	}

	// Инициализируем пакетный процессор
	db.batchProcessor = batch.NewBatchProcessor(
		db.chanOutputVue,
		db.chanSystemMess,
		db.config.BatchWriting,
		db.config.ID,
		"data_batch",
	)

	db.batchProcessorMess = batch.NewBatchProcessor(
		db.chanOutputVue,
		db.chanSystemMess,
		db.config.BatchWriting,
		db.config.ID,
		"mess_batch",
	)

	db.batchProcessorHistA = batch.NewBatchProcessor(
		db.chanOutputDbsA,
		db.chanSystemMess,
		db.config.BatchWriting,
		db.config.ID,
		"alarms_batch",
	)

	db.batchProcessorHistT = batch.NewBatchProcessor(
		db.chanOutputDbsT,
		db.chanSystemMess,
		db.config.BatchWriting,
		db.config.ID,
		"trends_batch",
	)

	// Запускаем обработчики в отдельных горутинах
	go db.processStatus()
	go db.processMessages()
	go db.batchProcessor.Start(db.ctx)
	go db.batchProcessorMess.Start(db.ctx)
	go db.batchProcessorHistA.Start(db.ctx)
	go db.batchProcessorHistT.Start(db.ctx)
	go db.trendTicker()

	// Запускаем обработчик трендов если включено
	/*if db.config.TrendConfig.Enable && (db.config.TrendConfig.SaveType == 1 || db.config.TrendConfig.SaveType == 2) {
		go db.trendTicker()
	}*/

	db.sendMessStatus("<%v> module started", db.config.ID)
}

// 1.1 получаем упакованные данные, отправляем на распаковку в отдельной рутине -> processMessage
func (db *Database) processMessages() {
	defer func() {
		if r := recover(); r != nil {
			db.sendMessError("processMessages panic: %v", r)
		}
	}()

	for {
		select {
		case <-db.ctx.Done():
			return
		case msg, ok := <-db.chanInputGen:
			if !ok {
				db.sendMessError("Database: chanInputGen channel closed")
				return
			}
			db.statsMu.Lock()
			db.cntMsgGet++
			db.statsMu.Unlock()

			go func(m types.Message) {
				taskCtx, cancel := context.WithTimeout(db.ctx, time.Duration(db.config.LimitTimeMs)*time.Millisecond)
				defer cancel()

				db.processMessage(taskCtx, m)
			}(msg)

		case msg, ok := <-db.chanInputVue:
			if !ok {
				db.sendMessError("Database: chanInputVue channel closed")
				return
			}
			db.statsMu.Lock()
			db.cntMsgGet++
			db.statsMu.Unlock()

			go func(m types.Message) {

				//log.Printf("xxx: %+v", m)
				switch m.Source {
				case "sendCommand":
					objects.CommandExecute(m)

				case "alarms_get_data":
					select {
					case db.chanOutputDbsA <- m:
					case <-db.ctx.Done():
						return
					default:

					}

				case "trends_get_data":
					select {
					case db.chanOutputDbsT <- m:
					case <-db.ctx.Done():
						return
					default:

					}

				}

			}(msg)

		}

	}
}

// 1.2. распаковываем данные, отправляем каждый тэг на обработку -> processTagValue
func (db *Database) processMessage(ctx context.Context, msg types.Message) {
	defer func() {
		if r := recover(); r != nil {
			db.statsMu.Lock()
			db.cntErr++
			db.statsMu.Unlock()
			db.sendMessError("processMessage panic: %v", r)
		}
	}()

	// Распаковываем TagValue массив
	var tagValues []types.TagValue
	if err := json.Unmarshal(msg.Data, &tagValues); err != nil {
		db.statsMu.Lock()
		db.cntErr++
		db.statsMu.Unlock()
		db.sendMessError("Failed to unmarshal tag values: %v", err)
		return
	}

	// Обрабатываем каждый тег
	for _, tagValue := range tagValues {
		select {
		case <-ctx.Done():
			db.sendMessError("processTagValue cancelled: %s", tagValue.Tag)
			return
		default:
		}

		db.processTagValue(tagValue)
	}

	db.statsMu.Lock()
	db.cntMsgSet += len(tagValues)
	db.statsMu.Unlock()
}

// 1.3. распаковываем данные, отправляем каждый тэг на обработку -> processTagValue
// processTagValue обрабатывает обновление значения тега
func (db *Database) processTagValue(tagValue types.TagValue) {
	defer func() {
		if r := recover(); r != nil {
			db.sendMessError("processTagValue panic for tag %s: %v", tagValue.Alias, r)
		}
	}()

	// 1. Находим тег в базе по alias (безопасный доступ)
	dbTagInterface, exists := db.dbTags.Load(tagValue.Alias)
	if !exists {
		//db.sendMessInfo("Tag not found: %s", tagValue.Alias)
		return
	}

	dbTag, ok := dbTagInterface.(types.DatabaseTag)
	if !ok {
		db.sendMessInfo("Invalid tag type for: %s", tagValue.Alias)
		return
	}

	//if !dbTag.Enable {
	//	return // Тег отключен, пропускаем обработку
	//}

	// 2. Обновляем данные тега
	now := time.Now()
	oldValue := dbTag.Data.Value // Сохраняем старое значение перед обновлением

	// Обновляем основную структуру данных тега
	dbTag.Data = tagValue
	dbTag.ValueOld = oldValue
	dbTag.TimeLastUpdate = now

	// 3. Проверяем изменение значения
	valueChanged := false
	if dbTag.Data.Value != dbTag.ValueOld {
		dbTag.TimeLastChange = now
		valueChanged = true
	}

	// Обновляем тег в базе (безопасная запись)
	db.dbTags.Store(tagValue.Alias, dbTag)

	// 4. Если значение изменилось, обновляем состояние связанных объектов
	if valueChanged {
		db.updateObjectState(tagValue.Alias, tagValue, oldValue)
		// Добавляем обработку трендов при изменении значения
		db.updateTrendData(tagValue.Alias, tagValue)
	}
}

// === STATUS ==========================================================

func (db *Database) processStatus() {
	statusTicker := time.NewTicker(time.Duration(db.config.StatusTimeS) * time.Second)
	defer statusTicker.Stop()

	for {
		select {
		case <-db.ctx.Done():
			return
		case <-statusTicker.C:
			db.sendStatus()
		}
	}
}

func (db *Database) sendStatus() {
	db.statsMu.Lock()
	defer db.statsMu.Unlock()

	status := types.ServiceStatus{
		ModuleID:         db.config.ID,
		Status:           "running",
		LastUpdate:       time.Now(),
		MessagesSent:     db.cntMsgSet,
		MessagesRecv:     db.cntMsgGet,
		ErrorsCount:      db.cntErr,
		StatusOutChannel: db.batchProcessor.GetChannelStats(),
	}

	if db.chanStatus != nil {

		data, _ := json.Marshal(status)
		msg := types.Message{
			Type:   "status",
			Data:   data,
			Source: db.config.ID,
		}

		select {
		case db.chanStatus <- msg:
		default:
			log.Printf("Error channel full")
		}
	}

	db.cntMsgGet = 0
	db.cntMsgSet = 0
	db.cntErr = 0
}

// === MESSAGE ==========================================================

// sendMessage универсальный метод отправки сообщений
func (db *Database) sendMessage(msgType string, format string, args ...interface{}) {
	content := fmt.Sprintf(format, args...)

	// Если канал не nil, отправляем сообщение
	if db.chanSystemMess != nil {
		messageData := types.MessageData{
			Message: content,
			Time:    time.Now().Format(time.RFC3339),
			Source:  db.config.ID,
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
			Source:   db.config.ID,
		}

		// Неблокирующая отправка
		select {
		case db.chanSystemMess <- msg:
			// Сообщение отправлено
		case <-time.After(100 * time.Millisecond):
			log.Printf("WARNING: Message channel timeout for type: %s", msgType)
		case <-db.ctx.Done():
			// Контекст отменен
		}
	}
}

// Специализированные методы для разных типов сообщений
func (db *Database) sendMessError(format string, args ...interface{}) {
	db.sendMessage(types.MessageTypeError, format, args...)
}

func (db *Database) sendMessAlarm(format string, args ...interface{}) {
	db.sendMessage(types.MessageTypeAlarm, format, args...)
}

func (db *Database) sendMessWarning(format string, args ...interface{}) {
	db.sendMessage(types.MessageTypeWarning, format, args...)
}

func (db *Database) sendMessInfo(format string, args ...interface{}) {
	db.sendMessage(types.MessageTypeInfo, format, args...)
}

func (db *Database) sendMessDebug(format string, args ...interface{}) {
	db.sendMessage(types.MessageTypeDebug, format, args...)
}

func (db *Database) sendMessStatus(format string, args ...interface{}) {
	db.sendMessage(types.MessageTypeStatus, format, args...)
}
