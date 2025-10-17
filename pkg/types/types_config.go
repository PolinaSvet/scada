package types

import "time"

// ObjectConfig - общие структуры для конфигураций
// =========================================================
type ModuleConfig struct {
	Enabled    bool   `json:"enabled"`
	ConfigFile string `json:"config_file"`
}

// DatabaseMainConfig - основная структура конфигурации Database
// =========================================================
type DatabaseMainConfig struct {
	LimitTimeMs        int                     `json:"limit_time_ms"`
	StatusTimeS        int                     `json:"status_time_s"`
	ID                 string                  `json:"id"`
	BatchWriting       BatchConfig             `json:"batch_writing"`
	DatabaseConfigPath string                  `json:"database_config_path"`
	Objects            map[string]ModuleConfig `json:"objects"`
}

// =========================================================
type ObjectReference struct {
	ObjectType string // "sensor", "di", etc.
	ObjectKey  string // "sensor_0", "di_1", etc.
}

// ObjectConfig - конфигурация объекта
// =========================================================
type ObjectConfig struct {
	ID         uint16            `json:"id"`
	Type       uint16            `json:"type"`
	CtrlEnable bool              `json:"ctrlEnable"`
	ErrType    uint16            `json:"errType"`
	Info       ObjectInfoConfig  `json:"info"`
	State      ObjectStateConfig `json:"state"`
	Unit       ObjectUnitConfig  `json:"unit"`
	Uso        ObjectUsoConfig   `json:"uso"`
	Alias      map[string]string `json:"alias"`
	Cmd        map[string]string `json:"cmd"`
}

type ObjectInfoConfig struct {
	Tag  string `json:"tag"`
	Desc string `json:"desc"`
	Name string `json:"name"`
	Txt  string `json:"txt"`
	Type uint16 `json:"type"`
}

type ObjectStateConfig struct {
	ColorOff string `json:"colorOff"`
	ColorOn  string `json:"colorOn"`
	TxtOff   string `json:"txtOff"`
	TxtOn    string `json:"txtOn"`
}

type ObjectUnitConfig struct {
	Txt    string `json:"txt"`
	Format string `json:"format"`
}

type ObjectUsoConfig struct {
	ID  uint16 `json:"id"`
	Txt string `json:"txt"`
}

/*type VueObjectState struct {
	StateColor   string                 `json:"stateColor"`
	StateTxt     string                 `json:"stateTxt"`
	State        uint                   `json:"state"`
	Mask         bool                   `json:"mask"`
	Imit         bool                   `json:"imit"`
	Ack          bool                   `json:"ack"`
	RealInput    bool                   `json:"realInput"`
	ChainControl bool                   `json:"chainControl"`
	InputValue   string                 `json:"inputValue"`
	Error        uint                   `json:"error"`
	Alias        map[string]interface{} `json:"alias"`
}*/

type ObjectStateForVue struct {
	ID        string       `json:"id"`
	Type      string       `json:"type"`
	ObjInfo   ObjectConfig `json:"objInfo"`
	ObjVue    interface{}  `json:"objVue"` //VueObjectState
	Timestamp time.Time    `json:"timestamp"`
}

// BatchConfig - конфигурация пакетной обработки
// =========================================================
type BatchConfig struct {
	BufferSize         int `json:"buffer_size"`
	FlushIntervalMs    int `json:"flush_interval_ms"`
	DelayBetweenPackMs int `json:"delay_between_pack_ms"`
	MaxPackSize        int `json:"max_pack_size"`
}
