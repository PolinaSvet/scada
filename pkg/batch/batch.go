package batch

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"server-system/pkg/types"
	"sync"
	"time"
)

type BatchProcessor struct {
	buffer         []interface{}
	bufferMu       sync.Mutex // Защита буфера от конкурентного доступа
	outputChan     chan<- types.Message
	systemMessChan chan<- types.Message
	config         types.BatchConfig
	moduleID       string

	// Детектор "мертвых" каналов
	stats struct {
		consecutiveTimeouts int
		totalSent           int
		totalDropped        int
		lastSuccessTime     time.Time
		channelAlive        bool
	}
	statsMu sync.Mutex
}

func NewBatchProcessor(outputChan chan<- types.Message, systemMessChan chan<- types.Message,
	config types.BatchConfig, moduleID string) *BatchProcessor {

	return &BatchProcessor{
		buffer:         make([]interface{}, 0, config.BufferSize),
		outputChan:     outputChan,
		systemMessChan: systemMessChan,
		config:         config,
		moduleID:       moduleID,
	}
}

func (bp *BatchProcessor) Start(ctx context.Context) {
	defer bp.handlePanic("BatchProcessor Start")

	bp.sendMessStatus("<%v> batch processor config: %+v", bp.moduleID, bp.config)

	flushInterval := time.Duration(bp.config.FlushIntervalMs) * time.Millisecond
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			bp.flush() // Флашим при остановке
			log.Printf("%s: Batch processor stopped", bp.moduleID)
			return
		case <-ticker.C:
			bp.flush()
		}
	}
}

func (bp *BatchProcessor) Add(item interface{}) {
	defer bp.handlePanic("BatchProcessor Add")

	bp.bufferMu.Lock()
	defer bp.bufferMu.Unlock()

	bp.buffer = append(bp.buffer, item)

	// Если буфер заполнен, флашим немедленно
	if len(bp.buffer) >= bp.config.BufferSize {
		go bp.flush() // Запускаем в отдельной горутине чтобы не блокировать Add
	}
}

func (bp *BatchProcessor) flush() {
	defer bp.handlePanic("BatchProcessor flush")

	bp.bufferMu.Lock()
	defer bp.bufferMu.Unlock()

	if len(bp.buffer) == 0 {
		return
	}

	log.Printf("%s: Flushing buffer with %d items", bp.moduleID, len(bp.buffer))

	// Делим на пакеты по MaxPackSize
	for len(bp.buffer) > 0 {
		batchSize := len(bp.buffer)
		if batchSize > bp.config.MaxPackSize {
			batchSize = bp.config.MaxPackSize
		}

		// Проверяем границы массива
		if batchSize <= 0 || batchSize > len(bp.buffer) {
			log.Printf("%s: Invalid batch size: %d, buffer length: %d",
				bp.moduleID, batchSize, len(bp.buffer))
			break
		}

		// Создаем пакет
		batchItems := make([]interface{}, batchSize)
		copy(batchItems, bp.buffer[:batchSize])

		// Безопасно обрезаем буфер
		if batchSize < len(bp.buffer) {
			bp.buffer = bp.buffer[batchSize:]
		} else {
			bp.buffer = bp.buffer[:0] // Полностью очищаем
		}

		// Отправляем пакет
		bp.sendBatch(batchItems)

		// Задержка между пакетами
		if len(bp.buffer) > 0 && bp.config.DelayBetweenPackMs > 0 {
			delay := time.Duration(bp.config.DelayBetweenPackMs) * time.Millisecond
			time.Sleep(delay)
		}
	}
}

func (bp *BatchProcessor) sendBatch(items []interface{}) {
	defer bp.handlePanic("BatchProcessor sendBatch")

	if len(items) == 0 {
		return
	}
	if bp.isChannelDead() {
		// Канал считается мертвым, не пытаемся отправлять
		bp.recordDroppedItems(len(items))
		return
	}

	data, err := json.Marshal(items)
	if err != nil {
		bp.sendMessError("Failed to marshal batch %v", err)
		return
	}

	msg := types.Message{
		ID:       "batch_" + time.Now().Format("150405.000"),
		Type:     "data_batch",
		Data:     data,
		UpdateDT: time.Now(),
		Source:   bp.moduleID,
	}

	select {
	case bp.outputChan <- msg:
		bp.recordSuccess(len(items))
	case <-time.After(100 * time.Millisecond):
		bp.recordTimeout(len(items))
	}
}

// isChannelDead проверяет считается ли канал "мертвым"
func (bp *BatchProcessor) isChannelDead() bool {
	bp.statsMu.Lock()
	defer bp.statsMu.Unlock()

	// Канал мертв если:
	// - более 10 таймаутов подряд И
	// - прошло больше 30 секунд с последней успешной отправки
	isDead := bp.stats.consecutiveTimeouts > 10 &&
		time.Since(bp.stats.lastSuccessTime) > 30*time.Second

	if isDead && bp.stats.channelAlive {
		bp.stats.channelAlive = false
		bp.sendMessError("Output channel confirmed dead - stopping sends")
	}

	return isDead
}

func (bp *BatchProcessor) recordSuccess(count int) {
	bp.statsMu.Lock()
	defer bp.statsMu.Unlock()

	bp.stats.consecutiveTimeouts = 0
	bp.stats.totalSent += count
	bp.stats.lastSuccessTime = time.Now()

	if !bp.stats.channelAlive {
		bp.stats.channelAlive = true
		//bp.sendError("Output channel recovered - resuming sends", nil)
	}
}

func (bp *BatchProcessor) recordTimeout(count int) {
	bp.statsMu.Lock()
	defer bp.statsMu.Unlock()

	bp.stats.consecutiveTimeouts++
	bp.stats.totalDropped += count

	// Логируем только первые несколько таймаутов и потом периодически
	if bp.stats.consecutiveTimeouts <= 3 || bp.stats.consecutiveTimeouts%10 == 0 {
		bp.sendMessError(fmt.Sprintf("Output channel timeout %d, dropped %d items (total dropped: %d)",
			bp.stats.consecutiveTimeouts, count, bp.stats.totalDropped), nil)
	}
}

func (bp *BatchProcessor) recordDroppedItems(count int) {
	bp.statsMu.Lock()
	defer bp.statsMu.Unlock()

	bp.stats.totalDropped += count

	// Логируем потери каждые 1000 элементов
	if bp.stats.totalDropped%1000 == 0 {
		bp.sendMessError(fmt.Sprintf("Channel dead - total dropped items: %d",
			bp.stats.totalDropped), nil)
	}
}

// GetChannelStats возвращает статистику канала для мониторинга
func (bp *BatchProcessor) GetChannelStats() types.StatusChannelAlive {
	bp.statsMu.Lock()
	defer bp.statsMu.Unlock()

	return types.StatusChannelAlive{
		ChannelAlive:         bp.stats.channelAlive,
		ConsecutiveTimeouts:  bp.stats.consecutiveTimeouts,
		TotalSent:            bp.stats.totalSent,
		TotalDropped:         bp.stats.totalDropped,
		SecondsSinceLastSend: time.Since(bp.stats.lastSuccessTime).Seconds(),
	}
}

func (bp *BatchProcessor) handlePanic(method string) {
	if r := recover(); r != nil {
		bp.sendMessError("%s: PANIC in %s: %v", bp.moduleID, method, r)
	}
}

// ForceFlush принудительно отправляет все данные
func (bp *BatchProcessor) ForceFlush() {
	bp.flush()
}

// GetBufferSize возвращает текущий размер буфера (для отладки)
func (bp *BatchProcessor) GetBufferSize() int {
	bp.bufferMu.Lock()
	defer bp.bufferMu.Unlock()
	return len(bp.buffer)
}

// sendMessage универсальный метод отправки сообщений
func (bp *BatchProcessor) sendMessage(msgType string, format string, args ...interface{}) {
	content := fmt.Sprintf(format, args...)

	// Если канал не nil, отправляем сообщение
	if bp.systemMessChan != nil {
		messageData := types.MessageData{
			Message: content,
			Time:    time.Now().Format(time.RFC3339),
			Source:  bp.moduleID,
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
			Source:   bp.moduleID,
		}

		// Неблокирующая отправка
		select {
		case bp.systemMessChan <- msg:
			// Сообщение отправлено
		case <-time.After(100 * time.Millisecond):
			log.Printf("WARNING: Message channel timeout for type: %s", msgType)
			//case <-bp.ctx.Done():
			//	// Контекст отменен
		}
	}
}

// Специализированные методы для разных типов сообщений
func (bp *BatchProcessor) sendMessError(format string, args ...interface{}) {
	bp.sendMessage(types.MessageTypeError, format, args...)
}

func (bp *BatchProcessor) sendMessInfo(format string, args ...interface{}) {
	bp.sendMessage(types.MessageTypeInfo, format, args...)
}

func (bp *BatchProcessor) sendMessStatus(format string, args ...interface{}) {
	bp.sendMessage(types.MessageTypeStatus, format, args...)
}
