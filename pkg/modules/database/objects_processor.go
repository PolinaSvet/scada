package database

import (
	"server-system/pkg/objects"
	"server-system/pkg/types"
)

// === OBJECT ===========================================================

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
