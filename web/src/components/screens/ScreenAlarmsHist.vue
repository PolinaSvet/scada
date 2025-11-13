<template>
  <div class="screen-hist-alarms">

    <!-- Пагинация - перенесена вверх pagination-section top-pagination-->
    <div class="controls-section">

      <div class="pagination-controls">
        <button class="control-btn" @click="refreshData" title="Обновить данные">
          🔄 Обновить
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

        <div class="filter-dialog-actions">
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
              <span>Период даты/времени</span>
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
              <span>За определенный день</span>
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
          <span v-else></span>
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
    const pageInput = ref(1)
    const showFilterDialog = ref(false)
    const showControlDialog = ref(false)

    // Основные фильтры (активные)
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

    // Диалоговые фильтры (для редактирования)
    const dialogFilters = ref({ ...filters.value })
    const dialogFilterValues = ref({ ...filterValues.value })

    // Данные из исторического хранилища
    const displayAlarms = computed(() => getAlarmHistMessages.value)

    // Обновляем пагинацию при изменении данных
    watch(displayAlarms, (newAlarms) => {
      if (newAlarms.length > 0) {
        const firstAlarm = newAlarms[0]
        currentPage.value = firstAlarm.current_page || 1
        totalPages.value = firstAlarm.total_pages || 1
        pageInput.value = currentPage.value
      }
    })

    // Текст активных фильтров для статуса
    const activeFiltersText = computed(() => {
      return getCommandData(1)
    })

    const hasActiveFilters = computed(() => {
      return activeFiltersText.value.length > 0
    })

    // Статусная строка
    const statusText = computed(() => {
      const time = lastUpdateTime.value.toLocaleTimeString('ru-RU')
      const count = displayAlarms.value.length
      return `Обновлено: ${time} | Записей: ${count} | Страница: ${currentPage.value}/${totalPages.value}`
    })

    const getCommandData = (type = 0) => {
      const commandData = {
        pageNum: currentPage.value
      }

      // Добавляем фильтры даты/времени
      if (filters.value.dateTimeRangeEnabled && filterValues.value.dateTimeStart && filterValues.value.dateTimeEnd) {
        commandData.dtStart = new Date(filterValues.value.dateTimeStart).getTime()
        commandData.dtEnd = new Date(filterValues.value.dateTimeEnd).getTime()
      } else if (filters.value.dateDayEnabled && filterValues.value.dateDay) {
        const dayStart = new Date(filterValues.value.dateDay)
        dayStart.setHours(0, 0, 0, 0)
        const dayEnd = new Date(filterValues.value.dateDay)
        dayEnd.setHours(23, 59, 59, 999)
        commandData.dtStart = dayStart.getTime()
        commandData.dtEnd = dayEnd.getTime()
      } /*else {
        // По умолчанию - сегодня
        const todayStart = new Date()
        todayStart.setHours(0, 0, 0, 0)
        const now = Date.now()
        commandData.dtStart = todayStart.getTime()
        commandData.dtEnd = now
      }*/

      // Добавляем остальные активные фильтры
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

      // Форматируем вывод в зависимости от типа
      if (type === 0) {
        // Возвращаем как есть (объект)
        return commandData
      } else {
        // Форматируем в строку в одну линию
        const formattedData = []
        
        if (filters.value.dateTimeRangeEnabled || filters.value.dateDayEnabled){
          if (commandData.dtStart && commandData.dtEnd) {
            const startDate = new Date(commandData.dtStart).toLocaleString('ru-RU')
            const endDate = new Date(commandData.dtEnd).toLocaleString('ru-RU')
            formattedData.push(`period: ${startDate} - ${endDate}`)
          }
        }
        
        if (commandData.tagFind) formattedData.push(`tag: "${commandData.tagFind}"`)
        if (commandData.messFullFind) formattedData.push(`message: "${commandData.messFullFind}"`)
        if (commandData.usoTxtFind) formattedData.push(`diagnostic: "${commandData.usoTxtFind}"`)
        if (commandData.severityFind && commandData.severityFind > 0) {
          const severityText = getSeverityText(commandData.severityFind)
          formattedData.push(`severity: ${severityText}`)
        }
        if (commandData.kvitFind && commandData.kvitFind > 0) {
          const kvitText = commandData.kvitFind === 1 ? 'неквитированные' : 'квитированные'
          formattedData.push(`kvit: ${kvitText}`)
        }

        if (formattedData.length>0){
          formattedData.push(`pageNum: ${commandData.pageNum}`)
        }
        
        return formattedData.join(' | ')
      }
  }

  // Вспомогательная функция для получения текста тревоги
  const getSeverityText = (severity) => {
    switch (severity) {
      case 901: return 'Неисправность'
      case 1001: return 'Пожар'
      case 1101: return 'Внимание'
      default: return 'Все'
    }
  }

    // Функции управления
    const refreshData = () => {
      const commandData = getCommandData(0)
      
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
        pageInput.value = page
        refreshData()
      }
    }

    const goToPageInput = () => {
      goToPage(pageInput.value)
    }

    // Функции для диалога фильтров
    const applyFilters = () => {
      filters.value = { ...dialogFilters.value }
      filterValues.value = { ...dialogFilterValues.value }
      showFilterDialog.value = false
      currentPage.value = 1
      pageInput.value = 1
      refreshData()
    }

    const cancelFilters = () => {
      // Восстанавливаем исходные значения
      dialogFilters.value = { ...filters.value }
      dialogFilterValues.value = { ...filterValues.value }
      showFilterDialog.value = false
    }

    const clearAllFilters = () => {
      // Сбрасываем все чекбоксы и значения
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

      // Применяем очистку сразу
      filters.value = { ...dialogFilters.value }
      filterValues.value = { ...dialogFilterValues.value }
      showFilterDialog.value = false
      currentPage.value = 1
      pageInput.value = 1
      refreshData()
    }

    // Инициализация дат при монтировании
    onMounted(() => {
      // Устанавливаем текущую дату для фильтров
      const now = new Date()
      const today = now.toISOString().split('T')[0]
      
      filterValues.value.dateDay = today
      dialogFilterValues.value.dateDay = today
      
      // Устанавливаем период по умолчанию (сегодня)
      const todayStart = new Date()
      todayStart.setHours(0, 0, 0, 0)
      const todayEnd = new Date()
      todayEnd.setHours(23, 59, 59, 999)
      
      filterValues.value.dateTimeStart = todayStart.toISOString().slice(0, 16)
      filterValues.value.dateTimeEnd = todayEnd.toISOString().slice(0, 16)
      dialogFilterValues.value.dateTimeStart = filterValues.value.dateTimeStart
      dialogFilterValues.value.dateTimeEnd = filterValues.value.dateTimeEnd
      
      refreshData()
    })

    return {
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
      colorMode,
      statusText,
      activeFiltersText,
      hasActiveFilters,
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
      formatTime,
      getRowStyle,
      showControlDialog,
      getCommandData
    }
  }
}
</script>

<style scoped>
/* Основные стили вынесены в отдельный CSS файл */
@import '@/assets/styles/screen-alarms-hist.css';

/* Дополнительные стили для новых элементов */


.top-pagination {
  margin-bottom: 15px;
  padding: 10px;
  background: #f5f5f5;
  border-radius: 4px;
}

.pagination-controls {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.page-input-group {
  display: flex;
  align-items: center;
  gap: 5px;
  color: #495057;
}

.page-input {
  width: 60px;
  padding: 4px 8px;
  border: 1px solid #ddd;
  border-radius: 3px;
  text-align: center;
}

.filter-txt-color {
  color: #495057;
}

.filters-section-compact {
  margin-bottom: 15px;
}

.filter-controls {
  display: flex;
  align-items: center;
  gap: 15px;
}

.filter-main-btn {
  padding: 8px 16px;
  background: #007bff;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
}

.filter-main-btn:hover {
  background: #0056b3;
}

.filter-txt-status {
  padding: 6px 6px;
  font-size: 14px;
  width: 100%;
  position: relative;
  overflow: hidden;
  text-overflow: ellipsis;
  word-wrap: break-word;
  background: #f6f7f8;
  color: #000000;
}

.filter-status {
  padding: 6px 12px;
  background: #e9ecef;
  border-radius: 4px;
  font-size: 14px;
  color: #495057;
  min-width: 100px;
  width: 200px;
  cursor: help;
  position: relative;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* Кастомный тултип через data-атрибут */
.filter-status:hover::after {
  content: attr(data-tooltip);
  position: absolute;
  bottom: 100%;
  left: 50%;
  transform: translateX(-50%);
  background: rgba(0, 0, 0, 0.9);
  color: white;
  padding: 8px 12px;
  border-radius: 4px;
  font-size: 12px;
  white-space: normal;
  width: 300px;
  z-index: 1000;
  pointer-events: none;
  margin-bottom: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
  line-height: 1.4;
}

.filter-status:not(.has-filters) {
  background: #e9ecef;
  color: #6c757d;
  border: 1px solid #dee2e6;
}

.filter-status.has-filters {
  background: #9fcdff;
  color: #000000;
  border: 1px solid #007bff;
  font-weight: 500;
}

/* Стили для диалогового окна фильтров */
.filter-dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.filter-dialog {
  background: white;
  border-radius: 8px;
  width: 90%;
  max-width: 600px;
  max-height: 80vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
}

.filter-dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid #e9ecef;
  background: #f8f9fa;
  border-radius: 8px 8px 0 0;
}

.filter-dialog-header h3 {
  margin: 0;
  color: #495057;
}

.close-btn {
  background: none;
  border: none;
  font-size: 24px;
  cursor: pointer;
  color: #6c757d;
  padding: 0;
  width: 30px;
  height: 30px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.close-btn:hover {
  color: #495057;
  background: #e9ecef;
  border-radius: 50%;
}

.filter-dialog-content {
  padding: 20px;
  overflow-y: auto;
  flex: 1;
}

.filter-group-dialog {
  margin-bottom: 20px;
}

.filter-group-dialog label {
  display: block;
  margin-bottom: 10px;
  font-weight: 500;
  color: #495057;
}

.checkbox-label-dialog {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 10px;
  font-weight: normal;
}

.checkbox-label-dialog input[type="checkbox"] {
  margin: 0;
}

.checkbox-label-dialog span {
  min-width: 120px;
}

.date-time-inputs,
.date-inputs {
  margin-left: 25px;
  margin-top: 10px;
}

.date-input-group {
  margin-bottom: 10px;
  display: flex;
  align-items: center;
  gap: 10px;
}

.date-input-group label {
  min-width: 80px;
  margin: 0;
  font-weight: normal;
}

.date-input-group input {
  padding: 6px 10px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
}

.checkbox-filters-dialog {
  margin-left: 0;
}

.checkbox-filters-dialog .checkbox-label-dialog {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 12px;
}

.checkbox-filters-dialog input[type="text"],
.checkbox-filters-dialog select {
  flex: 1;
  padding: 6px 10px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
}

.filter-dialog-actions {
  display: flex;
  justify-content: space-between;
  padding: 16px 20px;
  border-top: 1px solid #e9ecef;
  background: #f8f9fa;
  border-radius: 0 0 8px 8px;
}

.dialog-action-buttons {
  display: flex;
  gap: 10px;
}

.dialog-btn {
  padding: 8px 16px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  transition: background-color 0.2s;
}

.dialog-btn.primary {
  background: #007bff;
  color: white;
}

.dialog-btn.primary:hover {
  background: #0056b3;
}

.dialog-btn.secondary {
  background: #6c757d;
  color: white;
}

.dialog-btn.secondary:hover {
  background: #545b62;
}

/* Адаптивность */
@media (max-width: 768px) {
  .filter-dialog {
    width: 95%;
    margin: 10px;
  }
  
  .date-input-group {
    flex-direction: column;
    align-items: flex-start;
  }
  
  .filter-dialog-actions {
    flex-direction: column;
    gap: 10px;
  }
  
  .dialog-action-buttons {
    width: 100%;
  }
  
  .dialog-btn {
    flex: 1;
  }

  
}
</style>