package historian

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"server-system/pkg/types"
)

// TrendProcessor обработчик трендов
type TrendProcessor struct {
	db     *TrendDB
	config *TrendConfig
	ctx    context.Context
	cancel context.CancelFunc
}

// NewTrendProcessor создает новый обработчик трендов
func NewTrendProcessor(config *TrendConfig) (*TrendProcessor, error) {
	if !config.Enable {
		return &TrendProcessor{config: config}, nil
	}

	db, err := NewTrendDB(&config.ConnectSettings)
	if err != nil {
		return nil, fmt.Errorf("failed to create trend db: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	processor := &TrendProcessor{
		db:     db,
		config: config,
		ctx:    ctx,
		cancel: cancel,
	}

	// Запускаем фоновые задачи обслуживания
	if config.Enable && config.AutoMaintenance {
		go processor.startMaintenanceTasks()
	}

	return processor, nil
}

// Close закрывает ресурсы обработчика
func (tp *TrendProcessor) Close() error {
	if tp.cancel != nil {
		tp.cancel()
	}
	if tp.db != nil {
		return tp.db.Close()
	}
	return nil
}

// ProcessBatch обрабатывает батч трендов
func (tp *TrendProcessor) ProcessBatch(ctx context.Context, data []byte) error {
	if !tp.config.Enable {
		return fmt.Errorf("trend processing is disabled")
	}

	// Распарсиваем данные как массив интерфейсов
	var rawItems []interface{}
	if err := json.Unmarshal(data, &rawItems); err != nil {
		return fmt.Errorf("failed to unmarshal raw batch: %w", err)
	}

	if len(rawItems) == 0 {
		return nil
	}

	// Конвертируем []interface{} в []TrendTag
	var trends []types.TrendTag
	for i, rawItem := range rawItems {
		itemData, err := json.Marshal(rawItem)
		if err != nil {
			return fmt.Errorf("failed to marshal item %d: %w", i, err)
		}

		var trendItems []types.TrendTag
		if err := json.Unmarshal(itemData, &trendItems); err != nil {
			return fmt.Errorf("failed to unmarshal item %d into TrendTag: %w", i, err)
		}

		trends = append(trends, trendItems...)
	}

	// Вставляем данные батчами по config.BatchSize
	batchSize := tp.config.BatchSize
	if batchSize <= 0 {
		batchSize = 1000
	}

	log.Println(len(trends), trends)

	totalInserted := 0
	for i := 0; i < len(trends); i += batchSize {
		end := i + batchSize
		if end > len(trends) {
			end = len(trends)
		}

		batch := trends[i:end]
		result, err := tp.db.InsertBatch(ctx, batch)
		if err != nil {
			return fmt.Errorf("failed to insert trend batch [%d:%d]: %w", i, end, err)
		}

		totalInserted += result.Inserted
	}

	log.Printf("trend batch processing completed: %d total records inserted", totalInserted)
	return nil
}

// ProcessGetData обрабатывает запрос данных трендов
func (tp *TrendProcessor) ProcessGetData(ctx context.Context, data map[string]interface{}, outputChan chan<- types.Message, moduleID string) error {
	if !tp.config.Enable {
		return fmt.Errorf("trend processing is disabled")
	}

	// Получаем данные из БД
	results, err := tp.db.GetData(ctx, data)
	if err != nil {
		return fmt.Errorf("failed to get trend data: %w", err)
	}

	// Отправляем результаты в канал
	responseData, err := json.Marshal(results)
	if err != nil {
		return fmt.Errorf("failed to marshal trend response: %w", err)
	}

	msg := types.Message{
		Type:     "trends_set_data",
		Data:     responseData,
		InitDT:   time.Now(),
		UpdateDT: time.Now(),
		Source:   moduleID,
	}

	select {
	case outputChan <- msg:
		log.Printf("trend data sent: %d records", len(results))
	case <-ctx.Done():
		return ctx.Err()
	default:
		return fmt.Errorf("output channel is full")
	}

	return nil
}

// CleanupOldData выполняет очистку старых данных
func (tp *TrendProcessor) CleanupOldData(ctx context.Context, retentionMonths int) error {
	if !tp.config.Enable {
		return fmt.Errorf("trend processing is disabled")
	}

	return tp.db.CleanupOldData(ctx, retentionMonths)
}

// GetStorageStats возвращает статистику хранилища
func (tp *TrendProcessor) GetStorageStats(ctx context.Context) (*StorageStats, error) {
	if !tp.config.Enable {
		return nil, fmt.Errorf("trend processing is disabled")
	}

	return tp.db.GetStorageStats(ctx)
}

// TagsInfoLoadFromJSON загружает данные в tags_info
func (tp *TrendProcessor) TagsInfoLoadFromJSON(ctx context.Context, data []types.TrendTagInfo) (*TagsInfoLoadResult, error) {
	if !tp.config.Enable {
		return nil, fmt.Errorf("trend processing is disabled")
	}

	return tp.db.TagsInfoLoadFromJSON(ctx, data)
}

// TagsInfoGetAllJSON получает все данные из tags_info
func (tp *TrendProcessor) TagsInfoGetAllJSON(ctx context.Context) ([]types.TrendTagInfo, error) {
	if !tp.config.Enable {
		return nil, fmt.Errorf("trend processing is disabled")
	}

	return tp.db.TagsInfoGetAllJSON(ctx)
}

// DailyMaintenance выполняет ежедневное обслуживание
func (tp *TrendProcessor) DailyMaintenance(ctx context.Context) error {
	if !tp.config.Enable {
		return fmt.Errorf("trend processing is disabled")
	}

	return tp.db.DailyMaintenance(ctx)
}

// ReindexTables выполняет переиндексацию таблиц
func (tp *TrendProcessor) ReindexTables(ctx context.Context) (map[string]interface{}, error) {
	if !tp.config.Enable {
		return nil, fmt.Errorf("trend processing is disabled")
	}

	return tp.db.ReindexTables(ctx)
}

// startMaintenanceTasks запускает фоновые задачи обслуживания
func (tp *TrendProcessor) startMaintenanceTasks() {
	// Даем время на инициализацию
	time.Sleep(30 * time.Second)

	// Ежедневное обслуживание
	go tp.startDailyMaintenance()

	// Еженедельная переиндексация
	go tp.startWeeklyReindex()
}

// startDailyMaintenance запускает ежедневное обслуживание с тикером
func (tp *TrendProcessor) startDailyMaintenance() {
	maintenanceTime := tp.config.MaintenanceTime
	if maintenanceTime == "" {
		maintenanceTime = "02:00"
	}

	retentionMonths := tp.config.RetentionMonths
	if retentionMonths <= 0 {
		retentionMonths = 12
	}

	// Вычисляем время до первого запуска
	now := time.Now()
	targetTime, err := time.Parse("15:04", maintenanceTime)
	if err != nil {
		log.Printf("ERROR: Failed to parse maintenance time: %v, using default 02:00", err)
		targetTime, _ = time.Parse("15:04", "02:00")
	}

	nextRun := time.Date(now.Year(), now.Month(), now.Day(),
		targetTime.Hour(), targetTime.Minute(), 0, 0, now.Location())

	if now.After(nextRun) {
		nextRun = nextRun.Add(24 * time.Hour)
	}

	initialDelay := nextRun.Sub(now)
	log.Printf("First daily maintenance at: %s (in %v)", nextRun.Format(time.RFC3339), initialDelay)

	// Создаем таймер для первого запуска
	firstTimer := time.NewTimer(initialDelay)
	defer firstTimer.Stop()

	select {
	case <-firstTimer.C:
		// Выполняем первое обслуживание
		tp.executeMaintenance(retentionMonths)
	case <-tp.ctx.Done():
		return
	}

	// Создаем тикер для последующих запусков (каждые 24 часа)
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			tp.executeMaintenance(retentionMonths)
		case <-tp.ctx.Done():
			return
		}
	}
}

// startWeeklyReindex запускает еженедельную переиндексацию с тикером
func (tp *TrendProcessor) startWeeklyReindex() {
	reindexDay := tp.config.ReindexDay
	if reindexDay < 0 || reindexDay > 6 {
		reindexDay = 0 // Воскресенье по умолчанию
	}

	// Вычисляем время до первого запуска
	now := time.Now()
	currentWeekday := int(now.Weekday())

	// Вычисляем следующий день переиндексации
	daysUntilReindex := (reindexDay - currentWeekday + 7) % 7
	if daysUntilReindex == 0 {
		// Если сегодня нужный день, проверяем время (по умолчанию 03:00)
		targetTime, _ := time.Parse("15:04", "03:00")
		nextRun := time.Date(now.Year(), now.Month(), now.Day(),
			targetTime.Hour(), targetTime.Minute(), 0, 0, now.Location())

		if now.After(nextRun) {
			daysUntilReindex = 7 // Переходим на следующую неделю
		} else {
			// Сегодня в указанное время
			initialDelay := nextRun.Sub(now)
			log.Printf("First reindex today at: %s (in %v)", nextRun.Format(time.RFC3339), initialDelay)

			timer := time.NewTimer(initialDelay)
			select {
			case <-timer.C:
				tp.executeReindex()
			case <-tp.ctx.Done():
				timer.Stop()
				return
			}

			daysUntilReindex = 7 // Следующая переиндексация через неделю
		}
	}

	if daysUntilReindex > 0 {
		// Первый запуск через N дней
		nextRun := time.Date(now.Year(), now.Month(), now.Day()+daysUntilReindex,
			3, 0, 0, 0, now.Location())

		initialDelay := nextRun.Sub(now)
		log.Printf("First reindex at: %s (in %v)", nextRun.Format(time.RFC3339), initialDelay)

		timer := time.NewTimer(initialDelay)
		select {
		case <-timer.C:
			tp.executeReindex()
		case <-tp.ctx.Done():
			timer.Stop()
			return
		}
	}

	// Создаем тикер для еженедельных запусков
	ticker := time.NewTicker(7 * 24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			tp.executeReindex()
		case <-tp.ctx.Done():
			return
		}
	}
}

// executeMaintenance выполняет операции обслуживания
func (tp *TrendProcessor) executeMaintenance(retentionMonths int) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	log.Println("Starting daily maintenance...")

	// Ежедневное обслуживание
	if err := tp.db.DailyMaintenance(ctx); err != nil {
		log.Printf("ERROR: Daily maintenance failed: %v", err)
	} else {
		log.Println("Daily maintenance completed successfully")
	}

	// Очистка старых данных
	if err := tp.db.CleanupOldData(ctx, retentionMonths); err != nil {
		log.Printf("ERROR: Cleanup old data failed: %v", err)
	} else {
		log.Printf("Cleanup old data completed (retention: %d months)", retentionMonths)
	}
}

// executeReindex выполняет переиндексацию
func (tp *TrendProcessor) executeReindex() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Hour)
	defer cancel()

	log.Println("Starting weekly reindex...")

	if _, err := tp.db.ReindexTables(ctx); err != nil {
		log.Printf("ERROR: Reindex failed: %v", err)
	} else {
		log.Println("Reindex completed successfully")
	}
}
