// objects/sensor.go
package objects

import (
	"server-system/pkg/types"
)

// UpdateSensor основной обработчик для сенсоров
func SensorUpdate(config *types.ObjectConfig, stateInterface interface{}, tagValue types.TagValue, alias string) {
	// Находим тип регистра по алиасу
	registerType := GetRegisterTypeByAlias(config, alias)
	if registerType == "" {
		return
	}

	// Преобразуем к конкретному типу
	state, ok := stateInterface.(*VueObjectSensorsState)
	if !ok {
		return
	}

	// Выбираем логику обработки
	switch registerType {
	case RegisterValue:
		sensorUpdateInputValue(config, state, tagValue)
	case RegisterError:
		sensorUpdateError(state, tagValue)
	case RegisterState:
		sensorUpdateState(config, state, tagValue)
	default:

	}

}

// обновляет значение ввода
func sensorUpdateInputValue(config *types.ObjectConfig, state *VueObjectSensorsState, tagValue types.TagValue) {
	state.RawRegInput = tagValue.Value
	state.InputValue = FormatValueWithUnit(tagValue.Value, config.Unit)
}

// обновляет поле ошибки
func sensorUpdateError(state *VueObjectSensorsState, tagValue types.TagValue) {
	state.RawRegError = tagValue.Value
	state.Error = SafeConvertToUint(tagValue.Value, types.DataTypeUINT16)
}

// обновляет битовые поля состояния
func sensorUpdateState(config *types.ObjectConfig, state *VueObjectSensorsState, tagValue types.TagValue) {
	state.RawRegState = tagValue.Value
	stateVal := SafeConvertToUint(tagValue.Value, types.DataTypeUINT16)

	// Разбираем битовые поля
	state.State = stateVal % 16
	state.Mask = (stateVal%64)/32 > 0
	state.Imit = (stateVal%32)/16 > 0
	state.Ack = (stateVal%128)/64 > 0
	state.RealInput = (stateVal%256)/128 > 0

	// Устанавливаем цвет и текст состояния
	sensorUpdateStateColorAndText(config, state)

}

// updateStateColorAndText устанавливает цвет и текст состояния
func sensorUpdateStateColorAndText(config *types.ObjectConfig, state *VueObjectSensorsState) {
	stateValue := state.State

	switch stateValue {
	case 1:
		state.StateColor = config.State.ColorOff
		state.StateTxt = config.State.TxtOff
	case 3:
		state.StateColor = config.State.ColorOn
		state.StateTxt = config.State.TxtOn
	default:
		if color, exists := sensorArrStateColors[stateValue]; exists {
			state.StateColor = color
		} else {
			state.StateColor = sensorArrStateColors[0]
		}

		if text, exists := sensorArrStateTexts[stateValue]; exists {
			state.StateTxt = text
		} else {
			state.StateTxt = sensorArrStateTexts[0]
		}
	}
}

type VueObjectSensorsState struct {
	StateColor  string      `json:"stateColor"`
	StateTxt    string      `json:"stateTxt"`
	State       uint        `json:"state"`
	Mask        bool        `json:"mask"`
	Imit        bool        `json:"imit"`
	Ack         bool        `json:"ack"`
	RealInput   bool        `json:"realInput"`
	InputValue  string      `json:"inputValue"`
	Error       uint        `json:"error"`
	RawRegState interface{} `json:"rawRegState"`
	RawRegError interface{} `json:"rawRegError"`
	RawRegInput interface{} `json:"rawRegInput"`
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
