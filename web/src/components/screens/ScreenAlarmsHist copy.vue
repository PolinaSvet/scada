<template>
  <div class="screen-hist-alarms">
    <div class="screen-header">
      <h1>📊 ИСТОРИЯ СОБЫТИЙ</h1>
      <p>Исторические данные алармов и сообщений системы</p>
    </div>

    <!-- Управление и фильтры -->
    <div class="controls-section">
      <div class="control-buttons">
        <button class="control-btn" @click="refreshData" title="Обновить данные">
          🔄 Обновить
        </button>
        <button class="control-btn" @click="toggleTableHeader" title="Переключить заголовок таблицы">
          {{ showTableHeader ? '📋' : '📄' }} Заголовок
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

    <!-- Основная таблица -->
    <div class="table-section">
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
            <td class="time-cell">{{ alarm.dt_txt || formatTime(alarm.dt) }}</td>
            <td class="tag-cell">{{ alarm.tag || '-' }}</td>
            <td class="desc-cell">{{ alarm.mess_name || '-' }}</td>
            <td class="message-cell">{{ alarm.mess_full || '-' }}</td>
            <td class="uso-cell">{{ alarm.mess_state || '-' }}</td>
            <td class="type-cell">{{ alarm.severity }}</td>
            <td class="type-cell">{{ alarm.opermess }}</td>
            <td class="kvit-cell">{{ alarm.kvit ? '✅' : '❌' }}</td>
          </tr>
        </tbody>
      </table>
      
      <!-- Сообщение если нет данных -->
      <div v-if="displayAlarms.length === 0" class="no-data-message">
        <p>📭 Нет исторических данных для отображения</p>
        <p>Используйте кнопку "Обновить" для загрузки данных</p>
      </div>
    </div>

    <!-- Пагинация и информация -->
    <div class="pagination-section" v-if="displayAlarms.length > 0">
      <div class="pagination-info">
        Показано записей: {{ displayAlarms.length }} 
        <span v-if="currentPage && totalPages"> | Страница {{ currentPage }} из {{ totalPages }}</span>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, computed, onMounted } from 'vue'
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
    const lastUpdateTime = ref(new Date())
    const currentPage = ref(1)
    const totalPages = ref(1)

    // Данные из исторического хранилища
    const displayAlarms = computed(() => getAlarmHistMessages.value)

    // Функции управления
    const refreshData = () => {
      const filterParams = {
        dtStart: Date.now() - 24 * 60 * 60 * 1000, // последние 24 часа
        dtEnd: Date.now(),
        pageNum: currentPage.value
      }

      const commandData = {
        dtStart: filterParams.dtStart || null,
        dtEnd: filterParams.dtEnd || null,
        tagFind: filterParams.tagFind || '',
        messFullFind: filterParams.messFullFind || '',
        usoTxtFind: filterParams.usoTxtFind || '',
        severityFind: filterParams.severityFind || 0,
        opermessFind: filterParams.opermessFind || 0,
        kvitFind: filterParams.kvitFind || 0,
        pageNum: filterParams.pageNum || 1
      }

      objectsStore.sendCommand(
        'alarms_system',
        'alarms_get_data',
        'command_historian',
        commandData
      )
      
      lastUpdateTime.value = new Date()
    }

    const toggleTableHeader = () => {
      showTableHeader.value = !showTableHeader.value
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

    // При монтировании загружаем данные
    onMounted(() => {
      refreshData()
    })

    return {
      displayAlarms,
      showTableHeader,
      currentPage,
      totalPages,
      colorMode,
      refreshData,
      toggleTableHeader,
      handleColorToggle,
      exportData,
      clearData,
      formatTime,
      getRowStyle
    }
  }
}
</script>

<style scoped>
.screen-hist-alarms {
  padding: 20px;
  height: 100%;
  display: flex;
  flex-direction: column;
  background-color: #f5f5f5;
}

.screen-header {
  background: white;
  padding: 20px;
  border-radius: 8px;
  margin-bottom: 20px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

.screen-header h1 {
  margin: 0 0 8px 0;
  color: #2c3e50;
  font-size: 24px;
}

.screen-header p {
  margin: 0 0 16px 0;
  color: #7f8c8d;
}

.header-stats {
  display: flex;
  gap: 20px;
  font-weight: 500;
}

.total-messages {
  color: #3498db;
}

.alarm-count {
  color: #e74c3c;
}

.error-count {
  color: #f39c12;
}

.controls-section {
  background: white;
  padding: 15px 20px;
  border-radius: 8px;
  margin-bottom: 20px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
  display: flex;
  justify-content: between;
  align-items: center;
  gap: 20px;
}

.control-buttons {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.control-btn {
  padding: 8px 16px;
  border: 1px solid #ddd;
  border-radius: 4px;
  background: white;
  cursor: pointer;
  transition: all 0.3s;
  font-size: 14px;
}

.control-btn:hover {
  background: #3498db;
  color: white;
  border-color: #3498db;
}

.status-text {
  color: #7f8c8d;
  font-size: 14px;
  flex-grow: 1;
  text-align: right;
}

.table-section {
  flex-grow: 1;
  background: white;
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

.alarms-table {
  width: 100%;
  border-collapse: collapse;
}

.alarms-table th {
  background: #34495e;
  color: white;
  padding: 12px 8px;
  text-align: left;
  font-weight: 600;
  border-bottom: 2px solid #2c3e50;
}

.alarms-table td {
  padding: 8px;
  border-bottom: 1px solid #ecf0f1;
}

.alarms-table tr:hover {
  background-color: #f8f9fa;
}

.number-cell {
  width: 60px;
  text-align: center;
  font-weight: 600;
}

.id-cell {
  width: 80px;
  text-align: center;
}

.time-cell {
  width: 160px;
  white-space: nowrap;
}

.tag-cell {
  width: 120px;
}

.desc-cell {
  width: 200px;
}

.message-cell {
  min-width: 250px;
  max-width: 400px;
}

.uso-cell {
  width: 150px;
}

.type-cell {
  width: 60px;
  text-align: center;
}

.kvit-cell {
  width: 60px;
  text-align: center;
}

.no-data-message {
  padding: 60px 20px;
  text-align: center;
  color: #7f8c8d;
  font-size: 16px;
}

.no-data-message p {
  margin: 10px 0;
}

.pagination-section {
  background: white;
  padding: 15px 20px;
  border-radius: 8px;
  margin-top: 20px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

.pagination-info {
  color: #7f8c8d;
  font-size: 14px;
}

/* Адаптивность */
@media (max-width: 1200px) {
  .controls-section {
    flex-direction: column;
    align-items: stretch;
  }
  
  .status-text {
    text-align: left;
  }
}
</style>