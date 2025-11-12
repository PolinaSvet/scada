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

	// статистика
	cntMsgGet int
	cntMsgSet int
	cntErr    int
	statsMu   sync.Mutex

	// база тэгов - используем sync.Map для безопасного доступа из горутин
	dbTags sync.Map // ключ: string (alias), значение: types.DatabaseTag

	// Быстрый доступ по алиасам - sync.Map
	aliasIndex sync.Map // ключ: string (alias), значение: []types.ObjectReference

	// объекты - также используем sync.Map
	//objSensors sync.Map // ключ: string, значение: types.ObjectConfig
	//objDi      sync.Map // ключ: string, значение: types.ObjectConfig

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
		db.sendMessError("Failed to load objects data: %v", err)
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

	// Запускаем обработчики в отдельных горутинах
	go db.processStatus()
	go db.processMessages()
	go db.batchProcessor.Start(db.ctx)
	go db.batchProcessorMess.Start(db.ctx)
	go db.batchProcessorHistA.Start(db.ctx)

	//log.Printf("[%v] module started", db.config.ID)
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

			/*taskCtx, cancel := context.WithTimeout(db.ctx, time.Duration(db.config.LimitTimeMs)*time.Millisecond)
			go func(m types.Message, cancelFunc context.CancelFunc) {
				defer cancelFunc()

				db.processMessage(taskCtx, m)
			}(msg, cancel)*/
			go func(m types.Message) {
				taskCtx, cancel := context.WithTimeout(db.ctx, time.Duration(db.config.LimitTimeMs)*time.Millisecond)
				defer cancel()

				db.processMessage(taskCtx, m)
			}(msg)

			//go db.processMessage(msg)

		case msg, ok := <-db.chanInputVue:
			if !ok {
				db.sendMessError("Database: chanInputVue channel closed")
				return
			}
			db.statsMu.Lock()
			db.cntMsgGet++
			db.statsMu.Unlock()

			go func(m types.Message) {
				//taskCtx, cancel := context.WithTimeout(db.ctx, time.Duration(db.config.LimitTimeMs)*time.Millisecond)
				//defer cancel()

				//db.processMessage(taskCtx, m)

				//log.Println("chanInputVue:", m)
				objects.CommandExecute(m)
			}(msg)

			//go db.processMessage(msg)
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
	}
}

// updateObjectState обновляет состояние объектов, связанных с тегом через алиас
func (db *Database) updateObjectState(alias string, tagValue types.TagValue, oldValue interface{}) {
	defer func() {
		if r := recover(); r != nil {
			db.sendMessError("updateObjectState panic for alias %s: %v", alias, r)
		}
	}()

	objectRefsInterface, exists := db.aliasIndex.Load(alias)
	if !exists {
		return
	}

	objectRefs, ok := objectRefsInterface.([]types.ObjectReference)
	if !ok {
		return
	}

	for _, ref := range objectRefs {
		if !db.objectsManager.HasStorage(ref.ObjectType) {
			continue
		}

		// Загружаем конфиг
		config, exists := db.objectsManager.LoadConfig(ref.ObjectType, ref.ObjectKey)
		if !exists {
			continue
		}

		// Загружаем состояние как interface{} или создаем новое
		stateInterface, exists := db.objectsManager.LoadState(ref.ObjectType, ref.ObjectKey)
		if !exists {
			stateInterface = db.objectsManager.CreateNewState(ref.ObjectType)
			if stateInterface == nil {
				continue
			}
		}

		// Вызываем обработчик с конкретным типом
		if handler, exists := objects.Handlers[objects.ObjectType(ref.ObjectType)]; exists {
			// Передаем конкретный тип состояния
			alarmMessages := []types.AlarmMessDBType{}
			handler(&config, &alarmMessages, stateInterface, tagValue, alias, oldValue)

			//log.Println(len(alarmMessages), alarmMessages)

			// Сохраняем обновленные данные
			//db.objectsManager.StoreConfig(ref.ObjectType, ref.ObjectKey, config)
			db.objectsManager.StoreState(ref.ObjectType, ref.ObjectKey, stateInterface)

			// Создаем stateForVue с конкретным типом состояния
			stateForVue := &types.ObjectStateForVue{
				ID:        ref.ObjectKey,
				Type:      ref.ObjectType,
				ObjInfo:   config,
				ObjVue:    stateInterface,
				Timestamp: tagValue.Timestamp,
			}

			db.batchProcessor.Add(stateForVue)
			if len(alarmMessages) > 0 {
				db.batchProcessorMess.Add(alarmMessages)
				db.batchProcessorHistA.Add(alarmMessages)
			}
		}
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

/*
// ========================================

    db.printMapStats()
	db.printSyncMapContents()
	db.findAlias("alias_sensor_0")

func (db *Database) getMapStats() map[string]int {
	stats := make(map[string]int)

	// Считаем dbTags
	db.dbTags.Range(func(key, value interface{}) bool {
		stats["dbTags"]++
		return true
	})

	// Считаем objSensors
	db.objSensors.Range(func(key, value interface{}) bool {
		stats["objSensors"]++
		return true
	})

	// Считаем objDi
	db.objDi.Range(func(key, value interface{}) bool {
		stats["objDi"]++
		return true
	})

	// Считаем aliasIndex
	db.aliasIndex.Range(func(key, value interface{}) bool {
		if refs, ok := value.([]types.ObjectReference); ok {
			stats["aliasIndex_unique"]++
			stats["aliasIndex_total_refs"] += len(refs)
		}
		return true
	})

	return stats
}

// printMapStats выводит статистику
func (db *Database) printMapStats() {
	stats := db.getMapStats()
	log.Println("=== Database Map Statistics ===")
	for key, value := range stats {
		log.Printf("  %s: %d", key, value)
	}
	log.Println("===============================")
}

func (db *Database) printSyncMapContents() {
	log.Println("=== Database Sync.Map Contents ===")

	// Выводим dbTags
	log.Println("--- dbTags ---")
	db.dbTags.Range(func(key, value interface{}) bool {
		if tag, ok := value.(types.DatabaseTag); ok {
			log.Printf("  %s: {Enable: %t, Tag: %s, DataType: %s}",
				key, tag.Enable, tag.Tag, tag.DataType)
		} else {
			log.Printf("  %s: [INVALID TYPE]", key)
		}
		return true
	})

	// Выводим objSensors
	log.Println("--- objSensors ---")
	db.objSensors.Range(func(key, value interface{}) bool {
		if obj, ok := value.(types.ObjectConfig); ok {
			log.Printf("  %s: {Tag: %s, Name: %s}",
				key, obj.Info.Tag, obj.Info.Name)
		} else {
			log.Printf("  %s: [INVALID TYPE]", key)
		}
		return true
	})

	// Выводим objDi
	log.Println("--- objDi ---")
	db.objDi.Range(func(key, value interface{}) bool {
		if obj, ok := value.(types.ObjectConfig); ok {
			log.Printf("  %s: {Tag: %s, Name: %s}",
				key, obj.Info.Tag, obj.Info.Name)
		} else {
			log.Printf("  %s: [INVALID TYPE]", key)
		}
		return true
	})

	// Выводим aliasIndex
	log.Println("--- aliasIndex ---")
	db.aliasIndex.Range(func(key, value interface{}) bool {
		if refs, ok := value.([]types.ObjectReference); ok {
			log.Printf("  Алиас '%s':", key)
			for i, ref := range refs {
				log.Printf("    %d. Type: %s, Key: %s", i+1, ref.ObjectType, ref.ObjectKey)
			}
		} else {
			log.Printf("  %s: [INVALID TYPE]", key)
		}
		return true
	})

	log.Println("=== End of Contents ===")
}

func (db *Database) findAlias(alias string) {
	refsInterface, exists := db.aliasIndex.Load(alias)
	if !exists {
		log.Printf("Алиас '%s' не найден", alias)
		return
	}

	refs, ok := refsInterface.([]types.ObjectReference)
	if !ok {
		log.Printf("Алиас '%s': неверный тип данных", alias)
		return
	}

	log.Printf("Алиас '%s' найден, ссылается на %d объектов:", alias, len(refs))
	for i, ref := range refs {
		log.Printf("  %d. Type: %s, Key: %s", i+1, ref.ObjectType, ref.ObjectKey)

		// Дополнительная информация об объекте
		switch ref.ObjectType {
		case "sensor":
			if objInterface, exists := db.objSensors.Load(ref.ObjectKey); exists {
				if obj, ok := objInterface.(types.ObjectConfig); ok {
					log.Printf("     Object: %s (%s)", obj.Info.Name, obj.Info.Tag)
				}
			}
		case "di":
			if objInterface, exists := db.objDi.Load(ref.ObjectKey); exists {
				if obj, ok := objInterface.(types.ObjectConfig); ok {
					log.Printf("     Object: %s (%s)", obj.Info.Name, obj.Info.Tag)
				}
			}
		}
	}
}
*/
