// database/objects_manager.go
package database

import (
	"server-system/pkg/types"
	"sync"
)

type ObjectsManager struct {
	// Хранилища конфигов объектов по типам
	configStorages map[string]*sync.Map // objectType -> sync.Map[key]ObjectConfig

	// Хранилища состояний объектов по типам (конкретные типы)
	stateStorages map[string]*sync.Map // objectType -> sync.Map[key]interface{}

	// Мапа конструкторов для каждого типа
	stateConstructors map[string]func() interface{}
}

func NewObjectsManager() *ObjectsManager {
	return &ObjectsManager{
		configStorages:    make(map[string]*sync.Map),
		stateStorages:     make(map[string]*sync.Map),
		stateConstructors: make(map[string]func() interface{}),
	}
}

// RegisterStorage регистрирует хранилища для типа объекта
func (om *ObjectsManager) RegisterStorage(objectType string, configStorage, stateStorage *sync.Map, constructor func() interface{}) {
	om.configStorages[objectType] = configStorage
	om.stateStorages[objectType] = stateStorage
	om.stateConstructors[objectType] = constructor
}

// === Методы для работы с конфигами ===

// LoadConfig загружает конфиг объекта
func (om *ObjectsManager) LoadConfig(objectType, objectKey string) (types.ObjectConfig, bool) {
	storage, exists := om.configStorages[objectType]
	if !exists {
		return types.ObjectConfig{}, false
	}

	objInterface, exists := storage.Load(objectKey)
	if !exists {
		return types.ObjectConfig{}, false
	}

	obj, ok := objInterface.(types.ObjectConfig)
	return obj, ok
}

// StoreConfig сохраняет конфиг объекта
func (om *ObjectsManager) StoreConfig(objectType, objectKey string, obj types.ObjectConfig) {
	storage, exists := om.configStorages[objectType]
	if !exists {
		return
	}

	storage.Store(objectKey, obj)
}

// === Методы для работы с состояниями (конкретные типы) ===

// LoadState загружает состояние объекта как interface{}
func (om *ObjectsManager) LoadState(objectType, objectKey string) (interface{}, bool) {
	storage, exists := om.stateStorages[objectType]
	if !exists {
		return nil, false
	}

	return storage.Load(objectKey)
}

// StoreState сохраняет состояние объекта
func (om *ObjectsManager) StoreState(objectType, objectKey string, state interface{}) {
	storage, exists := om.stateStorages[objectType]
	if !exists {
		return
	}

	storage.Store(objectKey, state)
}

// CreateNewState создает новое состояние для типа объекта
func (om *ObjectsManager) CreateNewState(objectType string) interface{} {
	constructor, exists := om.stateConstructors[objectType]
	if !exists {
		return nil
	}
	return constructor()
}

// HasStorage проверяет наличие хранилищ для типа
func (om *ObjectsManager) HasStorage(objectType string) bool {
	_, configExists := om.configStorages[objectType]
	_, stateExists := om.stateStorages[objectType]
	return configExists && stateExists
}

// GetStateConstructor возвращает конструктор для типа
func (om *ObjectsManager) GetStateConstructor(objectType string) (func() interface{}, bool) {
	constructor, exists := om.stateConstructors[objectType]
	return constructor, exists
}
