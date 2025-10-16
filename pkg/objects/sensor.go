// objects/sensor.go
package objects

import (
	"server-system/pkg/types"
)

// UpdateSensor основной обработчик для сенсоров
func SensorUpdate(config *types.ObjectConfig, state *types.ObjectStateForVue, tagValue types.TagValue, alias string) {
	// Находим тип регистра по алиасу
	registerType := GetRegisterTypeByAlias(config, alias)
	if registerType == "" {
		return
	}

	// Обновляем AliasVal в конфиге
	//UpdateAliasVal(config, registerType, tagValue.Value)
	if state.ObjVue.Alias == nil {
		state.ObjVue.Alias = make(map[string]interface{})
	}
	state.ObjVue.Alias[string(registerType)] = tagValue.Value

	// Выбираем логику обработки
	switch registerType {
	case RegisterValue:
		sensorUpdateInputValue(config, state, tagValue)
	case RegisterError:
		sensorUpdateError(config, state, tagValue)
	case RegisterState:
		sensorUpdateState(config, state, tagValue)
	default:

	}

}

// обновляет значение ввода
func sensorUpdateInputValue(config *types.ObjectConfig, state *types.ObjectStateForVue, tagValue types.TagValue) {
	state.ObjVue.InputValue = FormatValueWithUnit(tagValue.Value, config.Unit)
}

// обновляет поле ошибки
func sensorUpdateError(config *types.ObjectConfig, state *types.ObjectStateForVue, tagValue types.TagValue) {
	state.ObjVue.Error = SafeConvertToUint(tagValue.Value, types.DataTypeUINT16)
}

// обновляет битовые поля состояния
func sensorUpdateState(config *types.ObjectConfig, state *types.ObjectStateForVue, tagValue types.TagValue) {

	stateVal := SafeConvertToUint(tagValue.Value, types.DataTypeUINT16)

	// Разбираем битовые поля
	state.ObjVue.State = stateVal % 16
	state.ObjVue.Mask = (stateVal%64)/32 > 0
	state.ObjVue.Imit = (stateVal%32)/16 > 0
	state.ObjVue.Ack = (stateVal%128)/64 > 0
	state.ObjVue.RealInput = (stateVal%256)/128 > 0
	state.ObjVue.ChainControl = (stateVal%512)/256 > 0

	// Устанавливаем цвет и текст состояния
	sensorUpdateStateColorAndText(config, state)

}

// updateStateColorAndText устанавливает цвет и текст состояния
func sensorUpdateStateColorAndText(config *types.ObjectConfig, state *types.ObjectStateForVue) {
	stateValue := state.ObjVue.State

	switch stateValue {
	case 1:
		state.ObjVue.StateColor = config.State.ColorOff
		state.ObjVue.StateTxt = config.State.TxtOff
	case 3:
		state.ObjVue.StateColor = config.State.ColorOn
		state.ObjVue.StateTxt = config.State.TxtOn
	default:
		if color, exists := sensorArrStateColors[stateValue]; exists {
			state.ObjVue.StateColor = color
		} else {
			state.ObjVue.StateColor = sensorArrStateColors[0]
		}

		if text, exists := sensorArrStateTexts[stateValue]; exists {
			state.ObjVue.StateTxt = text
		} else {
			state.ObjVue.StateTxt = sensorArrStateTexts[0]
		}
	}
}

// Массивы для состояний
var sensorArrStateColors = map[uint]string{
	0: "#C0C0C0", // Silver
	1: "#FF0000", // Lime
	2: "#FFFF00", // Yellow
	3: "#FF0000", // Red
	4: "#FF00FF", // Fuchsia
}

var sensorArrStateTexts = map[uint]string{
	0: "НЕДОСТОВЕРНОСТЬ",
	1: "ДЕЖУРСТВО",
	2: "ОШИБКА",
	3: "ПОЖАР",
	4: "ВНИМАНИЕ",
}
