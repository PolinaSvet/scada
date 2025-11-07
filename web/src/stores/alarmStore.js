// objects.js
import { ref, computed } from 'vue'

// Хранилище для сообщений с opermess == 1 (максимум 500)
const alarmMessStore = ref([])
const maxAlarms = 500
let messageCounter = 0

// Функция для добавления сообщений в хранилище
export const addToAlarmStore = (messages) => {
  if (!Array.isArray(messages)) return
  
  const opermessMessages = messages.filter(msg => msg.opermess === 1)
  
  if (opermessMessages.length > 0) {
    // Добавляем уникальный ID к каждому сообщению
    const messagesWithId = opermessMessages.map(msg => ({
      ...msg,
      uniqueId: `msg_${messageCounter++}_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
    }))
    
    // Добавляем новые сообщения
    alarmMessStore.value = [...alarmMessStore.value, ...messagesWithId]
    
    // Ограничиваем размер хранилища
    if (alarmMessStore.value.length > maxAlarms) {
      alarmMessStore.value = alarmMessStore.value.slice(-maxAlarms)
    }
    
    //console.log(`?? Alarm store updated: ${opermessMessages.length} new, total: ${alarmMessStore.value.length}`)
  }
}

// Функция для очистки хранилища
export const clearAlarmStore = () => {
  alarmMessStore.value = []
  console.log('??? Alarm store cleared')
}

// Получение сообщений для отображения (последние 500, новые сверху)
export const getAlarmMessages = computed(() => {
  return [...alarmMessStore.value].reverse().map((msg, index) => ({
    ...msg,
    displayNumber: alarmMessStore.value.length - index // Номер записи
  }))
})

// Статистика для статусной строки
export const getAlarmStats = computed(() => {
  const total = alarmMessStore.value.length
  const alarms = alarmMessStore.value.filter(msg => msg.messType >= 1000 && msg.messType < 1100).length
  const errors = alarmMessStore.value.filter(msg => msg.messType === 901).length
  const normal = total - alarms
  
  return { total, normal, alarms, errors }
})

// Функция для получения категории сообщения
const getMessTypeCategory = (messType) => {
  // Алармы
  if (messType >= 1000 && messType < 1100) return 'alarm'
  // Предупреждения
  if (messType >= 1100 && messType < 1200) return 'warning'
  // Недостоверности
  if (messType >= 800 && messType < 900) return 'unreliable'
  // Неисправности
  if (messType === 901) return 'malfunction'
  // Прочее
  return 'other'
}

// Функция для получения иконки по категории
const getCategoryIcon = (category) => {
  const icons = {
    alarm: '🔥',
    warning: '⚠️',
    unreliable: '❓',
    malfunction: '⚡',
    other: 'ℹ️'
  }
  return icons[category] || 'ℹ️'
}

// Функция для получения цвета по категории
const getCategoryColor = (category) => {
  const colors = {
    alarm: '#ff4444',     // Красный для аварий
    warning: '#ff00ff',   // Оранжевый для предупреждений
    unreliable: '#C0C0C0', // Желтый для недостоверностей
    malfunction: '#ffaa00', // Пурпурный для неисправностей
    other: '#4287f5'      // Синий для прочего
  }
  return colors[category] || '#4287f5'
}

// Функция для получения наименования типа сообщения
const getMessTypeName = (messType) => {
  const typeNames = {
    0: 'Нет типа',
    
    900: 'Ошибка выключена',
    901: 'Ошибка включена',
    
    801: 'Статус ненадежен',
    101: 'Статус нормальный',
    1001: 'Статус пожара',
    1101: 'Статус внимания',
    
    3000: 'Имитация выключена',
    3001: 'Имитация включена',
    3010: 'Маскировка выключена',
    3011: 'Маскировка включена',
    3020: 'Квитирование выключено',
    3021: 'Квитирование включено',
    3030: 'Реальный режим выключен',
    3031: 'Реальный режим включен'
  }
  
  return typeNames[messType] || `Тип ${messType}`
}

// Функция для получения наименования категории
const getCategoryName = (category) => {
  const categoryNames = {
    alarm: 'Аварии',
    warning: 'Предупреждения',
    unreliable: 'Недостоверности',
    malfunction: 'Неисправности',
    other: 'Прочие события'
  }
  return categoryNames[category] || 'Неизвестная категория'
}

// Группировка по типам сообщений
export const getAlarmStatsInfo = computed(() => {
  const groups = {}
  
  alarmMessStore.value.forEach(msg => {
    const messType = msg.messType
    
    if (!groups[messType]) {
      const category = getMessTypeCategory(messType)
      groups[messType] = {
        messType: messType,
        count: 0,
        messColor: msg.messColor || getCategoryColor(category),
        name: getMessTypeName(messType),
        category: category,
        categoryName: getCategoryName(category),
        icon: getCategoryIcon(category)
      }
    }
    
    groups[messType].count++
  })
  
  // Преобразуем объект в массив и сортируем по messType
  return Object.values(groups).sort((a, b) => a.messType - b.messType)
})

// Группировка по категориям
export const getAlarmStatsByCategory = computed(() => {
  const categories = {}
  
  alarmMessStore.value.forEach(msg => {
    const category = getMessTypeCategory(msg.messType)
    
    if (!categories[category]) {
      categories[category] = {
        category: category,
        name: getCategoryName(category),
        icon: getCategoryIcon(category),
        color: getCategoryColor(category),
        count: 0,
        types: {}
      }
    }
    
    categories[category].count++
    
    // Также считаем по типам внутри категории
    const messType = msg.messType
    if (!categories[category].types[messType]) {
      categories[category].types[messType] = {
        messType: messType,
        name: getMessTypeName(messType),
        count: 0
      }
    }
    categories[category].types[messType].count++
  })
  
  // Преобразуем объект в массив и сортируем по приоритету категорий
  const categoryOrder = ['alarm', 'warning', 'malfunction', 'unreliable', 'other']
  return Object.values(categories)
    .sort((a, b) => categoryOrder.indexOf(a.category) - categoryOrder.indexOf(b.category))
})

// Экспортируем само хранилище для реактивности
export { alarmMessStore }