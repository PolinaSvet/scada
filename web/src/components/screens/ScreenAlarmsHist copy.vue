<template>
  <div class="screen-hist-alarms">

    <!-- Фильтры -->
    <div class="filters-section">
      <div class="filter-group">
        <label>Дата/время:</label>
        <div class="date-filters">
          <button class="filter-btn" @click="setDateRange('today')" :class="{ active: dateRange === 'today' }">
            Сегодня
          </button>
          <button class="filter-btn" @click="setDateRange('yesterday')" :class="{ active: dateRange === 'yesterday' }">
            Вчера
          </button>
          <button class="filter-btn" @click="setDateRange('week')" :class="{ active: dateRange === 'week' }">
            Неделя
          </button>
          <button class="filter-btn" @click="setDateRange('month')" :class="{ active: dateRange === 'month' }">
            Месяц
          </button>
        </div>
      </div>

      <div class="filter-group">
        <label>Параметры фильтрации:</label>
        <div class="checkbox-filters">
          <label class="checkbox-label">
            <input type="checkbox" v-model="filters.tagFind"> Тег
            <input type="text" v-model="filterValues.tagFind" :disabled="!filters.tagFind" placeholder="TAG_1">
          </label>
          <label class="checkbox-label">
            <input type="checkbox" v-model="filters.messFullFind"> Сообщение
            <input type="text" v-model="filterValues.messFullFind" :disabled="!filters.messFullFind" placeholder="Текст сообщения">
          </label>
          <label class="checkbox-label">
            <input type="checkbox" v-model="filters.usoTxtFind"> Диагностика
            <input type="text" v-model="filterValues.usoTxtFind" :disabled="!filters.usoTxtFind" placeholder="Текст диагностики">
          </label>
          <label class="checkbox-label">
            <input type="checkbox" v-model="filters.severityFind"> Тревога
            <select v-model="filterValues.severityFind" :disabled="!filters.severityFind">
              <option value="0">Все</option>
              <option value="901">Неисправность</option>
              <option value="1001">Пожар</option>
              <option value="1101">Внимание</option>
            </select>
          </label>
          <label class="checkbox-label">
            <input type="checkbox" v-model="filters.kvitFind"> Квитирование
            <select v-model="filterValues.kvitFind" :disabled="!filters.kvitFind">
              <option value="0">Все</option>
              <option value="1">Неквитированные</option>
              <option value="2">Квитированные</option>
            </select>
          </label>
        </div>
      </div>
    </div>

    <!-- Управление -->
    <div class="controls-section">
      <div class="control-buttons">
        <button class="control-btn" @click="refreshData" title="Обновить данные">
          🔄 Обновить
        </button>
        <button class="control-btn" @click="toggleTableHeader" title="Переключить заголовок таблицы">
          {{ showTableHeader ? '📋' : '📄' }} Заголовок
        </button>
        <button class="control-btn" @click="toggleColumnVisibility" title="Показать/скрыть дополнительные колонки">
          {{ showAllColumns ? '👁️' : '👁️‍🗨️' }} Колонки
        </button>
        <button class="control-btn" @click="handleColorToggle" title="Изменить стиль цветов">
          {{ colorMode === 'text' ? '🎨' : '📝' }} Цвета
        </button>
        <button class="control-btn" @click="exportData" title="Экспортировать данные">
          💾 Экспорт
        </button>
        <button class="control-btn" @click="clearData" title="Очистить данные">
          🗑️ Очистить
        </button>
      </div>
      
      <div class="status-text">
        {{ statusText }}
      </div>
    </div>

    <!-- Основная таблица -->
    <div class="table-container">
      <table class="alarms-table">
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
            <th>Квит.</th>
            <!-- Дополнительные колонки -->
            <th v-if="showAllColumns">ID Объекта</th>
            <th v-if="showAllColumns">Тип Объекта</th>
            <th v-if="showAllColumns">Пользователи</th>
            <th v-if="showAllColumns">Цвет</th>
            <th v-if="showAllColumns">Время квит.</th>
            <th v-if="showAllColumns">Текст квит.</th>
          </tr>
        </thead>
        <tbody>
          <tr 
            v-for="alarm in displayAlarms" 
            :key="alarm.uniqueId"
            :style="getRowStyle(alarm)"
          >
            <td class="number-cell">{{ alarm.id }}</td>
            <td class="id-cell">{{ alarm.code }}</td>
            <td class="time-cell">{{ alarm.dt_txt || formatTime(alarm.dt) }}</td>
            <td class="tag-cell">{{ alarm.tag || '-' }}</td>
            <td class="desc-cell">{{ alarm.mess_name || '-' }}</td>
            <td class="message-cell">{{ alarm.mess_full || '-' }}</td>
            <td class="uso-cell">{{ alarm.mess_state || '-' }}</td>
            <td class="type-cell">{{ alarm.severity }}</td>
            <td class="type-cell">{{ alarm.opermess }}</td>
            <td class="kvit-cell">{{ alarm.kvit ? '✅' : '❌' }}</td>
            <!-- Дополнительные колонки -->
            <td v-if="showAllColumns" class="id-cell">{{ alarm.id_obj || '-' }}</td>
            <td v-if="showAllColumns" class="type-cell">{{ alarm.type_obj || '-' }}</td>
            <td v-if="showAllColumns" class="users-cell">{{ alarm.users || '-' }}</td>
            <td v-if="showAllColumns" class="color-cell">
              <span class="color-indicator" :style="{ backgroundColor: alarm.color }"></span>
            </td>
            <td v-if="showAllColumns" class="time-cell">{{ alarm.dt_kvit_txt || formatTime(alarm.dt_kvit) }}</td>
            <td v-if="showAllColumns" class="text-cell">{{ alarm.dt_kvit_txt || '-' }}</td>
          </tr>
        </tbody>
      </table>
      
      <!-- Сообщение если нет данных -->
      <div v-if="displayAlarms.length === 0" class="no-data-message">
        <p>📭 Нет исторических данных для отображения</p>
        <p>Используйте кнопку "Обновить" для загрузки данных</p>
      </div>
    </div>

    <!-- Пагинация -->
    <div class="pagination-section" v-if="displayAlarms.length > 0 && totalPages > 1">
      <div class="pagination-info">
        Показано записей: {{ displayAlarms.length }} | Страница {{ currentPage }} из {{ totalPages }}
      </div>
      <div class="pagination-buttons">
        <button class="pagination-btn" @click="goToPage(1)" :disabled="currentPage === 1">
          ⏮️ Первая
        </button>
        <button class="pagination-btn" @click="goToPage(currentPage - 10)" :disabled="currentPage <= 10">
          -10
        </button>
        <button class="pagination-btn" @click="goToPage(currentPage - 1)" :disabled="currentPage === 1">
          ◀️ Назад
        </button>
        
        <span class="page-current">{{ currentPage }}</span>
        
        <button class="pagination-btn" @click="goToPage(currentPage + 1)" :disabled="currentPage >= totalPages">
          Вперед ▶️
        </button>
        <button class="pagination-btn" @click="goToPage(currentPage + 10)" :disabled="currentPage + 10 > totalPages">
          +10
        </button>
        <button class="pagination-btn" @click="goToPage(totalPages)" :disabled="currentPage === totalPages">
          Последняя ⏭️
        </button>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, computed, onMounted, watch } from 'vue'
import { 
  getAlarmHistMessages, 
  clearAlarmHistStore 
} from '@/stores/alarmStoreHist.js'
import {
  colorMode,
  toggleColorMode,
  formatTime,
  getRowStyle,
  saveAsHTML
} from '@/utils/funcAlarmStore.js'
import { useObjectsStore } from '@/stores/objects'


export default {
  name: 'ScreenHistAlarms',
  setup() {
    const objectsStore = useObjectsStore()
    const showTableHeader = ref(true)
    const showAllColumns = ref(false)
    const lastUpdateTime = ref(new Date())
    const currentPage = ref(1)
    const totalPages = ref(1)
    const dateRange = ref('today')

    // Фильтры
    const filters = ref({
      tagFind: false,
      messFullFind: false,
      usoTxtFind: false,
      severityFind: false,
      kvitFind: false
    })

    const filterValues = ref({
      tagFind: '',
      messFullFind: '',
      usoTxtFind: '',
      severityFind: 0,
      kvitFind: 0
    })

    // Данные из исторического хранилища
    const displayAlarms = computed(() => getAlarmHistMessages.value)

    // Обновляем пагинацию при изменении данных
    watch(displayAlarms, (newAlarms) => {
      if (newAlarms.length > 0) {
        const firstAlarm = newAlarms[0]
        currentPage.value = firstAlarm.current_page || 1
        totalPages.value = firstAlarm.total_pages || 1
      }
    })

    // Статусная строка
    const statusText = computed(() => {
      const time = lastUpdateTime.value.toLocaleTimeString('ru-RU')
      const count = displayAlarms.value.length
      return `Обновлено: ${time} | Записей: ${count} | Страница: ${currentPage.value}/${totalPages.value}`
    })

    // Функции управления
    const setDateRange = (range) => {
      dateRange.value = range
      refreshData()
    }

    const refreshData = () => {
      // Вычисляем даты в зависимости от выбранного диапазона
      const now = Date.now()
      let dtStart = now - 24 * 60 * 60 * 1000 // по умолчанию сутки

      switch (dateRange.value) {
        case 'today':
          dtStart = new Date().setHours(0, 0, 0, 0)
          break
        case 'yesterday':
          dtStart = new Date().setHours(0, 0, 0, 0) - 24 * 60 * 60 * 1000
          break
        case 'week':
          dtStart = now - 7 * 24 * 60 * 60 * 1000
          break
        case 'month':
          dtStart = now - 30 * 24 * 60 * 60 * 1000
          break
      }

      const commandData = {
        dtStart: dtStart,
        dtEnd: now,
        pageNum: currentPage.value
      }

      // Добавляем только активные фильтры
      if (filters.value.tagFind && filterValues.value.tagFind) {
        commandData.tagFind = filterValues.value.tagFind
      }
      if (filters.value.messFullFind && filterValues.value.messFullFind) {
        commandData.messFullFind = filterValues.value.messFullFind
      }
      if (filters.value.usoTxtFind && filterValues.value.usoTxtFind) {
        commandData.usoTxtFind = filterValues.value.usoTxtFind
      }
      if (filters.value.severityFind && filterValues.value.severityFind > 0) {
        commandData.severityFind = filterValues.value.severityFind
      }
      if (filters.value.kvitFind && filterValues.value.kvitFind > 0) {
        commandData.kvitFind = filterValues.value.kvitFind
      }

      objectsStore.sendCommand(
        'alarms_system',
        'command',
        'alarms_get_data',
        commandData
      )
      
      lastUpdateTime.value = new Date()
    }

    const toggleTableHeader = () => {
      showTableHeader.value = !showTableHeader.value
    }

    const toggleColumnVisibility = () => {
      showAllColumns.value = !showAllColumns.value
    }

    const handleColorToggle = () => {
      toggleColorMode()
    }

    const exportData = () => {
      saveAsHTML(displayAlarms.value, 'history_alarms')
    }

    const clearData = () => {
      clearAlarmHistStore()
      lastUpdateTime.value = new Date()
      console.log('Исторические данные очищены')
    }

    const goToPage = (page) => {
      if (page >= 1 && page <= totalPages.value) {
        currentPage.value = page
        refreshData()
      }
    }

    // При монтировании загружаем данные
    onMounted(() => {
      refreshData()
    })

    return {
      displayAlarms,
      showTableHeader,
      showAllColumns,
      currentPage,
      totalPages,
      dateRange,
      filters,
      filterValues,
      colorMode,
      statusText,
      refreshData,
      toggleTableHeader,
      toggleColumnVisibility,
      handleColorToggle,
      exportData,
      clearData,
      setDateRange,
      goToPage,
      formatTime,
      getRowStyle
    }
  }
}
</script>

<style scoped>
/* Основные стили вынесены в отдельный CSS файл */
@import '@/assets/styles/screen-alarms-hist.css';
</style>