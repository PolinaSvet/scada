// objects/types.go
package objects

import "server-system/pkg/types"

// ObjectType типы объектов
type ObjectType string

const (
	TypeSensor ObjectType = "sensor"
	TypeDi     ObjectType = "di"
	// добавить другие типы
)

// RegisterType типы регистров
type RegisterType string

const (
	RegisterValue RegisterType = "value"
	RegisterError RegisterType = "error"
	RegisterState RegisterType = "state"
)

var Handlers = map[ObjectType]func(config *types.ObjectConfig, state interface{}, tagValue types.TagValue, alias string){
	TypeSensor: SensorUpdate,
	//TypeDI:     DIUpdate,
}

// Handlers мапа обработчиков
//var Handlers = map[ObjectType]func(config *types.ObjectConfig, state *types.ObjectStateForVue, tagValue types.TagValue, alias string){
//	TypeSensor: SensorUpdate,
//	//TypeDI:     DiUpdate,
//}

type VueObjectDiState struct {
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
}
