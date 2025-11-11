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

// TrendConfig конфигурация трендов
type TrendConfig struct {
	Enable          bool            `json:"enable"`
	ConnectSettings ConnectSettings `json:"connect_settings"`
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
