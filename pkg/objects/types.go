// objects/types.go
package objects

import "server-system/pkg/types"

// ObjectType типы объектов
type ObjectType string

const (
	TypeSensor ObjectType = "sensor"
	TypeDI     ObjectType = "di"
	// добавить другие типы
)

// RegisterType типы регистров
type RegisterType string

const (
	RegisterValue RegisterType = "value"
	RegisterError RegisterType = "error"
	RegisterState RegisterType = "state"
)

// Handlers мапа обработчиков
var Handlers = map[ObjectType]func(config *types.ObjectConfig, state *types.ObjectStateForVue, tagValue types.TagValue, alias string){
	TypeSensor: SensorUpdate,
	//TypeDI:     DiUpdate,
}
