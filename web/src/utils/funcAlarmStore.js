// utils/funcAlarmMessages.js
import { ref } from 'vue'

// Реактивная переменная для режима цвета
export const colorMode = ref('background') // 'text' или 'background'

// Переключение режима цвета
export const toggleColorMode = () => {
  colorMode.value = colorMode.value === 'text' ? 'background' : 'text'
  console.log(`Режим цвета изменен на: ${colorMode.value}`)
  return colorMode.value
}

/*
// Форматирование времени
export const formatTime = (timestamp) => {
  if (!timestamp) return '-'
  const date = new Date(timestamp)
  return date.toLocaleTimeString('ru-RU', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}*/

// Форматирование времени с разными форматами
export const formatTime = (timestamp, format = 'default') => {
  if (!timestamp) return '-'
  
  const date = new Date(timestamp)
  
  switch (format) {
    case 'full':
      // DD.MM.YYYY HH:MM:SS.ZZZ
      const day = String(date.getDate()).padStart(2, '0')
      const month = String(date.getMonth() + 1).padStart(2, '0')
      const year = date.getFullYear()
      const hours = String(date.getHours()).padStart(2, '0')
      const minutes = String(date.getMinutes()).padStart(2, '0')
      const seconds = String(date.getSeconds()).padStart(2, '0')
      const milliseconds = String(date.getMilliseconds()).padStart(3, '0')
      
      return `${day}.${month}.${year} ${hours}:${minutes}:${seconds}.${milliseconds}`
    
    case 'time':
      // Только время HH:MM:SS
      return date.toLocaleTimeString('ru-RU', {
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit'
      })
    
    case 'datetime':
      // Дата и время без миллисекунд
      return date.toLocaleString('ru-RU', {
        day: '2-digit',
        month: '2-digit',
        year: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit'
      }).replace(/,/g, '')
    
    case 'default':
    default:
      // Стандартный формат HH:MM:SS
      return date.toLocaleTimeString('ru-RU', {
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit'
      })
  }
}

// Получение текста типа сообщения
export const getMessTypeText = (messType) => {
  const types = {
    1: 'Информация',
    2: 'Авария',
    3: 'Предупреждение',
    4: 'Событие'
  }
  return types[messType] || `Тип ${messType}`
}

// Функция для вычисления контрастного цвета текста
export const getContrastColor = (hexcolor) => {
  if (!hexcolor || hexcolor === '') return '#000000'
  
  // Удаляем # если есть
  hexcolor = hexcolor.replace('#', '')
  
  // Если короткая форма, преобразуем в полную
  if (hexcolor.length === 3) {
    hexcolor = hexcolor.split('').map(char => char + char).join('')
  }
  
  // Проверяем валидность hex цвета
  if (hexcolor.length !== 6) return '#000000'
  
  try {
    const r = parseInt(hexcolor.substr(0, 2), 16)
    const g = parseInt(hexcolor.substr(2, 2), 16)
    const b = parseInt(hexcolor.substr(4, 2), 16)
    
    // Формула для вычисления яркости
    const brightness = ((r * 299) + (g * 587) + (b * 114)) / 1000
    
    return brightness > 128 ? '#000000' : '#FFFFFF'
  } catch (error) {
    return '#000000'
  }
}

// Стиль строки в зависимости от режима цвета
export const getRowStyle = (alarm) => {
  if (!alarm.messColor) return {}
  
  if (colorMode.value === 'text') {
    // В режиме текста добавляем контрастный фон
    const textColor = alarm.messColor
    const backgroundColor = '#F3F3F3'
    return { 
      color: textColor,
      backgroundColor: backgroundColor,
      fontWeight: 'bold'
    }
  } else {
    // В режиме фона используем messColor как фон и контрастный текст
    const backgroundColor = alarm.messColor
    const textColor = getContrastColor(backgroundColor)
    return {
      backgroundColor: backgroundColor,
      color: textColor,
      fontWeight: 'bold'
    }
  }
}

// Сохранение таблицы в HTML
export const saveAsHTML = (alarms, fileName = null) => {
  try {
    const htmlContent = generateHTML(alarms)
    const blob = new Blob([htmlContent], { type: 'text/html' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = fileName || `alarms_${new Date().toISOString().slice(0, 19).replace(/:/g, '-')}.html`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
    console.log('Таблица сохранена как HTML')
    return true
  } catch (error) {
    console.error('Ошибка при сохранении HTML:', error)
    return false
  }
}

// Генерация HTML содержимого
const generateHTML = (alarms) => {
  // Определяем стили в зависимости от режима цвета
  const getRowHTML = (alarm) => {
    if (colorMode.value === 'text') {
      const textColor = alarm.messColor
      const backgroundColor = getContrastColor(textColor) === '#F3F3F3' ? '#000000' : '#F3F3F3'
      return `
      <tr style="color: ${textColor}; background-color: ${backgroundColor}; font-weight: bold;">
        <td>${alarm.displayNumber || '-'}</td>
        <td>${alarm.id}</td>
        <td class="timestamp">${formatTime(alarm.timestamp)}</td>
        <td>${alarm.info?.tag || '-'}</td>
        <td>${alarm.info?.desc || '-'}</td>
        <td>${alarm.messTxt || '-'}</td>
        <td>${alarm.uso?.txt || '-'}</td>
        <td>${getMessTypeText(alarm.messType)}</td>
      </tr>`
    } else {
      const backgroundColor = alarm.messColor
      const textColor = getContrastColor(backgroundColor)
      return `
      <tr style="background-color: ${backgroundColor}; color: ${textColor}; font-weight: bold;">
        <td>${alarm.displayNumber || '-'}</td>
        <td>${alarm.id}</td>
        <td class="timestamp">${formatTime(alarm.timestamp)}</td>
        <td>${alarm.info?.tag || '-'}</td>
        <td>${alarm.info?.desc || '-'}</td>
        <td>${alarm.messTxt || '-'}</td>
        <td>${alarm.uso?.txt || '-'}</td>
        <td>${getMessTypeText(alarm.messType)}</td>
      </tr>`
    }
  }

  return `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Тревоги - ${new Date().toLocaleString('ru-RU')}</title>
    <style>
        body { 
            font-family: 'Courier New', monospace; 
            margin: 20px; 
            font-size: 12px;
        }
        table { 
            width: 100%; 
            border-collapse: collapse; 
            margin-top: 20px;
            font-family: 'Courier New', monospace;
            font-size: 12px;
        }
        th, td { 
            border: 1px solid #333; 
            padding: 6px 8px; 
            text-align: left;
            font-family: 'Courier New', monospace;
            font-size: 12px;
        }
        th { 
            background-color: #2C3E50; 
            color: white;
            font-weight: bold;
            font-family: 'Courier New', monospace;
            font-size: 12px;
        }
        tr:nth-child(even) { 
            background-color: #f9f9f9; 
        }
        .timestamp { 
            white-space: nowrap; 
            font-family: 'Courier New', monospace;
        }
        .header-info {
            margin-bottom: 15px;
            font-family: 'Courier New', monospace;
            font-size: 12px;
        }
    </style>
</head>
<body>
    <h1 style="font-family: 'Courier New', monospace;">Журнал тревог</h1>
    <div class="header-info">
        <p><strong>Сгенерировано:</strong> ${new Date().toLocaleString('ru-RU')}</p>
        <p><strong>Всего записей:</strong> ${alarms.length}</p>
        <p><strong>Режим отображения:</strong> ${colorMode.value === 'text' ? 'Цвет текста' : 'Цвет фона'}</p>
    </div>
    
    <table>
        <thead>
            <tr>
                <th>№</th>
                <th>ID</th>
                <th>Время</th>
                <th>Тег</th>
                <th>Описание</th>
                <th>Сообщение</th>
                <th>Использование</th>
                <th>Тип</th>
            </tr>
        </thead>
        <tbody>
            ${alarms.map(alarm => getRowHTML(alarm)).join('')}
        </tbody>
    </table>
</body>
</html>`
}

// Экспортируем все функции по умолчанию
export default {
  colorMode,
  toggleColorMode,
  formatTime,
  getMessTypeText,
  getContrastColor,
  getRowStyle,
  saveAsHTML
}