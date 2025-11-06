// objects/sensor.go
package objects

//color scheme
const (
	cBlack = "#000000"

	cMessErr    = "#B8860B"
	cMessImitOn = "#800080"
	cMessMaskOn = "#483D8B"
	cMessAckOn  = "#4B0082"

	cStUnrel = "#C0C0C0"
	cStNorma = "#00FF00"
	cStErr   = "#FFFF00"
	cStFire  = "#FF0000"
	cStAtent = "#FF00FF"
)

//type scheme
const (
	tNone = 0

	tErrOff = 900
	tErrOn  = 901

	tStUnrel = 801
	tStNorma = 101
	tStErr   = 901
	tStFire  = 1001
	tStAtent = 1101

	tImitOff = 3000
	tImitOn  = 3001
	tMaskOff = 3010
	tMaskOn  = 3011
	tAckOff  = 3020
	tAckOn   = 3021
	tRealOff = 3030
	tRealOn  = 3031
)

// === MAP ==========================================================

// Мапы для хранения сообщений об ошибках
var errorBitMessType0Map = map[uint]MessInfo{
	0: {MessTxtState0: "НЕТ ЗНАЧЕНИЯ: ОШИБКА СНЯТА",
		MessTxtState1: "НЕТ ЗНАЧЕНИЯ: ОШИБКА УСТАНОВЛЕНА",
		MessColor0:    cBlack, MessColor1: cMessErr, MessType0: tErrOff, MessType1: tErrOn},
	1: {MessTxtState0: "КОРОТКОЕ ЗАМЫКАНИЕ ЛИНИИ: ОШИБКА СНЯТА",
		MessTxtState1: "КОРОТКОЕ ЗАМЫКАНИЕ ЛИНИИ: ОШИБКА УСТАНОВЛЕНА",
		MessColor0:    cBlack, MessColor1: cMessErr, MessType0: tErrOff, MessType1: tErrOn},
	2: {MessTxtState0: "ОБРЫВ ЛИНИИ/ОТСУТСТВИЕ ПИТАНИЯ: ОШИБКА СНЯТА",
		MessTxtState1: "ОБРЫВ ЛИНИИ/ОТСУТСТВИЕ ПИТАНИЯ: ОШИБКА УСТАНОВЛЕНА",
		MessColor0:    cBlack, MessColor1: cMessErr, MessType0: tErrOff, MessType1: tErrOn},
	3: {MessTxtState0: "НЕТ СВЯЗИ: ОШИБКА СНЯТА",
		MessTxtState1: "НЕТ СВЯЗИ: ОШИБКА УСТАНОВЛЕНА",
		MessColor0:    cBlack, MessColor1: cMessErr, MessType0: tErrOff, MessType1: tErrOn},
	4: {MessTxtState0: "НЕ НАЙДЕНО: ОШИБКА СНЯТА",
		MessTxtState1: "НЕ НАЙДЕНО: ОШИБКА УСТАНОВЛЕНА",
		MessColor0:    cBlack, MessColor1: cMessErr, MessType0: tErrOff, MessType1: tErrOn},
	5: {MessTxtState0: "НИЖНИЙ ПРЕДЕЛ: ОШИБКА СНЯТА",
		MessTxtState1: "НИЖНИЙ ПРЕДЕЛ: ОШИБКА УСТАНОВЛЕНА",
		MessColor0:    cBlack, MessColor1: cMessErr, MessType0: tErrOff, MessType1: tErrOn},
	6: {MessTxtState0: "ВЕРХНИЙ ПРЕДЕЛ: ОШИБКА СНЯТА",
		MessTxtState1: "ВЕРХНИЙ ПРЕДЕЛ: ОШИБКА УСТАНОВЛЕНА",
		MessColor0:    cBlack, MessColor1: cMessErr, MessType0: tErrOff, MessType1: tErrOn},
	7: {MessTxtState0: "НЕИСПРАВНОСТЬ: ОШИБКА СНЯТА",
		MessTxtState1: "НЕИСПРАВНОСТЬ: ОШИБКА УСТАНОВЛЕНА",
		MessColor0:    cBlack, MessColor1: cMessErr, MessType0: tErrOff, MessType1: tErrOn},
	8: {MessTxtState0: "ПРОЧИЕ НЕИСПРАВНОСТИ: ОШИБКА СНЯТА",
		MessTxtState1: "ПРОЧИЕ НЕИСПРАВНОСТИ: ОШИБКА УСТАНОВЛЕНА",
		MessColor0:    cBlack, MessColor1: cMessErr, MessType0: tErrOff, MessType1: tErrOn},
	9: {MessTxtState0: "БИТ 9: ОШИБКА СНЯТА",
		MessTxtState1: "БИТ 9: ОШИБКА УСТАНОВЛЕНА",
		MessColor0:    cBlack, MessColor1: cMessErr, MessType0: tErrOff, MessType1: tErrOn},
	10: {MessTxtState0: "БИТ 10: ОШИБКА СНЯТА",
		MessTxtState1: "БИТ 10: ОШИБКА УСТАНОВЛЕНА",
		MessColor0:    cBlack, MessColor1: cMessErr, MessType0: tErrOff, MessType1: tErrOn},
	11: {MessTxtState0: "БИТ 11: ОШИБКА СНЯТА",
		MessTxtState1: "БИТ 11: ОШИБКА УСТАНОВЛЕНА",
		MessColor0:    cBlack, MessColor1: cMessErr, MessType0: tErrOff, MessType1: tErrOn},
	12: {MessTxtState0: "БИТ 12: ОШИБКА СНЯТА",
		MessTxtState1: "БИТ 12: ОШИБКА УСТАНОВЛЕНА",
		MessColor0:    cBlack, MessColor1: cMessErr, MessType0: tErrOff, MessType1: tErrOn},
	13: {MessTxtState0: "БИТ 13: ОШИБКА СНЯТА",
		MessTxtState1: "БИТ 13: ОШИБКА УСТАНОВЛЕНА",
		MessColor0:    cBlack, MessColor1: cMessErr, MessType0: tErrOff, MessType1: tErrOn},
	14: {MessTxtState0: "БИТ 14: ОШИБКА СНЯТА",
		MessTxtState1: "БИТ 14: ОШИБКА УСТАНОВЛЕНА",
		MessColor0:    cBlack, MessColor1: cMessErr, MessType0: tErrOff, MessType1: tErrOn},
	15: {MessTxtState0: "БИТ 15: ОШИБКА СНЯТА",
		MessTxtState1: "БИТ 15: ОШИБКА УСТАНОВЛЕНА",
		MessColor0:    cBlack, MessColor1: cMessErr, MessType0: tErrOff, MessType1: tErrOn},
}

var errorBitMessType1Map = map[uint]MessInfo{
	0: {MessTxtState0: "НЕТ ЗНАЧЕНИЯ: ОШИБКА СНЯТА",
		MessTxtState1: "НЕТ ЗНАЧЕНИЯ: ОШИБКА УСТАНОВЛЕНА",
		MessColor0:    cBlack, MessColor1: cMessErr, MessType0: tErrOff, MessType1: tErrOn},
	1: {MessTxtState0: "КОРОТКОЕ ЗАМЫКАНИЕ ЛИНИИ: ОШИБКА СНЯТА",
		MessTxtState1: "КОРОТКОЕ ЗАМЫКАНИЕ ЛИНИИ: ОШИБКА УСТАНОВЛЕНА",
		MessColor0:    cBlack, MessColor1: cMessErr, MessType0: tErrOff, MessType1: tErrOn},
	2: {MessTxtState0: "ОБРЫВ ЛИНИИ/ОТСУТСТВИЕ ПИТАНИЯ: ОШИБКА СНЯТА",
		MessTxtState1: "ОБРЫВ ЛИНИИ/ОТСУТСТВИЕ ПИТАНИЯ: ОШИБКА УСТАНОВЛЕНА",
		MessColor0:    cBlack, MessColor1: cMessErr, MessType0: tErrOff, MessType1: tErrOn},
	3: {MessTxtState0: "НЕТ СВЯЗИ: ОШИБКА СНЯТА",
		MessTxtState1: "НЕТ СВЯЗИ: ОШИБКА УСТАНОВЛЕНА",
		MessColor0:    cBlack, MessColor1: cMessErr, MessType0: tErrOff, MessType1: tErrOn},
	4: {MessTxtState0: "НЕ НАЙДЕНО: ОШИБКА СНЯТА",
		MessTxtState1: "НЕ НАЙДЕНО: ОШИБКА УСТАНОВЛЕНА",
		MessColor0:    cBlack, MessColor1: cMessErr, MessType0: tErrOff, MessType1: tErrOn},
	5: {MessTxtState0: "НИЖНИЙ ПРЕДЕЛ: ОШИБКА СНЯТА",
		MessTxtState1: "НИЖНИЙ ПРЕДЕЛ: ОШИБКА УСТАНОВЛЕНА",
		MessColor0:    cBlack, MessColor1: cMessErr, MessType0: tErrOff, MessType1: tErrOn},
	6: {MessTxtState0: "ВЕРХНИЙ ПРЕДЕЛ: ОШИБКА СНЯТА",
		MessTxtState1: "ВЕРХНИЙ ПРЕДЕЛ: ОШИБКА УСТАНОВЛЕНА",
		MessColor0:    cBlack, MessColor1: cMessErr, MessType0: tErrOff, MessType1: tErrOn},
	7: {MessTxtState0: "НЕИСПРАВНОСТЬ: ОШИБКА СНЯТА",
		MessTxtState1: "НЕИСПРАВНОСТЬ: ОШИБКА УСТАНОВЛЕНА",
		MessColor0:    cBlack, MessColor1: cMessErr, MessType0: tErrOff, MessType1: tErrOn},
	8: {MessTxtState0: "ПРОЧИЕ НЕИСПРАВНОСТИ: ОШИБКА СНЯТА",
		MessTxtState1: "ПРОЧИЕ НЕИСПРАВНОСТИ: ОШИБКА УСТАНОВЛЕНА",
		MessColor0:    cBlack, MessColor1: cMessErr, MessType0: tErrOff, MessType1: tErrOn},
	9: {MessTxtState0: "БИТ 9: ОШИБКА СНЯТА",
		MessTxtState1: "БИТ 9: ОШИБКА УСТАНОВЛЕНА",
		MessColor0:    cBlack, MessColor1: cMessErr, MessType0: tErrOff, MessType1: tErrOn},
	10: {MessTxtState0: "НЕИСПРАВНОСТЬ УФ КОЛБЫ / УФ ЗАСВЕТКА: ОШИБКА СНЯТА",
		MessTxtState1: "НЕИСПРАВНОСТЬ УФ КОЛБЫ / УФ ЗАСВЕТКА: ОШИБКА УСТАНОВЛЕНА",
		MessColor0:    cBlack, MessColor1: cMessErr, MessType0: tErrOff, MessType1: tErrOn},
	11: {MessTxtState0: "НЕИСПРАВНОСТЬ ТРАКТА ПОЖАР: ОШИБКА СНЯТА",
		MessTxtState1: "НЕИСПРАВНОСТЬ ТРАКТА ПОЖАР: ОШИБКА УСТАНОВЛЕНА",
		MessColor0:    cBlack, MessColor1: cMessErr, MessType0: tErrOff, MessType1: tErrOn},
	12: {MessTxtState0: "ЗАГРЯЗНЕНИЕ СТЕКЛА: ОШИБКА СНЯТА",
		MessTxtState1: "ЗАГРЯЗНЕНИЕ СТЕКЛА: ОШИБКА УСТАНОВЛЕНА",
		MessColor0:    cBlack, MessColor1: cMessErr, MessType0: tErrOff, MessType1: tErrOn},
	13: {MessTxtState0: "БИТ 13: ОШИБКА СНЯТА",
		MessTxtState1: "БИТ 13: ОШИБКА УСТАНОВЛЕНА",
		MessColor0:    cBlack, MessColor1: cMessErr, MessType0: tErrOff, MessType1: tErrOn},
	14: {MessTxtState0: "БИТ 14: ОШИБКА СНЯТА",
		MessTxtState1: "БИТ 14: ОШИБКА УСТАНОВЛЕНА",
		MessColor0:    cBlack, MessColor1: cMessErr, MessType0: tErrOff, MessType1: tErrOn},
	15: {MessTxtState0: "БИТ 15: ОШИБКА СНЯТА",
		MessTxtState1: "БИТ 15: ОШИБКА УСТАНОВЛЕНА",
		MessColor0:    cBlack, MessColor1: cMessErr, MessType0: tErrOff, MessType1: tErrOn},
}

var errorBitMessType2Map = map[uint]MessInfo{
	0: {MessTxtState0: "НЕГОТОВНОСТЬ IR-КОНВЕЙЕРА: ОШИБКА СНЯТА",
		MessTxtState1: "НЕГОТОВНОСТЬ IR-КОНВЕЙЕРА: ОШИБКА УСТАНОВЛЕНА",
		MessColor0:    cBlack, MessColor1: cMessErr, MessType0: tErrOff, MessType1: tErrOn},
	1: {MessTxtState0: "НЕГОТОВНОСТЬ АЦП IR-КОНВЕЙЕРА: ОШИБКА СНЯТА",
		MessTxtState1: "НЕГОТОВНОСТЬ АЦП IR-КОНВЕЙЕРА: ОШИБКА УСТАНОВЛЕНА",
		MessColor0:    cBlack, MessColor1: cMessErr, MessType0: tErrOff, MessType1: tErrOn},
	2: {MessTxtState0: "ОШИБКА EEPROM: ОШИБКА СНЯТА",
		MessTxtState1: "ОШИБКА EEPROM: ОШИБКА УСТАНОВЛЕНА",
		MessColor0:    cBlack, MessColor1: cMessErr, MessType0: tErrOff, MessType1: tErrOn},
	3: {MessTxtState0: "ПРЕДВАРИТЕЛЬНОЕ (ПОЛОВИННОЕ) ЗАГРЯЗНЕНИЕ СТЕКЛА (ДЕТЕКТОР ИНЕЯ): ОШИБКА СНЯТА",
		MessTxtState1: "ПРЕДВАРИТЕЛЬНОЕ (ПОЛОВИННОЕ) ЗАГРЯЗНЕНИЕ СТЕКЛА (ДЕТЕКТОР ИНЕЯ): ОШИБКА УСТАНОВЛЕНА",
		MessColor0:    cBlack, MessColor1: cMessErr, MessType0: tErrOff, MessType1: tErrOn},
	4: {MessTxtState0: "ЗАГРЯЗНЕНИЕ СТЕКЛА: ОШИБКА СНЯТА",
		MessTxtState1: "ЗАГРЯЗНЕНИЕ СТЕКЛА: ОШИБКА УСТАНОВЛЕНА",
		MessColor0:    cBlack, MessColor1: cMessErr, MessType0: tErrOff, MessType1: tErrOn},
	5: {MessTxtState0: "НЕИСПРАВНОСТЬ ТРАКТА ПОЖАР: ОШИБКА СНЯТА",
		MessTxtState1: "НЕИСПРАВНОСТЬ ТРАКТА ПОЖАР: ОШИБКА УСТАНОВЛЕНА",
		MessColor0:    cBlack, MessColor1: cMessErr, MessType0: tErrOff, MessType1: tErrOn},
	6: {MessTxtState0: "ВНУТРЕННЯЯ ТЕМПЕРАТУРА НИЖЕ МИНИМАЛЬНОЙ -40.0: ОШИБКА СНЯТА",
		MessTxtState1: "ВНУТРЕННЯЯ ТЕМПЕРАТУРА НИЖЕ МИНИМАЛЬНОЙ -40.0: ОШИБКА УСТАНОВЛЕНА",
		MessColor0:    cBlack, MessColor1: cMessErr, MessType0: tErrOff, MessType1: tErrOn},
	7: {MessTxtState0: "ВНУТРЕННЯЯ ТЕМПЕРАТУРА ВЫШЕ МАКСИМАЛЬНОЙ +85.0: ОШИБКА СНЯТА",
		MessTxtState1: "ВНУТРЕННЯЯ ТЕМПЕРАТУРА ВЫШЕ МАКСИМАЛЬНОЙ +85.0: ОШИБКА УСТАНОВЛЕНА",
		MessColor0:    cBlack, MessColor1: cMessErr, MessType0: tErrOff, MessType1: tErrOn},
	8: {MessTxtState0: "ОШИБКА ВИДЕО-FLASH: ОШИБКА СНЯТА",
		MessTxtState1: "ОШИБКА ВИДЕО-FLASH: ОШИБКА УСТАНОВЛЕНА",
		MessColor0:    cBlack, MessColor1: cMessErr, MessType0: tErrOff, MessType1: tErrOn},
	9: {MessTxtState0: "НЕИСПРАВНОСТЬ ВИДЕОКАНАЛА: ОШИБКА СНЯТА",
		MessTxtState1: "НЕИСПРАВНОСТЬ ВИДЕОКАНАЛА: ОШИБКА УСТАНОВЛЕНА",
		MessColor0:    cBlack, MessColor1: cMessErr, MessType0: tErrOff, MessType1: tErrOn},
	10: {MessTxtState0: "ОТСУТСТВУЕТ ЗАВОДСКАЯ КОНФИГУРАЦИЯ: ОШИБКА СНЯТА",
		MessTxtState1: "ОТСУТСТВУЕТ ЗАВОДСКАЯ КОНФИГУРАЦИЯ: ОШИБКА УСТАНОВЛЕНА",
		MessColor0:    cBlack, MessColor1: cMessErr, MessType0: tErrOff, MessType1: tErrOn},
	11: {MessTxtState0: "НОВАЯ КОНФИГУРАЦИЯ НЕ ЗАПИСАНА: ОШИБКА СНЯТА",
		MessTxtState1: "НОВАЯ КОНФИГУРАЦИЯ НЕ ЗАПИСАНА: ОШИБКА УСТАНОВЛЕНА",
		MessColor0:    cBlack, MessColor1: cMessErr, MessType0: tErrOff, MessType1: tErrOn},
	12: {MessTxtState0: "ПЕРЕПОЛНЕНИЕ СТЕКА ВИДЕО-ЗАДАЧИ: ОШИБКА СНЯТА",
		MessTxtState1: "ПЕРЕПОЛНЕНИЕ СТЕКА ВИДЕО-ЗАДАЧИ: ОШИБКА УСТАНОВЛЕНА",
		MessColor0:    cBlack, MessColor1: cMessErr, MessType0: tErrOff, MessType1: tErrOn},
	13: {MessTxtState0: "НЕ НАСТРОЕН IR-ПОРОГ: ОШИБКА СНЯТА",
		MessTxtState1: "НЕ НАСТРОЕН IR-ПОРОГ: ОШИБКА УСТАНОВЛЕНА",
		MessColor0:    cBlack, MessColor1: cMessErr, MessType0: tErrOff, MessType1: tErrOn},
	14: {MessTxtState0: "ОТСУТСТВУЕТ СЕРИЙНЫЙ НОМЕР: ОШИБКА СНЯТА",
		MessTxtState1: "ОТСУТСТВУЕТ СЕРИЙНЫЙ НОМЕР: ОШИБКА УСТАНОВЛЕНА",
		MessColor0:    cBlack, MessColor1: cMessErr, MessType0: tErrOff, MessType1: tErrOn},
	15: {MessTxtState0: "ПОВРЕЖДЕНА КОЛБА: ОШИБКА СНЯТА",
		MessTxtState1: "ПОВРЕЖДЕНА КОЛБА: ОШИБКА УСТАНОВЛЕНА",
		MessColor0:    cBlack, MessColor1: cMessErr, MessType0: tErrOff, MessType1: tErrOn},
}
