// objects/sensor.go
package objects

import (
	"server-system/pkg/types"
	"time"
)

// UpdateSensor основной обработчик для сенсоров
func SensorUpdate(config *types.ObjectConfig, alarmMess *[]types.AlarmMessDBType, stateInterface interface{}, tagValue types.TagValue, alias string, oldValue interface{}) {
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
func sensorUpdateError(config *types.ObjectConfig, alarmMess *[]types.AlarmMessDBType, state *VueObjectSensorsState, tagValue types.TagValue, oldValue interface{}) {
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
func sensorErrorMessages(config *types.ObjectConfig, alarmMess *[]types.AlarmMessDBType, oldError, newError uint, timestamp time.Time) {
	errorMask, hasErrorConfig := config.Alarm["error"]
	if !hasErrorConfig {
		return
	}

	if oldError == newError {
		return
	}

	// Выбираем карту в зависимости от ErrType
	var errorBitMessMap map[uint]MessInfo
	switch config.ErrType {
	case 0:
		errorBitMessMap = errorBitMessType0Map
	case 1:
		errorBitMessMap = errorBitMessType1Map
	case 2:
		errorBitMessMap = errorBitMessType2Map
	default:
		errorBitMessMap = errorBitMessType0Map
	}

	// Обрабатываем каждый бит от 0 до 15
	for bit := uint(0); bit < 16; bit++ {
		processStateBitField(alarmMess, config, oldError, newError, errorMask, bit, errorBitMessMap, timestamp)
	}
}

// === STATE ==========================================================

// обновляет битовые поля состояния
func sensorUpdateState(config *types.ObjectConfig, alarmMess *[]types.AlarmMessDBType, state *VueObjectSensorsState, tagValue types.TagValue, oldValue interface{}) {
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
func sensorStateMessages(config *types.ObjectConfig, alarmMess *[]types.AlarmMessDBType, oldState, newState uint, timestamp time.Time) {

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
		if stateInfo, exists := sensorStateMessMap[stateValue]; exists {
			state.StateColor = stateInfo.StateColor
			state.StateTxt = stateInfo.StateTxt
		} else {
			state.StateColor = sensorStateMessMap[0].StateColor
			state.StateTxt = sensorStateMessMap[0].StateTxt
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

// Мапы для хранения сообщений о состояниях
var sensorStateMessMap = map[uint]MessInfo{
	0: {MessTxtState0: "НЕДОСТОВЕРНОСТЬ",
		MessColor0: cStUnrel,
		MessType0:  tStUnrel,
		StateTxt:   "НЕДОСТОВЕРНОСТЬ",
		StateColor: cStUnrel},
	1: {MessTxtState0: "ДЕЖУРСТВО",
		MessColor0: cStNorma,
		MessType0:  tStNorma,
		StateTxt:   "ДЕЖУРСТВО",
		StateColor: cStNorma},
	2: {MessTxtState0: "НЕИСПРАВНОСТЬ",
		MessColor0: cMessErr,
		MessType0:  tStErr,
		StateTxt:   "НЕИСПРАВНОСТЬ",
		StateColor: cStErr},
	3: {MessTxtState0: "ПОЖАР",
		MessColor0: cStFire,
		MessType0:  tStFire,
		StateTxt:   "ПОЖАР",
		StateColor: cStFire},
	4: {MessTxtState0: "ВНИМАНИЕ",
		MessColor0: cStAtent,
		MessType0:  tStAtent,
		StateTxt:   "ВНИМАНИЕ",
		StateColor: cStAtent},
}

var sensorStateBitMessMap = map[uint]MessInfo{
	4: {MessTxtState0: "ИМИТАЦИЯ СНЯТА", MessColor0: cBlack, MessType0: tImitOff,
		MessTxtState1: "ИМИТАЦИЯ УСТАНОВЛЕНА", MessColor1: cMessImitOn, MessType1: tImitOn},
	5: {MessTxtState0: "МАСКА СНЯТА", MessColor0: cBlack, MessType0: tMaskOff,
		MessTxtState1: "МАСКА УСТАНОВЛЕНА", MessColor1: cMessMaskOn, MessType1: tMaskOn},
	6: {MessTxtState0: "КВИТИРОВАНИЕ СНЯТО", MessColor0: cBlack, MessType0: tAckOff,
		MessTxtState1: "КВИТИРОВАНИЕ УСТАНОВЛЕНО", MessColor1: cMessAckOn, MessType1: tAckOn},
	7: {MessTxtState0: "РЕАЛЬНЫЙ СИГНАЛ СНЯТ", MessColor0: cBlack, MessType0: tRealOff,
		MessTxtState1: "РЕАЛЬНЫЙ СИГНАЛ УСТАНОВЛЕН", MessColor1: cBlack, MessType1: tRealOn},
}
