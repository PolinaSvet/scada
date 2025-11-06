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
    
    console.log(`📊 Alarm store updated: ${opermessMessages.length} new, total: ${alarmMessStore.value.length}`)
  }
}

// Функция для очистки хранилища
export const clearAlarmStore = () => {
  alarmMessStore.value = []
  console.log('🗑️ Alarm store cleared')
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
  //const alarms = alarmMessStore.value.filter(msg => msg.messType === 3).length
  const alarms = alarmMessStore.value.filter(msg => msg.messType >= 1000 && msg.messType <= 2000).length
  const errors = alarmMessStore.value.filter(msg => msg.messType === 901).length
  const normal = total - alarms
  
  return { total, normal, alarms, errors }
})

// Экспортируем само хранилище для реактивности
export { alarmMessStore }