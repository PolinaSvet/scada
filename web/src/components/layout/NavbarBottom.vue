<template>
  <footer class="navbar-bottom" :class="panelClass">
    <div class="controls">
      <!-- Кнопки управления -->
      <div class="control-buttons">
        <button class="control-btn icon-minimize" @click="minimizePanel" title="Свернуть">
        </button>
        <button class="control-btn icon-expand" @click="expandTo200" title="Развернуть до 200px">
        </button>
        <button class="control-btn icon-maximize" @click="maximizePanel" title="Максимально развернуть">
        </button>
      </div>
      
      <!-- Текстовое поле с последним значением -->
      <div class="status-text">
        {{ statusText }}
      </div>
    </div>
    
    <!-- Расширенное содержимое -->
    <div v-if="isExpanded" class="expanded-content">
      <div class="table-section">
        <table>
          <thead>
            <tr>
              <th>ID</th>
              <th>Значение</th>
              <th>Статус</th>
              <th>Время</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="sensor in recentSensors" :key="sensor.id">
              <td>{{ sensor.id }}</td>
              <td>{{ sensor.value.toFixed(1) }}</td>
              <td :class="sensor.status">{{ sensor.status === 'normal' ? 'НОРМА' : 'АВАРИЯ' }}</td>
              <td>{{ sensor.lastUpdate }}</td>
            </tr>
          </tbody>
        </table>
      </div>
      
      <div class="buttons-section">
        <button class="action-btn">Обновить</button>
        <button class="action-btn">Настройки</button>
        <button class="action-btn">Экспорт</button>
        <button class="action-btn">Фильтр</button>
        <button class="action-btn">Сортировка</button>
      </div>
    </div>
  </footer>
</template>

<script>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useLayoutStore } from '@/stores/layout'

export default {
  name: 'NavbarBottom',
  setup() {
    const layoutStore = useLayoutStore()
    
    const sensors = ref([])
    const lastUpdateTime = ref(new Date())

    // Используем состояние из store
    const panelState = computed(() => layoutStore.bottomPanelState)
    
    // Вычисляемые свойства
    const isExpanded = computed(() => panelState.value !== 'minimized')
    
    const panelClass = computed(() => ({
      'minimized': panelState.value === 'minimized',
      'expanded': panelState.value === 'expanded200',
      'maximized': panelState.value === 'maximized'
    }))

    const statusText = computed(() => {
      const time = lastUpdateTime.value.toLocaleTimeString('ru-RU')
      const count = sensors.value.length
      const normalCount = sensors.value.filter(s => s.status === 'normal').length
      const warningCount = sensors.value.filter(s => s.status === 'warning').length
      
      return `Обновлено: ${time} | Всего: ${count} | Норма: ${normalCount} | Аварии: ${warningCount}`
    })

    const recentSensors = computed(() => sensors.value.slice(0, 500))

    // Методы управления панелью - теперь обновляют store
    const minimizePanel = () => {
      layoutStore.setBottomPanelState('minimized')
      console.log('Нижняя панель свернута до 40px')
    }

    const expandTo200 = () => {
      layoutStore.setBottomPanelState('expanded200')
      console.log('Нижняя панель развернута до 200px')
    }

    const maximizePanel = () => {
      layoutStore.setBottomPanelState('maximized')
      console.log('Нижняя панель максимально развернута до 400px')
    }

    // Форматирование времени для сенсоров
    const formatTime = (date) => {
      return date.toLocaleTimeString('ru-RU', {
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit'
      })
    }

    // Инициализация тестовых данных
    const initializeSensors = () => {
      sensors.value = Array.from({ length: 250 }, (_, i) => ({
        id: `SENSOR_${String(i + 1).padStart(3, '0')}`,
        value: Math.random() * 100,
        status: Math.random() > 0.3 ? 'normal' : 'warning',
        lastUpdate: formatTime(new Date())
      }))
      lastUpdateTime.value = new Date()
    }

    // Имитация обновления данных
    const startDataUpdates = () => {
      return setInterval(() => {
        // Обновляем несколько случайных сенсоров
        const updatesCount = Math.floor(Math.random() * 3) + 1 // 1-3 обновления
        for (let i = 0; i < updatesCount; i++) {
          if (sensors.value.length > 0) {
            const randomIndex = Math.floor(Math.random() * sensors.value.length)
            sensors.value[randomIndex].value = Math.random() * 100
            sensors.value[randomIndex].status = Math.random() > 0.3 ? 'normal' : 'warning'
            sensors.value[randomIndex].lastUpdate = formatTime(new Date())
          }
        }
        lastUpdateTime.value = new Date()
      }, 2000) // Обновление каждые 2 секунды
    }

    onMounted(() => {
      initializeSensors()
      const updateInterval = startDataUpdates()

      expandTo200()
      
      // Очистка при размонтировании
      onUnmounted(() => {
        clearInterval(updateInterval)
      })
    })

    return {
      panelState,
      isExpanded,
      panelClass,
      statusText,
      recentSensors,
      minimizePanel,
      expandTo200,
      maximizePanel
    }
  }
}
</script>

<style scoped>
@import '@/assets/styles/navbar-bottom.css';
</style>