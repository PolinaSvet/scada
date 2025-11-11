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
        <button class="control-btn icon-header" @click="toggleTableHeader" title="Изменить стиль отображения заголовка">
        <span>{{ showTableHeader ? '📋' : '📄' }}</span>
        </button>
        <button class="control-btn icon-color" @click="handleColorToggle" title="Изменить стиль отображения цвета">
        <span>{{ colorMode === 'text' ? '🔲' : '🔳' }}</span>
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
          <thead v-if="showTableHeader">
            <tr>
              <th>№</th>
              <th>ID</th>
              <th>Время</th>
              <th>Тег</th>
              <th>Описание</th>
              <th>Сообщение</th>
              <th>Диагностика</th>
              <th>Т.C.</th>
              <th>Т.O.</th>
            </tr>
          </thead>
          <tbody>
            <tr 
              v-for="alarm in displayAlarms" 
              :key="alarm.uniqueId"
              :style="getRowStyle(alarm)"
            >
              <td class="number-cell">{{ alarm.displayNumber }}</td>
              <td class="id-cell">{{ alarm.code }}</td>
              <td class="time-cell">{{ alarm.dt_txt}}</td>
              <td class="tag-cell">{{ alarm.tag || '-' }}</td>
              <td class="desc-cell">{{ alarm.mess_name || '-' }}</td>
              <td class="message-cell">{{ alarm.messTxt || '-' }}</td>
              <td class="uso-cell">{{ alarm.mess_state || '-' }}</td>
              <td class="type-cell">{{ alarm.severity }}</td>
              <td class="type-cell">{{  alarm.type_obj }}</td>
            </tr>
          </tbody>
        </table>
      </div>
      
      <div class="buttons-section">
        <button class="action-btn" @click="confirmAlarms">
          <span class="icon-confirm">✅</span> Очистить
        </button>
        <button class="action-btn" @click="handleSave">
          <span class="icon-save">💾</span> Сохранить
        </button>
      </div>
    </div>
  </footer>
</template>

<script>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useLayoutStore } from '@/stores/layout'
import { 
  getAlarmMessages, 
  getAlarmStats, 
  clearAlarmStore 
} from '@/stores/alarmStore.js'
import {
  colorMode,
  toggleColorMode,
  formatTime,
  getRowStyle,
  saveAsHTML
} from '@/utils/funcAlarmStore.js'

export default {
  name: 'NavbarBottom',
  setup() {
    const layoutStore = useLayoutStore()
    const showTableHeader = ref(true) // Показывать заголовки таблицы
    
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

    // Сообщения из хранилища
    const displayAlarms = computed(() => getAlarmMessages.value)

    // Статистика из хранилища
    const alarmStats = computed(() => getAlarmStats.value)

    const statusText = computed(() => {
      const time = lastUpdateTime.value.toLocaleTimeString('ru-RU')
      const count = alarmStats.value.total
      //const normalCount = alarmStats.value.normal
      const alarmCount = alarmStats.value.alarms
      const errorCount = alarmStats.value.errors;
      
      return `Обновлено: ${time} | Всего: ${count} | Неисправность: ${errorCount} | Пожар: ${alarmCount}`
    })

    // Методы управления панелью
    const minimizePanel = () => {
      layoutStore.setBottomPanelState('minimized')
    }

    const expandTo200 = () => {
      layoutStore.setBottomPanelState('expanded200')
    }

    const maximizePanel = () => {
      layoutStore.setBottomPanelState('maximized')
    }

    // Обработчики кнопок
    const handleColorToggle = () => {
      toggleColorMode()
    }

    const handleSave = () => {
      saveAsHTML(displayAlarms.value)
    }

    // Переключение отображения заголовков таблицы
    const toggleTableHeader = () => {
      showTableHeader.value = !showTableHeader.value
    }

    // Кнопка подтверждения - очистка хранилища
    const confirmAlarms = () => {
      clearAlarmStore()
      lastUpdateTime.value = new Date()
      console.log('Хранилище тревог очищено')
    }

    // Следим за изменениями в хранилище для обновления времени
    watch(displayAlarms, () => {
      lastUpdateTime.value = new Date()
    })

    onMounted(() => {
      expandTo200()
    })

    return {
      panelState,
      isExpanded,
      panelClass,
      statusText,
      displayAlarms,
      toggleTableHeader,
      showTableHeader,
      minimizePanel,
      expandTo200,
      maximizePanel,
      formatTime,
      getRowStyle,
      confirmAlarms,
      handleSave,
      handleColorToggle,
      colorMode
    }
  }
}
</script>

<style scoped>
@import '@/assets/styles/navbar-bottom.css';

/* Дополнительные стили для иконок кнопок */
.icon-save, .icon-confirm, .icon-color {
  margin-right: 5px;
  font-size: 14px;
}
</style>