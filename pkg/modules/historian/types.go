// database/config.go
package historian

// HistorianConfig основная конфигурация модуля
type HistorianConfig struct {
	LimitTimeMs int         `json:"limit_time_ms"`
	StatusTimeS int         `json:"status_time_s"`
	ID          string      `json:"id"`
	Alarm       AlarmConfig `json:"alarm"`
	Trend       TrendConfig `json:"trend"`
}

// AlarmConfig конфигурация алармов
type AlarmConfig struct {
	Enable          bool            `json:"enable"`
	ConnectSettings ConnectSettings `json:"connect_settings"`
	NewTable        NewTableConfig  `json:"new_table"`
}

// TrendConfig конфигурация обработчика трендов
type TrendConfig struct {
	Enable          bool            `json:"enable"`
	ConnectSettings ConnectSettings `json:"connect_settings"`
	BatchSize       int             `json:"batch_size"`
	MaintenanceTime string          `json:"maintenance_time"` // Время ежедневного обслуживания (формат: "02:00")
	ReindexDay      int             `json:"reindex_day"`      // День недели для переиндексации (0-6, где 0=воскресенье)
	RetentionMonths int             `json:"retention_months"` // Сколько месяцев хранить данные
	AutoMaintenance bool            `json:"auto_maintenance"` // Автоматическое обслуживание
}

// ConnectSettings настройки подключения к БД
type ConnectSettings struct {
	Type     string `json:"type"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"` // Зашифрованный пароль из конфига
	DBName   string `json:"dbname"`
	SSLMode  string `json:"sslmode"`

	// DecryptedPassword будет заполняться после дешифровки
	DecryptedPassword string
}

// NewTableConfig конфигурация создания новых таблиц
type NewTableConfig struct {
	Enable      bool   `json:"enable"`
	Name        string `json:"name"`
	Every       string `json:"every"`
	DurationDay int    `json:"duration_day"`
}

// InsertResult результат вставки данных
type InsertResult struct {
	Inserted      int     `json:"inserted"`
	ExecutionDtMs float64 `json:"execution_dt_ms"`
}

// StorageStats статистика хранилища
type StorageStats struct {
	TotalPartitions int     `json:"total_partitions"`
	TotalRecords    int     `json:"total_records"`
	StorageSizeGB   float64 `json:"storage_size_gb"`
	OldestDataDays  float64 `json:"oldest_data_days"`
}

// TagsInfoLoadResult результат загрузки tags_info
type TagsInfoLoadResult struct {
	DeletedRecords  int    `json:"deleted_records"`
	InsertedRecords int    `json:"inserted_records"`
	Status          string `json:"status"`
}
