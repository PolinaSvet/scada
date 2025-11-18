package database

import (
	"log"
	"server-system/pkg/types"
	"time"
)

// === TREND ===========================================================
// SaveType = 0, пишем только по изменению тэга
// SaveType = 1, пишем по изменению тэга и каждые SaveTimeMs
// SaveType = 2, пишем каждые SaveTimeMs

// updateTrendData обновляет данные трендов и отправляет в batchProcessorHistT при необходимости
func (db *Database) updateTrendData(alias string, tagValue types.TagValue) {
	defer func() {
		if r := recover(); r != nil {
			db.sendMessError("updateTrendData panic for alias %s: %v", alias, r)
		}
	}()

	// Проверяем, включена ли функциональность трендов
	if !db.config.TrendConfig.Enable {
		return
	}

	// Ищем запись в dbTrend
	trendInfoInterface, exists := db.dbTrend.Load(alias)
	if !exists {
		return
	}

	trendInfo, ok := trendInfoInterface.(types.TrendTagInfo)
	if !ok {
		db.sendMessInfo("Invalid trend tag type for: %s", alias)
		return
	}

	// Проверяем, включен ли тренд для этого тега
	if !trendInfo.Enable {
		return
	}

	// Преобразуем значение в float64
	var floatValue float64
	switch v := tagValue.Value.(type) {
	case float64:
		floatValue = v
	case float32:
		floatValue = float64(v)
	case int:
		floatValue = float64(v)
	case int64:
		floatValue = float64(v)
	case bool:
		if v {
			floatValue = 1.0
		} else {
			floatValue = 0.0
		}
	default:
		floatValue = 0.0
		db.sendMessInfo("Unknown value type for trend %s: %T", alias, tagValue.Value)
	}

	// Обновляем данные тренда
	trendInfo.Data = types.TrendTag{
		IdObj:   trendInfo.ID,
		Value:   floatValue,
		Quality: tagValue.Quality,
		Dt:      tagValue.Timestamp.UTC().UnixMilli(),
	}

	// Сохраняем обновленные данные
	db.dbTrend.Store(alias, trendInfo)

	// Отправляем данные в batchProcessorHistT в зависимости от SaveType
	saveType := db.config.TrendConfig.SaveType
	if saveType == 0 || saveType == 1 {
		// Создаем массив с одним элементом
		trendBatch := []types.TrendTag{trendInfo.Data}
		db.sendTrendToBatch(trendBatch)
	}
}

// sendTrendToBatch отправляет массив данных тренда в batchProcessorHistT
func (db *Database) sendTrendToBatch(trendData []types.TrendTag) {
	defer func() {
		if r := recover(); r != nil {
			db.sendMessError("sendTrendToBatch panic: %v", r)
		}
	}()

	if len(trendData) == 0 {
		return
	}

	db.batchProcessorHistT.Add(trendData)
}

// trendTicker периодически отправляет данные трендов в batchProcessorHistT
func (db *Database) trendTicker() {
	defer func() {
		if r := recover(); r != nil {
			db.sendMessError("trendTicker panic: %v", r)
		}
	}()

	if !db.config.TrendConfig.Enable || db.config.TrendConfig.SaveType == 0 {
		return
	}

	saveTimeMs := db.config.TrendConfig.SaveTimeMs
	if saveTimeMs <= 0 {
		saveTimeMs = 1000 // значение по умолчанию
	}

	ticker := time.NewTicker(time.Duration(saveTimeMs) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-db.ctx.Done():
			return
		case <-ticker.C:
			db.processAllTrends()
		}
	}
}

// processAllTrends обрабатывает все тренды и отправляет их в batchProcessorHistT батчами
func (db *Database) processAllTrends() {
	defer func() {
		if r := recover(); r != nil {
			db.sendMessError("processAllTrends panic: %v", r)
		}
	}()

	batchSize := db.config.TrendConfig.BatchSize
	if batchSize <= 0 {
		batchSize = 1000
	}

	var trendBatch []types.TrendTag
	count := 0

	// Перебираем все записи в dbTrend
	db.dbTrend.Range(func(key, value interface{}) bool {
		_, ok := key.(string)
		if !ok {
			return true
		}

		trendInfo, ok := value.(types.TrendTagInfo)
		if !ok {
			return true
		}

		// Проверяем, включен ли тренд для этого тега
		if !trendInfo.Enable {
			return true
		}

		// Добавляем в текущий батч
		trendBatch = append(trendBatch, trendInfo.Data)
		count++

		// Если батч заполнен, отправляем и создаем новый
		if len(trendBatch) >= batchSize {
			db.sendTrendToBatch(trendBatch)
			trendBatch = nil // Сбрасываем батч
		}

		return true
	})

	// Отправляем оставшиеся данные, если они есть
	if len(trendBatch) > 0 {
		db.sendTrendToBatch(trendBatch)
	}

	log.Printf("Processed %d trend tags in batches of %d", count, batchSize)
}
