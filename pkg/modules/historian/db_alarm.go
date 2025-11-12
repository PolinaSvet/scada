package historian

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"server-system/pkg/types"

	"github.com/jackc/pgx/v4/pgxpool"
)

// AlarmDB управление подключением к БД алармов
type AlarmDB struct {
	pool   *pgxpool.Pool
	config *ConnectSettings
}

// NewAlarmDB создает новый экземпляр AlarmDB
func NewAlarmDB(config *ConnectSettings) (*AlarmDB, error) {
	if !config.IsEnabled() {
		return nil, fmt.Errorf("alarm db is not enabled")
	}

	connStr := config.GetConnectionString()

	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	pool, err := pgxpool.ConnectConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to alarm database: %w", err)
	}

	// Проверяем подключение
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping alarm database: %w", err)
	}

	log.Printf("alarm database connected: %s:%d/%s", config.Host, config.Port, config.DBName)

	return &AlarmDB{
		pool:   pool,
		config: config,
	}, nil
}

// Close закрывает соединение с БД
func (a *AlarmDB) Close() error {
	if a.pool != nil {
		a.pool.Close()
	}
	return nil
}

/*// InsertBatch вставляет батч алармов в БД
func (a *AlarmDB) InsertBatch(ctx context.Context, alarms []types.AlarmMessDBType) (int, error) {
	if len(alarms) == 0 {
		return 0, nil
	}

	var insertedCount int
	query := `SELECT sinkross_insert_mess_batch($1)`

	err := a.pool.QueryRow(ctx, query, alarms).Scan(&insertedCount)
	if err != nil {
		return 0, fmt.Errorf("failed to insert alarm batch: %w", err)
	}

	return insertedCount, nil
}*/

// InsertBatch вставляет батч алармов в БД
func (a *AlarmDB) InsertBatch(ctx context.Context, alarms []types.AlarmMessDBType) (int, error) {
	if len(alarms) == 0 {
		return 0, nil
	}

	var insertedCount int
	query := `SELECT sinkross_insert_mess_batch($1)`

	// Преобразуем в JSON
	jsonData, err := json.Marshal(alarms)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal alarms to JSON: %w", err)
	}

	err = a.pool.QueryRow(ctx, query, jsonData).Scan(&insertedCount)
	if err != nil {
		return 0, fmt.Errorf("failed to insert alarm batch: %w", err)
	}

	return insertedCount, nil
}

// GetData получает данные алармов из БД
func (a *AlarmDB) GetData(ctx context.Context, params types.AlarmMessGetType) ([]types.AlarmMessDBType, error) {
	query := `SELECT * FROM sinkross_histmess_getdata($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	rows, err := a.pool.Query(ctx, query,
		params.DtStart,
		params.DtEnd,
		params.TagFind,
		params.MessFullFind,
		params.UsoTxtFind,
		params.SeverityFind,
		params.OpermessFind,
		params.KvitFind,
		params.PageNum,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to query alarm data: %w", err)
	}
	defer rows.Close()

	var results []types.AlarmMessDBType
	for rows.Next() {
		var alarm types.AlarmMessDBType
		err := rows.Scan(
			&alarm.ID,
			&alarm.Code,
			&alarm.Dt,
			&alarm.DtTxt,
			&alarm.Tag,
			&alarm.MessFull,
			&alarm.MessName,
			&alarm.MessState,
			&alarm.UsoID,
			&alarm.UsoTxt,
			&alarm.Users,
			&alarm.Severity,
			&alarm.Opermess,
			&alarm.Color,
			&alarm.Kvit,
			&alarm.DtKvit,
			&alarm.DtKvitTxt,
			&alarm.CurrentPage,
			&alarm.TotalPages,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan alarm row: %w", err)
		}
		results = append(results, alarm)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating alarm rows: %w", err)
	}

	return results, nil
}

// HealthCheck проверяет состояние подключения
func (a *AlarmDB) HealthCheck(ctx context.Context) error {
	return a.pool.Ping(ctx)
}

// Stats возвращает статистику пула соединений
func (a *AlarmDB) Stats() *pgxpool.Stat {
	return a.pool.Stat()
}
