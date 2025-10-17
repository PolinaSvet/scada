// database/config.go
package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"server-system/pkg/types"
)

// основная конфигурация модуля
func (db *Database) loadMainConfig(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("не удалось прочитать файл: %w", err)
	}

	var config types.DatabaseMainConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		return fmt.Errorf("ошибка парсинга JSON: %w", err)
	}
	db.config = config

	if err := db.validateMainConfig(); err != nil {
		return fmt.Errorf("ошибка валидации конфига: %w", err)
	}

	return nil
}

func (db *Database) saveMainConfig(filename string) error {
	data, err := json.MarshalIndent(db.config, "", "  ")
	if err != nil {
		return fmt.Errorf("ошибка сериализации JSON: %w", err)
	}

	err = ioutil.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("не удалось записать файл: %w", err)
	}

	return nil
}

func (db *Database) validateMainConfig() error {
	if len(db.config.Objects) == 0 {
		return errors.New("поле 'objects' обязательно и должно быть заполнено")
	}
	return nil
}

// конфигурация тэгов
func (db *Database) loadTagsConfig(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	var config types.TagsConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}

	for key, tag := range config.Tags {
		db.dbTags.Store(key, tag)
	}

	return nil
}

// конфигурация объектов через ObjectsManager
func (db *Database) loadObjectsConfig() error {
	for moduleName, moduleConfig := range db.config.Objects {
		if !moduleConfig.Enabled {
			continue
		}

		data, err := ioutil.ReadFile(moduleConfig.ConfigFile)
		if err != nil {
			return err
		}

		var moduleData types.ObjectsConfigFile
		if err := json.Unmarshal(data, &moduleData); err != nil {
			return err
		}

		// Определяем тип объекта по имени модуля
		objectType := moduleName
		if objectType == "" {
			continue // Пропускаем неизвестные типы
		}

		// Загружаем объекты через ObjectsManager
		for key, obj := range moduleData.Objects {
			db.objectsManager.StoreConfig(objectType, key, obj)
		}
	}

	db.buildAliasIndex()
	return nil
}

// buildAliasIndex строит индекс алиасов через ObjectsManager
func (db *Database) buildAliasIndex() {
	// Получаем все зарегистрированные типы объектов
	for objectType := range db.objectsManager.configStorages {
		// Для каждого типа объектов строим индекс
		storage, exists := db.objectsManager.configStorages[objectType]
		if !exists {
			continue
		}

		storage.Range(func(key, value interface{}) bool {
			obj, ok := value.(types.ObjectConfig)
			if !ok {
				return true
			}

			// Индексируем все алиасы объекта
			for _, alias := range obj.Alias {
				refsInterface, _ := db.aliasIndex.LoadOrStore(alias, []types.ObjectReference{})
				refs := refsInterface.([]types.ObjectReference)
				refs = append(refs, types.ObjectReference{
					ObjectType: objectType, // Используем единый тип
					ObjectKey:  key.(string),
				})
				db.aliasIndex.Store(alias, refs)
			}
			return true
		})
	}
}

/*package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"server-system/pkg/types"
)

// основная конфигурация модуля
// MainConfig configs/config_database.json
// ================================================================================
func (db *Database) loadMainConfig(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("не удалось прочитать файл: %w", err)
	}

	var config types.DatabaseMainConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		return fmt.Errorf("ошибка парсинга JSON: %w", err)
	}
	db.config = config

	if err := db.validateMainConfig(); err != nil {
		return fmt.Errorf("ошибка валидации конфига: %w", err)

	}

	return nil
}

func (db *Database) saveMainConfig(filename string) error {

	data, err := json.MarshalIndent(db.config, "", "  ")
	if err != nil {
		return fmt.Errorf("ошибка сериализации JSON: %w", err)
	}

	err = ioutil.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("не удалось записать файл: %w", err)
	}

	return nil
}

func (db *Database) validateMainConfig() error {
	if len(db.config.Objects) == 0 {
		return errors.New("поле 'objects' обязательно и должно быть заполнено")
	}
	// Добавляем необходимые проверки
	return nil
}

// конфигурация тэгов
// "database_config_path": "config/objects/database.json"
// ================================================================================

// loadTagsConfig загружает теги (адаптирован для sync.Map)
func (db *Database) loadTagsConfig(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	var config types.TagsConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}

	for key, tag := range config.Tags {
		db.dbTags.Store(key, tag)
	}

	return nil
}

// конфигурация объектов
// "config/objects/objsensor.json", "config/objects/objdi.json", ...
// ========================================================================
// loadObjectsConfig загружает объекты (адаптирован для sync.Map)
func (db *Database) loadObjectsConfig() error {
	for moduleName, moduleConfig := range db.config.Objects {
		if !moduleConfig.Enabled {
			continue
		}

		data, err := ioutil.ReadFile(moduleConfig.ConfigFile)
		if err != nil {
			return err
		}

		var moduleData types.ObjectsConfigFile
		if err := json.Unmarshal(data, &moduleData); err != nil {
			return err
		}

		switch moduleName {
		case "objsensor":
			for key, obj := range moduleData.Objects {
				db.objSensorsConfig.Store(key, obj)
			}
		case "objdi":
			for key, obj := range moduleData.Objects {
				db.objDiConfig.Store(key, obj)
			}
		}
	}

	db.buildAliasIndex()
	return nil
}

// buildAliasIndex строит индекс алиасов (адаптирован для sync.Map)
func (db *Database) buildAliasIndex() {
	// Обрабатываем сенсоры
	db.objSensorsConfig.Range(func(key, value interface{}) bool {
		obj, ok := value.(types.ObjectConfig)
		if !ok {
			return true
		}

		for _, alias := range obj.Alias {
			refsInterface, _ := db.aliasIndex.LoadOrStore(alias, []types.ObjectReference{})
			refs := refsInterface.([]types.ObjectReference)
			refs = append(refs, types.ObjectReference{
				ObjectType: "sensor",
				ObjectKey:  key.(string),
			})
			db.aliasIndex.Store(alias, refs)
		}
		return true
	})

	// Обрабатываем DI объекты
	db.objDiConfig.Range(func(key, value interface{}) bool {
		obj, ok := value.(types.ObjectConfig)
		if !ok {
			return true
		}

		for _, alias := range obj.Alias {
			refsInterface, _ := db.aliasIndex.LoadOrStore(alias, []types.ObjectReference{})
			refs := refsInterface.([]types.ObjectReference)
			refs = append(refs, types.ObjectReference{
				ObjectType: "di",
				ObjectKey:  key.(string),
			})
			db.aliasIndex.Store(alias, refs)
		}
		return true
	})
}*/
