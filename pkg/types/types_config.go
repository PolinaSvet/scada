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
	Alarm      map[string]int    `json:"alarm"`
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

// AlarmMess - конфигурация сообщений
// =========================================================
// AlarmMessGetType параметры для запроса данных алармов
type AlarmMessGetType struct {
	DtStart      int64  `json:"dt_start"`
	DtEnd        int64  `json:"dt_end"`
	TagFind      string `json:"tag_find"`
	MessFullFind string `json:"mess_full_find"`
	UsoTxtFind   string `json:"uso_txt_find"`
	SeverityFind int    `json:"severity_find"`
	OpermessFind int    `json:"opermess_find"`
	KvitFind     int    `json:"kvit_find"`
	PageNum      int    `json:"page_num"`
}

// AlarmMessDBType соответствует структуре из БД
type AlarmMessDBType struct {
	ID          int64  `json:"id"`
	IdObj       uint16 `json:"id_obj"`
	TypeObj     uint16 `json:"type_obj"`
	Code        int64  `json:"code"`
	Dt          int64  `json:"dt"`
	DtTxt       string `json:"dt_txt"`
	Tag         string `json:"tag"`
	MessFull    string `json:"mess_full"`
	MessName    string `json:"mess_name"`
	MessState   string `json:"mess_state"`
	UsoID       int    `json:"uso_id"`
	UsoTxt      string `json:"uso_txt"`
	Users       string `json:"users"`
	Severity    int    `json:"severity"`
	Opermess    int    `json:"opermess"`
	Color       string `json:"color"`
	Kvit        bool   `json:"kvit"`
	DtKvit      int64  `json:"dt_kvit"`
	DtKvitTxt   string `json:"dt_kvit_txt"`
	CurrentPage int    `json:"current_page"`
	TotalPages  int    `json:"total_pages"`
}

/*type AlarmMess struct {
	ID        uint16           `json:"id"`
	Type      uint16           `json:"type"`
	Info      ObjectInfoConfig `json:"info"`
	Uso       ObjectUsoConfig  `json:"uso"`
	MessColor string           `json:"messColor"`
	MessTxt   string           `json:"messTxt"`
	MessType  int              `json:"messType"`
	Opermess  int              `json:"opermess"`
	Timestamp time.Time        `json:"timestamp"`
}*/
