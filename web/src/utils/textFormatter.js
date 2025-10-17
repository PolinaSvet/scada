// utils/textFormatter.js

/**
 * Форматирует текст с авто-переносом
 * @param {string} text - Исходный текст
 * @param {number} maxLength - Максимальная длина строки до переноса
 * @returns {string} Текст с переносами строк
 */
export function formatTextWithWrap(text, maxLength = 12) {
    if (!text || text.length <= maxLength) return text || ''
    
    // Поиск пробела для разбивки
    const spaceIndex = text.lastIndexOf(' ', maxLength)
    if (spaceIndex !== -1 && spaceIndex > 0) {
      return text.substring(0, spaceIndex) + '\n' + text.substring(spaceIndex + 1)
    }
    
    // Если пробелов нет, разбиваем по полам
    const mid = Math.floor(text.length / 2)
    return text.substring(0, mid) + '\n' + text.substring(mid)
  }
  
  /**
   * Создает стили для текстового элемента
   * @param {Object} options - Настройки стиля
   * @param {string} options.fontSize - Размер шрифта
   * @param {string} options.color - Цвет текста
   * @param {string} options.textAlign - Горизонтальное выравнивание (left|center|right)
   * @param {string} options.verticalAlign - Вертикальное выравнивание (top|middle|bottom)
   * @param {string} options.fontWeight - Жирность шрифта
   * @param {string} options.fontFamily - Семейство шрифтов
   * @param {number} options.lineHeight - Высота строки
   * @returns {Object} Объект стилей
   */
  export function createTextStyle(options = {}) {
    const {
      fontSize = '10px',
      color = '#000000',
      textAlign = 'center',
      verticalAlign = 'bottom',
      fontWeight = 'bold',
      fontFamily = 'Arial, Helvetica, sans-serif',
      lineHeight = 1.2
    } = options
  
    // Преобразуем verticalAlign в alignItems для flex
    const alignItemsMap = {
      top: 'flex-start',
      middle: 'center',
      bottom: 'flex-end'
    }
  
    // Преобразуем textAlign в justifyContent для flex
    const justifyContentMap = {
      left: 'flex-start',
      center: 'center',
      right: 'flex-end'
    }
  
    return {
      fontSize,
      fontWeight,
      textAlign,
      color,
      lineHeight,
      whiteSpace: 'pre-line',
      overflow: 'hidden',
      display: 'flex',
      alignItems: alignItemsMap[verticalAlign] || 'center',
      justifyContent: justifyContentMap[textAlign] || 'center',
      fontFamily,
      height: '100%',
      width: '100%'
  
    }
  }
  
  /**
   * Композитная функция для быстрого создания форматированного текста
   * @param {string} text - Исходный текст
   * @param {Object} options - Настройки
   * @returns {Object} { formattedText, textStyle }
   */
  export function createFormattedText(text, options = {}) {
    const {
      maxLength = 12,
      fontSize = '12px',
      color = '#000000',
      textAlign = 'center',
      verticalAlign = 'middle',
      fontWeight = 'bold',
      fontFamily = 'Arial, Helvetica, sans-serif',
      lineHeight = 1.2
    } = options
  
    return {
      formattedText: formatTextWithWrap(text, maxLength),
      textStyle: createTextStyle({
        fontSize,
        color,
        textAlign,
        verticalAlign,
        fontWeight,
        fontFamily,
        lineHeight
      })
    }
  }

/*
  1. Без засечек (Sans-serif) - наиболее читаемые:

javascript
fontFamily: 'Arial, Helvetica, sans-serif'
fontFamily: 'Tahoma, Geneva, sans-serif'
fontFamily: 'Trebuchet MS, Helvetica, sans-serif'
fontFamily: 'Verdana, Geneva, sans-serif'
2. С засечками (Serif):

javascript
fontFamily: 'Georgia, serif'
fontFamily: 'Times New Roman, Times, serif'
fontFamily: 'Palatino, Palatino Linotype, serif'
3. Моноширинные (Monospace):

javascript
fontFamily: 'Courier New, Courier, monospace'
fontFamily: 'Lucida Console, Monaco, monospace'
fontFamily: 'Consolas, monospace'
4. Универсальные комбинации:

javascript
// Самая безопасная комбинация
fontFamily: 'Arial, Helvetica, sans-serif'

// Для лучшей читаемости
fontFamily: 'Segoe UI, Tahoma, Geneva, Verdana, sans-serif'

// Современный вариант
fontFamily: 'system-ui, -apple-system, BlinkMacSystemFont, sans-serif'

Рекомендации по использованию:

Для основного текста: 'Arial, Helvetica, sans-serif'
Для заголовков: 'Georgia, serif'
Для кода/технической информации: 'Courier New, Courier, monospace'
Для современных интерфейсов: 'Segoe UI, Tahoma, Geneva, Verdana, sans-serif'

Преимущества Arial, Helvetica, sans-serif:
Есть во всех Windows, macOS, Linux
Хорошая читаемость в маленьких размерах
Поддерживает кириллицу и латиницу
Единообразное отображение в разных браузерах
  */