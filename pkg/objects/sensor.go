// objects/sensor.go
package objects

import (
	"server-system/pkg/types"
	"time"
)

/*
// UpdateSensor основной обработчик для сенсоров
func SensorUpdate(config *types.ObjectConfig, stateInterface interface{}, tagValue types.TagValue, alias string, oldValue interface{}) {
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
		sensorUpdateInputValue(config, state, tagValue, oldValue)
	case RegisterError:
		sensorUpdateError(state, tagValue, oldValue)
	case RegisterState:
		sensorUpdateState(config, state, tagValue, oldValue)
	default:

	}

}

// обновляет значение ввода
func sensorUpdateInputValue(config *types.ObjectConfig, state *VueObjectSensorsState, tagValue types.TagValue, oldValue interface{}) {
	state.RawRegInput = tagValue.Value
	state.InputValue = FormatValueWithUnit(tagValue.Value, config.Unit)
}

// обновляет поле ошибки
func sensorUpdateError(state *VueObjectSensorsState, tagValue types.TagValue, oldValue interface{}) {
	state.RawRegError = tagValue.Value
	state.Error = SafeConvertToUint(tagValue.Value, types.DataTypeUINT16)
}

// обновляет битовые поля состояния
func sensorUpdateState(config *types.ObjectConfig, state *VueObjectSensorsState, tagValue types.TagValue, oldValue interface{}) {
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

}*/

// UpdateSensor основной обработчик для сенсоров
func SensorUpdate(config *types.ObjectConfig, alarmMess *[]types.AlarmMess, stateInterface interface{}, tagValue types.TagValue, alias string, oldValue interface{}) {
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
		sensorUpdateInputValue(config, state, tagValue, oldValue)
	case RegisterError:
		sensorUpdateError(config, alarmMess, state, tagValue, oldValue)
	case RegisterState:
		sensorUpdateState(config, alarmMess, state, tagValue, oldValue)
	}
}

// === INPUT ==========================================================

// обновляет значение ввода
func sensorUpdateInputValue(config *types.ObjectConfig, state *VueObjectSensorsState, tagValue types.TagValue, oldValue interface{}) {
	state.RawRegInput = tagValue.Value
	state.InputValue = FormatValueWithUnit(tagValue.Value, config.Unit)
}

// === ERROR ==========================================================

// обновляет поле ошибки
func sensorUpdateError(config *types.ObjectConfig, alarmMess *[]types.AlarmMess, state *VueObjectSensorsState, tagValue types.TagValue, oldValue interface{}) {
	oldError := SafeConvertToUint(oldValue, types.DataTypeUINT16)
	newError := SafeConvertToUint(tagValue.Value, types.DataTypeUINT16)

	state.RawRegError = tagValue.Value
	state.Error = newError

	// Формируем сообщения если включена обработка аварий
	if config.Alarm != nil && config.Alarm["enable"] == 1 {
		sensorErrorMessages(config, alarmMess, oldError, newError, tagValue.Timestamp)
	}
}

// Обрабатывает сообщения об ошибках
func sensorErrorMessages(config *types.ObjectConfig, alarmMess *[]types.AlarmMess, oldError, newError uint, timestamp time.Time) {
	errorMask, hasErrorConfig := config.Alarm["error"]
	if !hasErrorConfig {
		return
	}

	if oldError == newError {
		return
	}

	// Обрабатываем каждый бит от 0 до 15
	for bit := uint(0); bit < 16; bit++ {
		processStateBitField(alarmMess, config, oldError, newError, errorMask, bit, sensorErrorBitMessMap, timestamp)
	}
}

// === STATE ==========================================================

// обновляет битовые поля состояния
func sensorUpdateState(config *types.ObjectConfig, alarmMess *[]types.AlarmMess, state *VueObjectSensorsState, tagValue types.TagValue, oldValue interface{}) {
	oldState := SafeConvertToUint(oldValue, types.DataTypeUINT16)
	newState := SafeConvertToUint(tagValue.Value, types.DataTypeUINT16)

	state.RawRegState = tagValue.Value
	stateVal := newState

	// Разбираем битовые поля
	state.State = stateVal % 16
	state.Mask = (stateVal%64)/32 > 0
	state.Imit = (stateVal%32)/16 > 0
	state.Ack = (stateVal%128)/64 > 0
	state.RealInput = (stateVal%256)/128 > 0

	// Устанавливаем цвет и текст состояния
	sensorUpdateStateColorAndText(config, state)

	// Формируем сообщения если включена обработка аварий
	if config.Alarm != nil && config.Alarm["enable"] == 1 {
		sensorStateMessages(config, alarmMess, oldState, newState, tagValue.Timestamp)
	}
}

// Обрабатывает сообщения о состояниях
func sensorStateMessages(config *types.ObjectConfig, alarmMess *[]types.AlarmMess, oldState, newState uint, timestamp time.Time) {

	stateMask, hasStateConfig := config.Alarm["state"]
	if !hasStateConfig {
		return
	}

	if oldState == newState {
		return
	}

	// Обработка основного состояния
	processStateField(alarmMess, config, oldState%16, newState%16, stateMask%16, sensorStateMessMap, timestamp)

	// Обработка битовых полей состояния
	processStateBitField(alarmMess, config, oldState, newState, stateMask, 4, sensorStateBitMessMap, timestamp)
	processStateBitField(alarmMess, config, oldState, newState, stateMask, 5, sensorStateBitMessMap, timestamp)
	processStateBitField(alarmMess, config, oldState, newState, stateMask, 6, sensorStateBitMessMap, timestamp)
	processStateBitField(alarmMess, config, oldState, newState, stateMask, 7, sensorStateBitMessMap, timestamp)
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
		if stateInfo, exists := sensorStateInfo[stateValue]; exists {
			state.StateColor = stateInfo.Color
			state.StateTxt = stateInfo.Text
		} else {
			state.StateColor = sensorStateInfo[0].Color
			state.StateTxt = sensorStateInfo[0].Text
		}
	}
}

// === STRUCT ==========================================================

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

// === MAP ==========================================================

// Объединенная мапа для состояний
var sensorStateInfo = map[uint]StateInfo{
	0: {"#C0C0C0", "НЕДОСТОВЕРНОСТЬ"},
	1: {"#FF0000", "ДЕЖУРСТВО"},
	2: {"#FFFF00", "НЕИСПРАВНОСТЬ"},
	3: {"#FF0000", "ПОЖАР"},
	4: {"#FF00FF", "ВНИМАНИЕ"},
}

// Мапы для хранения сообщений об ошибках
var sensorErrorBitMessMap = map[uint]MessInfo{
	0:  {MessTxtState0: "Ошибка бит 0 снята", MessTxtState1: "Ошибка бит 0 установлена", MessColor0: "#000000", MessColor1: "#FFFF00", MessType0: 0, MessType1: 1},
	1:  {MessTxtState0: "Ошибка бит 1 снята", MessTxtState1: "Ошибка бит 1 установлена", MessColor0: "#000000", MessColor1: "#FFFF00", MessType0: 0, MessType1: 1},
	2:  {MessTxtState0: "Ошибка бит 2 снята", MessTxtState1: "Ошибка бит 2 установлена", MessColor0: "#000000", MessColor1: "#FFFF00", MessType0: 0, MessType1: 1},
	3:  {MessTxtState0: "Ошибка бит 3 снята", MessTxtState1: "Ошибка бит 3 установлена", MessColor0: "#000000", MessColor1: "#FFFF00", MessType0: 0, MessType1: 1},
	4:  {MessTxtState0: "Ошибка бит 4 снята", MessTxtState1: "Ошибка бит 4 установлена", MessColor0: "#000000", MessColor1: "#FFFF00", MessType0: 0, MessType1: 1},
	5:  {MessTxtState0: "Ошибка бит 5 снята", MessTxtState1: "Ошибка бит 5 установлена", MessColor0: "#000000", MessColor1: "#FFFF00", MessType0: 0, MessType1: 1},
	6:  {MessTxtState0: "Ошибка бит 6 снята", MessTxtState1: "Ошибка бит 6 установлена", MessColor0: "#000000", MessColor1: "#FFFF00", MessType0: 0, MessType1: 1},
	7:  {MessTxtState0: "Ошибка бит 7 снята", MessTxtState1: "Ошибка бит 7 установлена", MessColor0: "#000000", MessColor1: "#FFFF00", MessType0: 0, MessType1: 1},
	8:  {MessTxtState0: "Ошибка бит 8 снята", MessTxtState1: "Ошибка бит 8 установлена", MessColor0: "#000000", MessColor1: "#FFFF00", MessType0: 0, MessType1: 1},
	9:  {MessTxtState0: "Ошибка бит 9 снята", MessTxtState1: "Ошибка бит 9 установлена", MessColor0: "#000000", MessColor1: "#FFFF00", MessType0: 0, MessType1: 1},
	10: {MessTxtState0: "Ошибка бит 10 снята", MessTxtState1: "Ошибка бит 10 установлена", MessColor0: "#000000", MessColor1: "#FFFF00", MessType0: 0, MessType1: 1},
	11: {MessTxtState0: "Ошибка бит 11 снята", MessTxtState1: "Ошибка бит 11 установлена", MessColor0: "#000000", MessColor1: "#FFFF00", MessType0: 0, MessType1: 1},
	12: {MessTxtState0: "Ошибка бит 12 снята", MessTxtState1: "Ошибка бит 12 установлена", MessColor0: "#000000", MessColor1: "#FFFF00", MessType0: 0, MessType1: 1},
	13: {MessTxtState0: "Ошибка бит 13 снята", MessTxtState1: "Ошибка бит 13 установлена", MessColor0: "#000000", MessColor1: "#FFFF00", MessType0: 0, MessType1: 1},
	14: {MessTxtState0: "Ошибка бит 14 снята", MessTxtState1: "Ошибка бит 14 установлена", MessColor0: "#000000", MessColor1: "#FFFF00", MessType0: 0, MessType1: 1},
	15: {MessTxtState0: "Ошибка бит 15 снята", MessTxtState1: "Ошибка бит 15 установлена", MessColor0: "#000000", MessColor1: "#FFFF00", MessType0: 0, MessType1: 1},
}

// Мапы для хранения сообщений о состояниях
var sensorStateMessMap = map[uint]MessInfo{
	0: {MessTxtState0: "НЕДОСТОВЕРНОСТЬ", MessColor0: "#C0C0C0", MessType0: 0},
	1: {MessTxtState0: "ДЕЖУРСТВО", MessColor0: "#FF0000", MessType0: 0},
	2: {MessTxtState0: "НЕИСПРАВНОСТЬ", MessColor0: "#FFFF00", MessType0: 0},
	3: {MessTxtState0: "ПОЖАР", MessColor0: "#FF0000", MessType0: 0},
	4: {MessTxtState0: "ВНИМАНИЕ", MessColor0: "#FF00FF", MessType0: 0},
}

var sensorStateBitMessMap = map[uint]MessInfo{
	4: {MessTxtState0: "Имитация снята", MessTxtState1: "Имитация установлена", MessColor0: "#000000", MessColor1: "#FFA500", MessType0: 0, MessType1: 3},
	5: {MessTxtState0: "Маска снята", MessTxtState1: "Маска установлена", MessColor0: "#000000", MessColor1: "#808080", MessType0: 0, MessType1: 4},
	6: {MessTxtState0: "Квитирование снято", MessTxtState1: "Квитирование установлено", MessColor0: "#000000", MessColor1: "#008000", MessType0: 0, MessType1: 5},
	7: {MessTxtState0: "Реальный ввод снят", MessTxtState1: "Реальный ввод установлен", MessColor0: "#000000", MessColor1: "#0000FF", MessType0: 0, MessType1: 6},
}
