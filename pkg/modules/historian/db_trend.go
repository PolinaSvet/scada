package historian

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/jackc/pgx/v4/pgxpool"

	"server-system/pkg/types"
)

// TrendDB управление подключением к БД трендов
type TrendDB struct {
	pool   *pgxpool.Pool
	config *ConnectSettings
}

// NewTrendDB создает новый экземпляр TrendDB
func NewTrendDB(config *ConnectSettings) (*TrendDB, error) {
	if !config.IsEnabled() {
		return nil, fmt.Errorf("trend db is not enabled")
	}

	connStr := config.GetConnectionString()

	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	pool, err := pgxpool.ConnectConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to trend database: %w", err)
	}

	// Проверяем подключение
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping trend database: %w", err)
	}

	log.Printf("trend database connected: %s:%d/%s", config.Host, config.Port, config.DBName)

	return &TrendDB{
		pool:   pool,
		config: config,
	}, nil
}

// Close закрывает соединение с БД
func (t *TrendDB) Close() error {
	if t.pool != nil {
		t.pool.Close()
	}
	return nil
}

// InsertBatch вставляет батч трендов в БД
func (t *TrendDB) InsertBatch(ctx context.Context, trends []types.TrendTag) (*InsertResult, error) {
	if len(trends) == 0 {
		return &InsertResult{Inserted: 0}, nil
	}

	var resultJSON []byte
	query := `SELECT sinkross_insert_mess_batch($1)`

	jsonData, err := json.Marshal(trends)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal trends to JSON: %w", err)
	}

	err = t.pool.QueryRow(ctx, query, jsonData).Scan(&resultJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to insert trend batch: %w", err)
	}

	var insertResult InsertResult
	if err := json.Unmarshal(resultJSON, &insertResult); err != nil {
		return nil, fmt.Errorf("failed to parse insert result: %w", err)
	}

	return &insertResult, nil
}

// GetData получает данные трендов из БД
func (t *TrendDB) GetData(ctx context.Context, params map[string]interface{}) ([]types.TrendTag, error) {
	if len(params) == 0 {
		return nil, fmt.Errorf("empty parameters")
	}

	jsonData, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal params to JSON: %w", err)
	}

	query := `SELECT * FROM sinkross_histmess_getdata_json($1)`

	rows, err := t.pool.Query(ctx, query, jsonData)
	if err != nil {
		return nil, fmt.Errorf("failed to query trend data with JSON: %w", err)
	}
	defer rows.Close()

	var results []types.TrendTag
	for rows.Next() {
		var trend types.TrendTag

		err := rows.Scan(
			&trend.ID,
			&trend.IdObj,
			&trend.Value,
			&trend.Quality,
			&trend.Dt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan trend row: %w", err)
		}

		results = append(results, trend)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating trend rows: %w", err)
	}

	return results, nil
}

// CleanupOldData удаляет старые данные
func (t *TrendDB) CleanupOldData(ctx context.Context, retentionMonths int) error {
	query := `SELECT cleanup_old_data($1)`

	var resultJSON []byte
	err := t.pool.QueryRow(ctx, query, retentionMonths).Scan(&resultJSON)
	if err != nil {
		return fmt.Errorf("failed to cleanup old data: %w", err)
	}

	return nil
}

// GetStorageStats возвращает статистику хранилища
func (t *TrendDB) GetStorageStats(ctx context.Context) (*StorageStats, error) {
	query := `SELECT * FROM get_storage_stats()`

	rows, err := t.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get storage stats: %w", err)
	}
	defer rows.Close()

	stats := &StorageStats{}
	for rows.Next() {
		var metric string
		var value float64

		err := rows.Scan(&metric, &value)
		if err != nil {
			return nil, fmt.Errorf("failed to scan storage stats: %w", err)
		}

		switch metric {
		case "total_partitions":
			stats.TotalPartitions = int(value)
		case "total_records":
			stats.TotalRecords = int(value)
		case "storage_size_gb":
			stats.StorageSizeGB = value
		case "oldest_data_days":
			stats.OldestDataDays = value
		}
	}

	return stats, nil
}

// TagsInfoLoadFromJSON загружает данные в tags_info из JSON
func (t *TrendDB) TagsInfoLoadFromJSON(ctx context.Context, data []types.TrendTagInfo) (*TagsInfoLoadResult, error) {
	var resultJSON []byte
	query := `SELECT tags_info_load_from_json($1)`

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tags info to JSON: %w", err)
	}

	err = t.pool.QueryRow(ctx, query, jsonData).Scan(&resultJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to load tags info: %w", err)
	}

	var loadResult TagsInfoLoadResult
	if err := json.Unmarshal(resultJSON, &loadResult); err != nil {
		return nil, fmt.Errorf("failed to parse load result: %w", err)
	}

	return &loadResult, nil
}

// TagsInfoGetAllJSON получает все данные из tags_info в JSON
func (t *TrendDB) TagsInfoGetAllJSON(ctx context.Context) ([]types.TrendTagInfo, error) {
	query := `SELECT tags_info_get_all_json()`

	var resultJSON []byte
	err := t.pool.QueryRow(ctx, query).Scan(&resultJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to get tags info: %w", err)
	}

	var tagsInfo []types.TrendTagInfo
	if err := json.Unmarshal(resultJSON, &tagsInfo); err != nil {
		return nil, fmt.Errorf("failed to parse tags info: %w", err)
	}

	return tagsInfo, nil
}

// DailyMaintenance выполняет ежедневное обслуживание
func (t *TrendDB) DailyMaintenance(ctx context.Context) error {
	query := `SELECT daily_maintenance()`

	_, err := t.pool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to execute daily maintenance: %w", err)
	}

	return nil
}

// ReindexTables выполняет переиндексацию таблиц
func (t *TrendDB) ReindexTables(ctx context.Context) (map[string]interface{}, error) {
	query := `SELECT reindex_trend_tables()`

	var resultJSON []byte
	err := t.pool.QueryRow(ctx, query).Scan(&resultJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to reindex tables: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to parse reindex result: %w", err)
	}

	return result, nil
}

// HealthCheck проверяет состояние подключения
func (t *TrendDB) HealthCheck(ctx context.Context) error {
	return t.pool.Ping(ctx)
}

// Stats возвращает статистику пула соединений
func (t *TrendDB) Stats() *pgxpool.Stat {
	return t.pool.Stat()
}
