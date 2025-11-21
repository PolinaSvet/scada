<template>
  <div class="screen-hist-alarms">
    <!-- Панель управления -->
    <div class="controls-section">
      <div class="pagination-controls">
        <button class="control-btn" @click="refreshData" title="Обновить данные">
          🔄 Обновить
        </button>
        <button class="control-btn" @click="refreshTrend" title="Обновить данные">
          🔄 Trend
        </button>
        <button class="control-btn" @click="showFilterDialog = true">
          🔍 Фильтры
        </button>
        <div class="filter-status" 
            :class="{ 'has-filters': hasActiveFilters }"
            :title="hasActiveFilters ? activeFiltersText : 'Фильтры не применены'">
          <span v-if="hasActiveFilters">Активные фильтры: {{ activeFiltersText }}</span>
          <span v-else>Фильтры не применены</span>
        </div>
      </div>

      <div class="pagination-controls">
        <button class="pagination-btn" @click="goToPage(1)" :disabled="currentPage === 1">
         ⏮️ 1
        </button>
        <button class="pagination-btn" @click="goToPage(currentPage - 10)" :disabled="currentPage <= 10">
         ◀️ -10
        </button>
        <button class="pagination-btn" @click="goToPage(currentPage - 1)" :disabled="currentPage === 1">
         ◀️ -1
        </button>
        
        <div class="page-input-group">
          <span>Page</span>
          <input 
            type="number" 
            v-model.number="pageInput" 
            @keyup.enter="goToPageInput"
            :min="1" 
            :max="totalPages"
            class="page-input"
          >
          <span>of {{ totalPages }}</span>
        </div>
        
        <button class="pagination-btn" @click="goToPage(currentPage + 1)" :disabled="currentPage >= totalPages">
          +1 ▶️
        </button>
        <button class="pagination-btn" @click="goToPage(currentPage + 10)" :disabled="currentPage + 10 > totalPages">
          +10 ▶️
        </button>
        <button class="pagination-btn" @click="goToPage(totalPages)" :disabled="currentPage === totalPages">
         {{ totalPages }} ⏭️
        </button>
        <span>Rows {{ displayAlarms.length }}</span>
      </div>

      <div class="pagination-controls">
        <div class="filter-controls">
          <button class="control-btn" @click="showControlDialog = true">
            ⚙️ Управление
          </button>
        </div>
      </div>
    </div>

    <!-- Основная таблица -->
    <div class="table-container">
      <table class="alarms-table">
        <thead v-if="showTableHeader">
          <tr>
            <th>№</th>
            <th>ID</th>
            <th>КОД</th>
            <th>ДАТА</th>
            <th>ТЕГ</th>
            <th>НАИМЕНОВАНИЕ</th>
            <th>СООБЩЕНИЕ</th>
            <th>ДИАГНОСТИКА</th>
            <th>Т.C.</th>
            <th>Н.O.</th>
            <th>Т.O.</th>
            <th>КВИТ.</th>
            <!-- Дополнительные колонки -->
            <th v-if="showAllColumns">ДАТА КВИТ.</th>
            <th v-if="showAllColumns">ПОЛЬЗОВАТЕЛЬ</th>
            <th v-if="showAllColumns">ВИДИМОСТЬ</th>
            <th v-if="showAllColumns">ЦВЕТ</th>
            <th v-if="showAllColumns">USO</th>
          </tr>
        </thead>
        <tbody>
          <tr 
            v-for="alarm in displayAlarms" 
            :key="alarm.uniqueId"
            :style="getRowStyle(alarm)"
          >
            <td class="number-cell">{{ alarm.displayNumber }}</td>
            <td class="number-cell">{{ alarm.id }}</td>
            <td class="id-cell">{{ alarm.code }}</td>
            <td class="time-cell">{{ alarm.dt_txt }}</td>
            <td class="tag-cell">{{ alarm.tag || '-' }}</td>
            <td class="desc-cell">{{ alarm.mess_name || '-' }}</td>
            <td class="message-cell">{{ alarm.mess_state || '-' }}</td>
            <td class="uso-cell">{{ alarm.uso_txt || '-' }}</td>
            <td class="type-cell">{{ alarm.severity }}</td>
            <td class="type-cell">{{ alarm.id_obj }}</td>
            <td class="type-cell">{{ alarm.type_obj }}</td>
            <td class="kvit-cell">{{ alarm.kvit ? '✅' : '❌' }}</td>
            <!-- Дополнительные колонки -->
            <td v-if="showAllColumns" class="time-cell">{{ alarm.dt_kvit_txt|| '-' }}</td>
            <td v-if="showAllColumns" class="users-cell">{{ alarm.users || '-' }}</td>
            <td v-if="showAllColumns" class="type-cell">{{ alarm.opermess || '-' }}</td>
            <td v-if="showAllColumns" class="type-cell">{{ alarm.color || '-' }}</td>
            <td v-if="showAllColumns" class="type-cell">{{ alarm.uso_id || '-' }}</td>
          </tr>
        </tbody>
      </table>
      
      <!-- Сообщение если нет данных -->
      <div v-if="displayAlarms.length === 0" class="no-data-message">
        <p>🗑️ Нет исторических данных для отображения</p>
        <p>Используйте кнопку "Обновить" для загрузки данных</p>
        <p>или изменить фильтр: {{activeFiltersText}}</p>
      </div>
    </div>

    <!-- Диалоговое окно управления -->
    <div v-if="showControlDialog" class="filter-dialog-overlay" @click="showControlDialog = false">
      <div class="filter-dialog" @click.stop>
        <div class="filter-dialog-header">
          <h3>Управление</h3>
          <button class="close-btn" @click="showControlDialog = false">×</button>
        </div>

        <div class="controls-section">
          <div class="control-buttons">
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
        </div>
      </div>
    </div>

    <!-- Диалоговое окно фильтров -->
    <div v-if="showFilterDialog" class="filter-dialog-overlay" @click="showFilterDialog = false">
      <div class="filter-dialog" @click.stop>
        <div class="filter-dialog-header">
          <h3>Фильтры</h3>
          <button class="close-btn" @click="showFilterDialog = false">×</button>
        </div>
        
        <div class="filter-dialog-content">
          <!-- Фильтр по дате/времени - период -->
          <div class="filter-group-dialog">
            <label class="checkbox-label-dialog">
              <input type="checkbox" v-model="dialogFilters.dateTimeRangeEnabled">
              <span> Период даты/времени</span>
            </label>
            <div class="date-time-inputs" v-if="dialogFilters.dateTimeRangeEnabled">
              <div class="date-input-group">
                <label>Начало:</label>
                <input type="datetime-local" v-model="dialogFilters.dateTimeStart">
              </div>
              <div class="date-input-group">
                <label>Конец:</label>
                <input type="datetime-local" v-model="dialogFilters.dateTimeEnd">
              </div>
            </div>
          </div>

          <!-- Фильтр по дате/времени - день -->
          <div class="filter-group-dialog">
            <label class="checkbox-label-dialog">
              <input type="checkbox" v-model="dialogFilters.dateDayEnabled">
              <span> За определенный день</span>
            </label>
            <div class="date-inputs" v-if="dialogFilters.dateDayEnabled">
              <div class="date-input-group">
                <label>День:</label>
                <input type="date" v-model="dialogFilters.dateDay">
              </div>
            </div>
          </div>

          <!-- Остальные фильтры -->
          <div class="filter-group-dialog">
            <label>Параметры фильтрации:</label>
            <div class="checkbox-filters-dialog">
              <label class="checkbox-label-dialog">
                <input type="checkbox" v-model="dialogFilters.tagFind">
                <span>Тег</span>
                <input type="text" v-model="dialogFilterValues.tagFind" :disabled="!dialogFilters.tagFind" placeholder="TAG_1">
              </label>
              <label class="checkbox-label-dialog">
                <input type="checkbox" v-model="dialogFilters.messFullFind">
                <span>Сообщение</span>
                <input type="text" v-model="dialogFilterValues.messFullFind" :disabled="!dialogFilters.messFullFind" placeholder="Текст сообщения">
              </label>
              <label class="checkbox-label-dialog">
                <input type="checkbox" v-model="dialogFilters.usoTxtFind">
                <span>Диагностика</span>
                <input type="text" v-model="dialogFilterValues.usoTxtFind" :disabled="!dialogFilters.usoTxtFind" placeholder="Текст диагностики">
              </label>
              <label class="checkbox-label-dialog">
                <input type="checkbox" v-model="dialogFilters.severityFind">
                <span>Тревога</span>
                <select v-model="dialogFilterValues.severityFind" :disabled="!dialogFilters.severityFind">
                  <option value="0">Все</option>
                  <option value="901">Неисправность</option>
                  <option value="1001">Пожар</option>
                  <option value="1101">Внимание</option>
                </select>
              </label>
              <label class="checkbox-label-dialog">
                <input type="checkbox" v-model="dialogFilters.kvitFind">
                <span>Квитирование</span>
                <select v-model="dialogFilterValues.kvitFind" :disabled="!dialogFilters.kvitFind">
                  <option value="0">Все</option>
                  <option value="1">Неквитированные</option>
                  <option value="2">Квитированные</option>
                </select>
              </label>
            </div>
          </div>
        </div>

        <div class="filter-txt-status" :class="{ 'has-filters': hasActiveFilters }">
          <span v-if="hasActiveFilters">Активные фильтры: {{ activeFiltersText }}</span>
        </div>

        <div class="filter-dialog-actions">
          <button class="dialog-btn secondary" @click="clearAllFilters">
            Убрать все фильтры
          </button>
          <div class="dialog-action-buttons">
            <button class="dialog-btn secondary" @click="cancelFilters">
              Отмена
            </button>
            <button class="dialog-btn primary" @click="applyFilters">
              Применить
            </button>
          </div>
        </div>
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
  getRowStyle,
  saveAsHTML
} from '@/utils/funcAlarmStore.js'
import { useObjectsStore } from '@/stores/objects'

export default {
  name: 'ScreenHistAlarms',
  setup() {
    const objectsStore = useObjectsStore()
    
    // Состояние интерфейса
    const showTableHeader = ref(true)
    const showAllColumns = ref(false)
    const currentPage = ref(1)
    const totalPages = ref(1)
    const pageInput = ref(1)
    const showFilterDialog = ref(false)
    const showControlDialog = ref(false)

    // Фильтры (активные)
    const filters = ref({
      dateTimeRangeEnabled: false,
      dateDayEnabled: false,
      tagFind: false,
      messFullFind: false,
      usoTxtFind: false,
      severityFind: false,
      kvitFind: false
    })

    const filterValues = ref({
      dateTimeStart: '',
      dateTimeEnd: '',
      dateDay: '',
      tagFind: '',
      messFullFind: '',
      usoTxtFind: '',
      severityFind: 0,
      kvitFind: 0
    })

    // Фильтры для диалога
    const dialogFilters = ref({ ...filters.value })
    const dialogFilterValues = ref({ ...filterValues.value })

    // Данные
    const displayAlarms = computed(() => getAlarmHistMessages.value)

    // Вычисляемые свойства
    const activeFiltersText = computed(() => getCommandData(1))
    const hasActiveFilters = computed(() => activeFiltersText.value.length > 0)

    // Watchers
    watch(displayAlarms, (newAlarms) => {
      if (newAlarms.length > 0) {
        const firstAlarm = newAlarms[0]
        currentPage.value = firstAlarm.current_page || 1
        totalPages.value = firstAlarm.total_pages || 1
        pageInput.value = currentPage.value
      }
    })

    // Методы
    const getCommandData = (type = 0) => {
      const commandData = {
        page_num: currentPage.value
      }

      // Фильтры даты/времени
      if (filters.value.dateTimeRangeEnabled && filterValues.value.dateTimeStart && filterValues.value.dateTimeEnd) {
        commandData.dt_start = new Date(filterValues.value.dateTimeStart).getTime()
        commandData.dt_end = new Date(filterValues.value.dateTimeEnd).getTime()
      }

      if (filters.value.dateDayEnabled && filterValues.value.dateDay) {
        const dayStart = new Date(filterValues.value.dateDay + 'T00:00:00')
        const dayEnd = new Date(filterValues.value.dateDay + 'T23:59:59.999')
        commandData.dt_start = dayStart.getTime()
        commandData.dt_end = dayEnd.getTime()
      }

      // Остальные фильтры
      if (filters.value.tagFind && filterValues.value.tagFind) {
        commandData.tag_find = filterValues.value.tagFind
      }
      if (filters.value.messFullFind && filterValues.value.messFullFind) {
        commandData.mess_full_find = filterValues.value.messFullFind
      }
      if (filters.value.usoTxtFind && filterValues.value.usoTxtFind) {
        commandData.uso_txt_find = filterValues.value.usoTxtFind
      }
      if (filters.value.severityFind && filterValues.value.severityFind > 0) {
        commandData.severity_find = filterValues.value.severityFind
      }
      if (filters.value.kvitFind && filterValues.value.kvitFind > 0) {
        commandData.kvit_find = filterValues.value.kvitFind
      }

      if (type === 0) return commandData

      // Форматирование для отображения
      const formattedData = []
      
      if ((filters.value.dateTimeRangeEnabled || filters.value.dateDayEnabled) && commandData.dt_start && commandData.dt_end) {
        const startDate = new Date(commandData.dt_start).toLocaleString('ru-RU')
        const endDate = new Date(commandData.dt_end).toLocaleString('ru-RU')
        formattedData.push(`period: ${startDate} - ${endDate}`)
      }
      
      if (commandData.tag_find) formattedData.push(`tag: "${commandData.tag_find}"`)
      if (commandData.mess_full_find) formattedData.push(`message: "${commandData.mess_full_find}"`)
      if (commandData.uso_txt_find) formattedData.push(`diagnostic: "${commandData.uso_txt_find}"`)
      
      if (commandData.severity_find && commandData.severity_find > 0) {
        const severityText = getSeverityText(commandData.severity_find)
        formattedData.push(`severity: ${severityText}`)
      }
      
      if (commandData.kvit_find && commandData.kvit_find > 0) {
        const kvitText = commandData.kvit_find === 1 ? 'неквитированные' : 'квитированные'
        formattedData.push(`kvit: ${kvitText}`)
      }

      if (formattedData.length > 0) {
        formattedData.push(`pageNum: ${commandData.page_num}`)
      }
      
      return formattedData.join(' | ')
    }

    const getSeverityText = (severity) => {
      switch (severity) {
        case 901: return 'Неисправность'
        case 1001: return 'Пожар'
        case 1101: return 'Внимание'
        default: return 'Все'
      }
    }

    const refreshData = () => {
      const commandData = getCommandData(0)
      objectsStore.sendCommand('alarms_system', 'command', 'alarms_get_data', commandData)
    }

    const refreshTrend = () => {
      const commandData = {
        id_obj: 1,
        dt_start: 1763530080349,
        dt_end:   1763531754468,
        type: 0,
        limit: 100,
        max_period_days: 30
      }

      objectsStore.sendCommand('alarms_system', 'command', 'trends_get_data', commandData)
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
    }

    const goToPage = (page) => {
      if (page >= 1 && page <= totalPages.value) {
        currentPage.value = page
        pageInput.value = page
        refreshData()
      }
    }

    const goToPageInput = () => {
      goToPage(pageInput.value)
    }

    const applyFilters = () => {
      filters.value = { ...dialogFilters.value }
      filterValues.value = { ...dialogFilterValues.value }
      showFilterDialog.value = false
      currentPage.value = 1
      pageInput.value = 1
      refreshData()
    }

    const cancelFilters = () => {
      dialogFilters.value = { ...filters.value }
      dialogFilterValues.value = { ...filterValues.value }
      showFilterDialog.value = false
    }

    const clearAllFilters = () => {
      dialogFilters.value = {
        dateTimeRangeEnabled: false,
        dateDayEnabled: false,
        tagFind: false,
        messFullFind: false,
        usoTxtFind: false,
        severityFind: false,
        kvitFind: false
      }
      
      dialogFilterValues.value = {
        dateTimeStart: '',
        dateTimeEnd: '',
        dateDay: '',
        tagFind: '',
        messFullFind: '',
        usoTxtFind: '',
        severityFind: 0,
        kvitFind: 0
      }

      filters.value = { ...dialogFilters.value }
      filterValues.value = { ...dialogFilterValues.value }
      showFilterDialog.value = false
      currentPage.value = 1
      pageInput.value = 1
      refreshData()
    }

    const getLocalDateTimeString = (date) => {
      return date.toISOString().slice(0, 16)
    }

    // Инициализация
    onMounted(() => {
      const now = new Date()
      const today = now.toISOString().split('T')[0]
      
      filterValues.value.dateDay = today
      dialogFilterValues.value.dateDay = today
      
      const todayStart = new Date()
      todayStart.setHours(0, 0, 0, 0)
      const todayEnd = new Date()
      todayEnd.setHours(23, 59, 0, 0)
      
      filterValues.value.dateTimeStart = getLocalDateTimeString(todayStart)
      filterValues.value.dateTimeEnd = getLocalDateTimeString(todayEnd)
      dialogFilterValues.value.dateTimeStart = filterValues.value.dateTimeStart
      dialogFilterValues.value.dateTimeEnd = filterValues.value.dateTimeEnd
      
      refreshData()
    })

    return {
      // Состояние
      displayAlarms,
      showTableHeader,
      showAllColumns,
      currentPage,
      totalPages,
      pageInput,
      filters,
      filterValues,
      dialogFilters,
      dialogFilterValues,
      showFilterDialog,
      showControlDialog,
      colorMode,
      
      // Вычисляемые свойства
      activeFiltersText,
      hasActiveFilters,
      
      // Методы
      refreshData,
      toggleTableHeader,
      toggleColumnVisibility,
      handleColorToggle,
      exportData,
      clearData,
      goToPage,
      goToPageInput,
      applyFilters,
      cancelFilters,
      clearAllFilters,
      getRowStyle,
      refreshTrend,
      getCommandData
    }
  }
}
</script>

<style scoped>
@import '@/assets/styles/screen-alarms-hist.css';
</style>