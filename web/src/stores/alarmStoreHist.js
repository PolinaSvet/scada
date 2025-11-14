import { ref, computed } from 'vue'

// Хранилище для исторических данных алармов (полная замена при обновлении)
const alarmStoreHist = ref([])

// Функция для полного обновления хранилища исторических данных
export const addMessHistBatch = (objectsArray) => {
  if (!Array.isArray(objectsArray)) return
  
  // Полностью заменяем содержимое хранилища
  alarmStoreHist.value = objectsArray.map((msg, index) => ({
    ...msg,
    uniqueId: `hist_${index}_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
  }))
  
  console.log(`?? Alarm history store updated: ${objectsArray.length} records`)
}

// Функция для очистки хранилища
export const clearAlarmHistStore = () => {
  alarmStoreHist.value = []
  console.log('??? Alarm history store cleared')
}

// Получение исторических сообщений для отображения (последние сверху)
export const getAlarmHistMessages = computed(() => {
  return [...alarmStoreHist.value].map((msg, index) => ({
    ...msg,
    displayNumber: alarmStoreHist.value.length - index // Номер записи
  }))
})



// Экспортируем само хранилище для реактивности
export { alarmStoreHist }