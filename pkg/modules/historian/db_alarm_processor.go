package historian

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"server-system/pkg/types"
)

// AlarmProcessor обработчик алармов
type AlarmProcessor struct {
	db     *AlarmDB
	config *AlarmConfig
}

// NewAlarmProcessor создает новый обработчик алармов
func NewAlarmProcessor(config *AlarmConfig) (*AlarmProcessor, error) {
	if !config.Enable {
		return &AlarmProcessor{config: config}, nil
	}

	db, err := NewAlarmDB(&config.ConnectSettings)
	if err != nil {
		return nil, fmt.Errorf("failed to create alarm db: %w", err)
	}

	return &AlarmProcessor{
		db:     db,
		config: config,
	}, nil
}

// Close закрывает ресурсы обработчика
func (ap *AlarmProcessor) Close() error {
	if ap.db != nil {
		return ap.db.Close()
	}
	return nil
}

/*
// ProcessBatch обрабатывает батч алармов
func (ap *AlarmProcessor) ProcessBatch(ctx context.Context, data []byte) error {
	if !ap.config.Enable {
		return fmt.Errorf("alarm processing is disabled")
	}

	log.Println(data)
	var alarms []types.AlarmMessDBType
	if err := json.Unmarshal(data, &alarms); err != nil {
		return fmt.Errorf("failed to unmarshal alarm batch: %w", err)
	}

	if len(alarms) == 0 {
		return nil
	}

	inserted, err := ap.db.InsertBatch(ctx, alarms)
	if err != nil {
		return fmt.Errorf("failed to insert alarm batch: %w", err)
	}

	log.Printf("alarm batch processed: %d/%d messages inserted", inserted, len(alarms))
	return nil
}*/

// ProcessBatch обрабатывает батч алармов
func (ap *AlarmProcessor) ProcessBatch(ctx context.Context, data []byte) error {
	if !ap.config.Enable {
		return fmt.Errorf("alarm processing is disabled")
	}

	// Распарсиваем данные как массив интерфейсов, поскольку BatchProcessor отправляет []interface{}
	var rawItems []interface{}
	if err := json.Unmarshal(data, &rawItems); err != nil {
		return fmt.Errorf("failed to unmarshal raw batch: %w", err)
	}

	if len(rawItems) == 0 {
		return nil
	}

	// Конвертируем []interface{} в []types.AlarmMessDBType
	for i, rawItem := range rawItems {

		itemData, err := json.Marshal(rawItem)
		if err != nil {
			return fmt.Errorf("failed to marshal item %d: %w", i, err)
		}

		var alarm []types.AlarmMessDBType
		if err := json.Unmarshal(itemData, &alarm); err != nil {
			return fmt.Errorf("failed to unmarshal item %d into AlarmMessDBType: %w", i, err)
		}

		_, err = ap.db.InsertBatch(ctx, alarm)
		if err != nil {
			return fmt.Errorf("failed to insert alarm batch: %w", err)
		}

	}
	return nil
}

// ProcessGetData обрабатывает запрос данных алармов
func (ap *AlarmProcessor) ProcessGetData(ctx context.Context, data map[string]interface{}, outputChan chan<- types.Message, moduleID string) error {
	if !ap.config.Enable {
		return fmt.Errorf("alarm processing is disabled")
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	// Передаем сырые JSON данные напрямую в БД
	results, err := ap.db.GetData(ctx, jsonData)
	if err != nil {
		return fmt.Errorf("failed to get alarm data: %w", err)
	}

	// Отправляем результаты в канал
	responseData, err := json.Marshal(results)
	if err != nil {
		return fmt.Errorf("failed to marshal alarm response: %w", err)
	}

	msg := types.Message{
		Type:     "alarms_set_data",
		Data:     responseData,
		InitDT:   time.Now(),
		UpdateDT: time.Now(),
		Source:   moduleID,
	}

	select {
	case outputChan <- msg:
		log.Printf("alarm data sent: %d records", len(results))
	case <-ctx.Done():
		return ctx.Err()
	default:
		return fmt.Errorf("output channel is full")
	}

	return nil
}

/*func (ap *AlarmProcessor) ProcessGetData(ctx context.Context, data []byte, outputChan chan<- types.Message, moduleID string) error {
	if !ap.config.Enable {
		return fmt.Errorf("alarm processing is disabled")
	}

	var params types.AlarmMessGetType
	if err := json.Unmarshal(data, &params); err != nil {
		return fmt.Errorf("failed to unmarshal alarm get params: %w", err)
	}

	results, err := ap.db.GetData(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to get alarm data: %w", err)
	}

	// Отправляем результаты в канал
	responseData, err := json.Marshal(results)
	if err != nil {
		return fmt.Errorf("failed to marshal alarm response: %w", err)
	}

	msg := types.Message{
		Type:     "alarms_set_data",
		Data:     responseData,
		InitDT:   time.Now(),
		UpdateDT: time.Now(),
		Source:   moduleID,
	}

	select {
	case outputChan <- msg:
		log.Printf("alarm data sent: %d records", len(results))
	case <-ctx.Done():
		return ctx.Err()
	default:
		return fmt.Errorf("output channel is full")
	}

	return nil
}*/
