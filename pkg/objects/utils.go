// objects/common.go
package objects

import (
	"fmt"
	"server-system/pkg/types"
	"strconv"
	"time"
)

// === Alias ================================================================================

// GetRegisterTypeByAlias находит тип регистра по алиасу
func GetRegisterTypeByAlias(config *types.ObjectConfig, alias string) RegisterType {
	for regType, regAlias := range config.Alias {
		if regAlias == alias {
			return RegisterType(regType)
		}
	}
	return ""
}

/*
// UpdateAliasVal обновляет значение в AliasVal
func UpdateAliasVal(config *types.ObjectConfig, registerType RegisterType, value interface{}) {
	if config.AliasVal == nil {
		config.AliasVal = make(map[string]interface{})
	}
	config.AliasVal[string(registerType)] = value
}*/

// === MESSAGE ================================================================================

// Общая функция для обработки полей состояния (не битовых)
func processStateField(alarmMess *[]types.AlarmMessDBType, config *types.ObjectConfig, oldValue, newValue uint, mask int, messMap map[uint]MessInfo, timestamp time.Time) {
	// Проверяем маску для этого поля
	if mask != 0 {
		return
	}

	if oldValue != newValue {
		messInfo, exists := messMap[newValue]
		if exists && messInfo.MessTxtState0 != "" {
			message := types.AlarmMessDBType{
				IdObj:     config.ID,
				TypeObj:   config.Type,
				Tag:       config.Info.Tag,
				UsoID:     int(config.Uso.ID),
				UsoTxt:    config.Uso.Txt,
				Color:     messInfo.MessColor0,
				MessFull:  config.Info.Desc + ": " + messInfo.MessTxtState0,
				MessName:  config.Info.Desc,
				MessState: messInfo.MessTxtState0,
				Severity:  messInfo.MessType0,
				Opermess:  config.Alarm["opermess"],
				Code:      777777,
				Users:     "test_user",
				Dt:        int64(timestamp.Nanosecond()),
				DtTxt:     TimeToPostgresFormat(timestamp),
			}
			*alarmMess = append(*alarmMess, message)
		}
	}
}

// Общая функция для обработки битовых полей
func processStateBitField(alarmMess *[]types.AlarmMessDBType, config *types.ObjectConfig, oldState, newState uint, mask int, bitPos uint, messMap map[uint]MessInfo, timestamp time.Time) {
	// Проверяем маску для этого бита
	bitMask := 1 << bitPos
	if mask&bitMask != 0 {
		return
	}

	oldBitValue := (oldState >> bitPos) & 1
	newBitValue := (newState >> bitPos) & 1

	// Если значение бита изменилось
	if oldBitValue != newBitValue {
		messInfo, exists := messMap[bitPos]
		if !exists {
			return
		}

		var messColor string
		var messText string
		var messType int

		if newBitValue == 1 {
			messColor = messInfo.MessColor1
			messText = messInfo.MessTxtState1
			messType = messInfo.MessType1
		} else {
			messColor = messInfo.MessColor0
			messText = messInfo.MessTxtState0
			messType = messInfo.MessType0
		}

		if messText != "" {
			message := types.AlarmMessDBType{
				IdObj:     config.ID,
				TypeObj:   config.Type,
				Tag:       config.Info.Tag,
				UsoID:     int(config.Uso.ID),
				UsoTxt:    config.Uso.Txt,
				Color:     messColor,
				MessFull:  config.Info.Desc + ": " + messText,
				MessName:  config.Info.Desc,
				MessState: messText,
				Severity:  messType,
				Opermess:  config.Alarm["opermess"],
				Code:      777777,
				Users:     "test_user",
				Dt:        int64(timestamp.Nanosecond()),
				DtTxt:     TimeToPostgresFormat(timestamp),
			}
			*alarmMess = append(*alarmMess, message)
		}
	}
}

// === Format Time ============================================================================

// TimeToPostgresFormat конвертирует time.Time в формат DD.MM.YYYY HH24:MI:SS.MS
func TimeToPostgresFormat(t time.Time) string {
	return t.Format("02.01.2006 15:04:05.000")
}

/*
	Millisecond: t.Nanosecond() / 1e6,
    Microsecond: t.Nanosecond() / 1e3,
    Nanosecond:  t.Nanosecond(),
*/

// === Format Data ============================================================================

// FormatValue форматирует значение согласно Unit конфигу
func FormatValue(value interface{}, unit types.ObjectUnitConfig) string {
	if value == nil {
		return ""
	}

	if unit.Format != "" {
		floatVal, err := toFloat64(value)
		if err == nil {
			return fmt.Sprintf(unit.Format, floatVal)
		}
	}

	return fmt.Sprintf("%v", value)
}

// FormatValueWithUnit форматирует значение с добавлением единиц измерения
func FormatValueWithUnit(value interface{}, unit types.ObjectUnitConfig) string {
	formatted := FormatValue(value, unit)
	if formatted == "" {
		return ""
	}

	if unit.Txt != "" {
		return fmt.Sprintf("%s %s", formatted, unit.Txt)
	}

	return formatted
}

// toFloat64 конвертирует interface{} в float64
func toFloat64(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case uint:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("unsupported type: %T", value)
	}
}

// ConvertToUint конвертирует interface{} в uint с учетом DataType
func ConvertToUint(value interface{}, dataType types.DataType) (uint, error) {
	switch v := value.(type) {
	case uint:
		return v, nil
	case uint8:
		return uint(v), nil
	case uint16:
		return uint(v), nil
	case uint32:
		return uint(v), nil
	case uint64:
		return uint(v), nil
	case int:
		if v < 0 {
			return 0, fmt.Errorf("negative value: %d", v)
		}
		return uint(v), nil
	case int8:
		if v < 0 {
			return 0, fmt.Errorf("negative value: %d", v)
		}
		return uint(v), nil
	case int16:
		if v < 0 {
			return 0, fmt.Errorf("negative value: %d", v)
		}
		return uint(v), nil
	case int32:
		if v < 0 {
			return 0, fmt.Errorf("negative value: %d", v)
		}
		return uint(v), nil
	case int64:
		if v < 0 {
			return 0, fmt.Errorf("negative value: %d", v)
		}
		return uint(v), nil
	case float32:
		if v < 0 {
			return 0, fmt.Errorf("negative value: %f", v)
		}
		return uint(v), nil
	case float64:
		if v < 0 {
			return 0, fmt.Errorf("negative value: %f", v)
		}
		return uint(v), nil
	case string:
		// Пробуем парсить как число
		if intVal, err := strconv.ParseUint(v, 10, 64); err == nil {
			return uint(intVal), nil
		}
		return 0, fmt.Errorf("cannot parse string as uint: %s", v)
	default:
		return 0, fmt.Errorf("unsupported type for uint conversion: %T", value)
	}
}

// SafeConvertToUint безопасно конвертирует в uint (возвращает 0 при ошибке)
func SafeConvertToUint(value interface{}, dataType types.DataType) uint {
	result, err := ConvertToUint(value, dataType)
	if err != nil {
		return 0
	}
	return result
}

// ConvertByDataType конвертирует значение согласно указанному типу данных
func ConvertByDataType(value interface{}, dataType types.DataType) (interface{}, error) {
	switch dataType {
	case types.DataTypeBOOL:
		return toBool(value)
	case types.DataTypeBYTE:
		return toUint8(value)
	case types.DataTypeINT16:
		return toInt16(value)
	case types.DataTypeUINT16:
		return toUint16(value)
	case types.DataTypeINT32:
		return toInt32(value)
	case types.DataTypeUINT32:
		return toUint32(value)
	case types.DataTypeINT64:
		return toInt64(value)
	case types.DataTypeUINT64:
		return toUint64(value)
	case types.DataTypeFLOAT32:
		return toFloat32(value)
	case types.DataTypeFLOAT64:
		return toFloat64(value)
	default:
		return value, nil // неизвестный тип - возвращаем как есть
	}
}

// Вспомогательные функции конвертации
func toBool(value interface{}) (bool, error) {
	switch v := value.(type) {
	case bool:
		return v, nil
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return v != 0, nil
	case float32, float64:
		return v != 0.0, nil
	case string:
		return strconv.ParseBool(v)
	default:
		return false, fmt.Errorf("unsupported type for bool: %T", value)
	}
}

func toUint8(value interface{}) (uint8, error) {
	u, err := ConvertToUint(value, types.DataTypeUINT8)
	return uint8(u), err
}

func toUint16(value interface{}) (uint16, error) {
	u, err := ConvertToUint(value, types.DataTypeUINT16)
	return uint16(u), err
}

func toUint32(value interface{}) (uint32, error) {
	u, err := ConvertToUint(value, types.DataTypeUINT32)
	return uint32(u), err
}

func toUint64(value interface{}) (uint64, error) {
	u, err := ConvertToUint(value, types.DataTypeUINT64)
	return uint64(u), err
}

func toInt16(value interface{}) (int16, error) {
	f, err := toFloat64(value)
	return int16(f), err
}

func toInt32(value interface{}) (int32, error) {
	f, err := toFloat64(value)
	return int32(f), err
}

func toInt64(value interface{}) (int64, error) {
	f, err := toFloat64(value)
	return int64(f), err
}

func toFloat32(value interface{}) (float32, error) {
	f, err := toFloat64(value)
	return float32(f), err
}
