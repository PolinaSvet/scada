<template>
  <div class="screen-hist-trends">
    <!-- Панель управления -->
    <div class="controls-section">
      <div class="pagination-controls">
        <button class="control-btn" @click="refreshData" title="Обновить данные">
          🔄 Обновить
        </button>
        <button class="control-btn" @click="clearAllTrends" title="Очистить все тренды">
          🗑️ Очистить
        </button>
        <button class="control-btn" @click="showFilterDialog = true">
          ⚙️ Фильтры
        </button>
        <div class="filter-status" 
            :class="{ 'has-filters': hasActiveFilters }"
            :title="hasActiveFilters ? activeFiltersText : 'Фильтры не применены'">
          <span v-if="hasActiveFilters">Активные фильтры: {{ activeFiltersText }}</span>
          <span v-else>Фильтры не применены</span>
        </div>
      </div>

      <div class="pagination-controls">
        <div class="trends-info">
          <span>Активные тренды: {{ activeTrendsCount }} / {{ maxTrends }}</span>
          <span>Точек: {{ totalPoints }}</span>
        </div>
      </div>

      <div class="pagination-controls">
        <button class="control-btn" @click="showControlDialog = true">
          🎛️ Управление
        </button>
        <button class="control-btn" @click="resetZoom" title="Сбросить масштаб">
          🔍 Сброс zoom
        </button>
      </div>
    </div>

    <!-- Область графика -->
    <div class="chart-container">
      <div ref="chartContainer" class="chart-area">
        <canvas v-show="hasChartData" ref="chartCanvas" class="trend-chart"></canvas>
        <div v-if="!hasChartData" class="no-chart-data">
          <p>📊 Нет данных для отображения графика</p>
          <p>Настройте фильтры и обновите данные</p>
        </div>
      </div>
    </div>

    <!-- Список активных трендов -->
    <div class="active-trends-section">
      <h3>Активные тренды ({{ activeTrendsCount }})</h3>
      <div class="trends-list">
        <div 
          v-for="(trend, index) in activeTrends" 
          :key="trend.id_obj"
          class="trend-item"
          :class="{ 'trend-visible': trend.visible !== false }"
        >
          <div class="trend-info">
            <div class="color-indicator" :style="{ backgroundColor: getTrendColor(index) }"></div>
            <span class="trend-id">ID: {{ trend.id_obj }}</span>
            <span class="trend-points">Точек: {{ trend.points ? trend.points.length : 0 }}</span>
            <span class="trend-stats" v-if="getTrendStats(trend.id_obj)">
              Min: {{ getTrendStats(trend.id_obj).min.toFixed(2) }}, 
              Max: {{ getTrendStats(trend.id_obj).max.toFixed(2) }}
            </span>
          </div>
          <div class="trend-actions">
            <button 
              class="trend-action-btn"
              @click="toggleTrendVisibility(trend.id_obj)"
              :title="trend.visible !== false ? 'Скрыть тренд' : 'Показать тренд'"
            >
              {{ trend.visible !== false ? '👁️' : '👁️‍🗨️' }}
            </button>
            <button 
              class="trend-action-btn"
              @click="removeTrend(trend.id_obj)"
              title="Удалить тренд"
            >
              ❌
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Диалоговое окно управления -->
    <div v-if="showControlDialog" class="filter-dialog-overlay" @click="showControlDialog = false">
      <div class="filter-dialog" @click.stop>
        <div class="filter-dialog-header">
          <h3>Управление трендами</h3>
          <button class="close-btn" @click="showControlDialog = false">×</button>
        </div>

        <div class="controls-section">
          <div class="pagination-controls">
            <button class="control-btn" @click="clearAllTrends" title="Очистить все тренды">
              🗑️ Очистить все
            </button>
            <button class="control-btn" @click="resetZoom" title="Сбросить масштаб графика">
              🔍 Сброс zoom
            </button>
            <button class="control-btn" @click="exportChartData" title="Экспортировать данные графика">
              📥 Экспорт данных
            </button>
          </div>
          <div class="pagination-controls">
            <button class="control-btn" @click="toggleLabels" :title="showLabels ? 'Скрыть подписи' : 'Показать подписи'">
              {{ showLabels ? '📝 Скрыть labels' : '📝 Показать labels' }}
            </button>
            <button class="control-btn" @click="toggleInteractionMode" :title="interactionMode === 'nearest' ? 'Режим: ближайшая точка' : 'Режим: пересечение'">
              {{ interactionMode === 'nearest' ? '🎯 Nearest' : '⚡ Intersect' }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Диалоговое окно фильтров -->
    <div v-if="showFilterDialog" class="filter-dialog-overlay" @click="showFilterDialog = false">
      <div class="filter-dialog filter-dialog-fixed" @click.stop>
        <div class="filter-dialog-header">
          <h3>Фильтры трендов</h3>
          <button class="close-btn" @click="showFilterDialog = false">×</button>
        </div>
        
        <div class="filter-dialog-content">
          <!-- Период выборки -->
          <div class="filter-group-dialog">
            <label class="filter-label-dialog">Период выборки:</label>
            <div class="date-time-inputs">
              <div class="date-input-group">
                <label>Начало:</label>
                <input type="datetime-local" v-model="dialogFilters.dt_start">
              </div>
              <div class="date-input-group">
                <label>Конец:</label>
                <input type="datetime-local" v-model="dialogFilters.dt_end">
              </div>
            </div>
          </div>

          <!-- Список ID объектов для трендов -->
          <div class="filter-group-dialog">
            <label class="filter-label-dialog">ID объектов (до {{ maxTrends }}):</label>
            <div class="id-objects-inputs">
              <div 
                v-for="(idObj, index) in dialogFilters.id_objects" 
                :key="index"
                class="id-input-group"
              >
                <input 
                  type="number" 
                  v-model.number="dialogFilters.id_objects[index]"
                  placeholder="ID объекта"
                  :min="1"
                  @change="validateIdObject(index)"
                >
                <button 
                  v-if="dialogFilters.id_objects.length > 1"
                  class="remove-id-btn"
                  @click="removeIdObject(index)"
                  title="Удалить ID"
                >
                  ×
                </button>
              </div>
              <button 
                v-if="dialogFilters.id_objects.length < maxTrends"
                class="add-id-btn"
                @click="addIdObject"
              >
                + Добавить ID
              </button>
              <div v-if="duplicateIdError" class="error-message">
                ❌ ID {{ duplicateIdError }} уже существует в списке
              </div>
            </div>
          </div>

          <!-- Дополнительные параметры -->
          <div class="filter-group-dialog">
            <label class="filter-label-dialog">Параметры выборки:</label>
            <div class="sampling-params">
              <div class="param-input">
                <label>Тип агрегации:</label>
                <select v-model="dialogFilters.type">
                  <option value="0">Первые N записей</option>
                  <option value="1">Среднее: берем каждую N-ю запись</option>
                  <option value="2">Мин-Макс: в каждом интервале берем мин и макс</option>
                  <option value="3">Минимум: в каждом интервале берем только минимум</option>
                  <option value="4">Максимум: в каждом интервале берем только максимум</option>
                </select>
              </div>
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
            <button class="dialog-btn primary" @click="applyFilters" :disabled="hasDuplicateIds">
              Применить
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, computed, onMounted, watch, nextTick, onUnmounted } from 'vue'
import { 
  getActiveTrends,
  getAllTrends,
  getChartData,
  clearTrendsHistStore,
  removeTrendByIdObj,
  toggleTrendVisibility,
  setCurrentPeriod,
  getTrendStats,
  getTrendColor,
  getActiveTrendsCount,
  getTotalPoints,
  maxTrends,
  getFilters,
  setFilters,
  getDefaultFilters,
  getTypeText,
  exportChartDataToHTML
} from '@/stores/trendStoreHist.js'
import { useObjectsStore } from '@/stores/objects'

// Chart.js импорты
import Chart from 'chart.js/auto'
import 'chartjs-adapter-date-fns'
import { ru } from 'date-fns/locale'
import zoomPlugin from 'chartjs-plugin-zoom'

Chart.register(zoomPlugin)

export default {
  name: 'ScreenHistTrends',
  setup() {
    const objectsStore = useObjectsStore()
    
    // Состояние интерфейса
    const showControlDialog = ref(false)
    const showFilterDialog = ref(false)
    const chartCanvas = ref(null)
    const chartInstance = ref(null)
    const isChartInitialized = ref(false)
    const duplicateIdError = ref(null)
    const showLabels = ref(true)
    const interactionMode = ref('nearest') // 'nearest' или 'index'

    // Фильтры из хранилища
    const filters = ref(getFilters())
    const dialogFilters = ref({ ...filters.value })

    // Данные из хранилища
    const activeTrends = computed(() => getActiveTrends.value)
    const chartData = computed(() => getChartData.value)
    const activeTrendsCount = computed(() => getActiveTrendsCount.value)
    const totalPoints = computed(() => getTotalPoints.value)

    // Вычисляемые свойства
    const hasChartData = computed(() => {
      return chartData.value.length > 0 && 
             chartData.value.some(series => series.data && series.data.length > 0) &&
             totalPoints.value > 0
    })
    
    const activeFiltersText = computed(() => {
      const parts = []
      const f = filters.value

      console.log("filters.value",filters.value)
      
      if (f.dt_start && f.dt_end) {
        const start = new Date(f.dt_start).toLocaleString('ru-RU')
        const end = new Date(f.dt_end).toLocaleString('ru-RU')
        parts.push(`период: ${start} - ${end}`)
      }
      
      if (f.id_objects.length > 0) {
        parts.push(`ID: ${f.id_objects.join(', ')}`)
      }
      
      //if (f.type > 0) {
      parts.push(`тип: ${getTypeText(f.type)}`)
      //}
      
      return parts.join(' | ')
    })

    const hasActiveFilters = computed(() => {
      const f = filters.value
      return f.dt_start !== '' && f.dt_end !== '' && f.id_objects.length > 0
    })

    const hasDuplicateIds = computed(() => {
      const ids = dialogFilters.value.id_objects
      return new Set(ids).size !== ids.length
    })

    // Методы
    const getCommandData = (id_obj) => ({
      id_obj: id_obj,
      dt_start: new Date(filters.value.dt_start).getTime(),
      dt_end: new Date(filters.value.dt_end).getTime(),
      type: filters.value.type
    })

    const refreshData = () => {
      if (!hasActiveFilters.value) {
        console.warn('⚠️ Не заданы параметры фильтрации')
        return
      }

      setCurrentPeriod(
        new Date(filters.value.dt_start).getTime(),
        new Date(filters.value.dt_end).getTime()
      )

      filters.value.id_objects.forEach(id_obj => {
        const commandData = getCommandData(id_obj)
        objectsStore.sendCommand('alarms_system', 'command', 'trends_get_data', commandData)
      })
    }

    const addIdObject = () => {
      if (dialogFilters.value.id_objects.length < maxTrends) {
        dialogFilters.value.id_objects.push(1)
      }
    }

    const removeIdObject = (index) => {
      if (dialogFilters.value.id_objects.length > 1) {
        dialogFilters.value.id_objects.splice(index, 1)
        duplicateIdError.value = null
      }
    }

    const validateIdObject = (index) => {
      const currentId = dialogFilters.value.id_objects[index]
      const duplicates = dialogFilters.value.id_objects.filter((id, i) => id === currentId && i !== index)
      duplicateIdError.value = duplicates.length > 0 ? currentId : null
    }

    const removeTrend = (id_obj) => {
      removeTrendByIdObj(id_obj)
      nextTick(updateChart)
    }

    const toggleTrendVisibility = (id_obj) => {
      toggleTrendVisibility(id_obj)
      nextTick(updateChart)
    }

    const clearAllTrends = () => {
      clearTrendsHistStore()
      destroyChart()
    }

    const toggleInteractionMode = () => {
      interactionMode.value = interactionMode.value === 'nearest' ? 'index' : 'nearest'
      if (chartInstance.value) {
        updateChartInteraction()
      }
    }

    const updateChartInteraction = () => {
      if (chartInstance.value) {
        chartInstance.value.options.interaction.mode = interactionMode.value
        chartInstance.value.update()
      }
    }

    const exportChartData = () => {
      const files = exportChartDataToHTML(chartData.value,activeFiltersText.value)
      
      files.forEach(file => {
        const blob = new Blob([file.content], { type: 'text/html' })
        const url = URL.createObjectURL(blob)
        const a = document.createElement('a')
        a.href = url
        a.download = file.filename
        document.body.appendChild(a)
        a.click()
        document.body.removeChild(a)
        URL.revokeObjectURL(url)
      })
      
      if (files.length === 0) {
        alert('Нет данных для экспорта')
      } else {
        alert(`Экспортировано ${files.length} файлов`)
      }
    }

    const applyFilters = () => {
      if (hasDuplicateIds.value) return
      setFilters(dialogFilters.value)
      filters.value = getFilters()
      showFilterDialog.value = false
      refreshData()
    }

    const cancelFilters = () => {
      dialogFilters.value = { ...filters.value }
      duplicateIdError.value = null
      showFilterDialog.value = false
    }

    const clearAllFilters = () => {
      dialogFilters.value = getDefaultFilters()
      duplicateIdError.value = null
    }

    const resetZoom = () => {
      chartInstance.value?.resetZoom()
    }

    const toggleLabels = () => {
      showLabels.value = !showLabels.value
      updateChart()
    }

    const destroyChart = () => {
      if (chartInstance.value) {
        chartInstance.value.destroy()
        chartInstance.value = null
        isChartInitialized.value = false
      }
    }

    // График
    const initChart = () => {
      if (!chartCanvas.value || !hasChartData.value || isChartInitialized.value) return

      try {
        const ctx = chartCanvas.value.getContext('2d')
        if (!ctx) return

        const timeFormat = determineTimeFormat()
        
        chartInstance.value = new Chart(ctx, {
          type: 'line',
          data: { datasets: prepareChartData() },
          options: {
            responsive: true,
            maintainAspectRatio: false,
            interaction: { 
              mode: interactionMode.value, 
              intersect: interactionMode.value === 'index' 
            },
            plugins: {
              legend: {
                display: showLabels.value,
                position: 'top',
                labels: { usePointStyle: true, boxWidth: 10, padding: 15 },
                onClick: (e, legendItem, legend) => {
                  const index = legendItem.datasetIndex
                  const chart = legend.chart
                  const meta = chart.getDatasetMeta(index)
                  meta.hidden = meta.hidden === null ? !chart.data.datasets[index].hidden : null
                  chart.update()
                }
              },
              tooltip: {
                mode: 'index',
                intersect: false,
                callbacks: {
                  title: (items) => {
                    if (items.length === 0) return ''
                    const date = new Date(items[0].parsed.x)
                    return date.toLocaleString('ru-RU', {
                      year: '2-digit', month: '2-digit', day: '2-digit',
                      hour: '2-digit', minute: '2-digit', second: '2-digit',
                      fractionalSecondDigits: 3
                    })
                  },
                  label: (context) => {
                    const label = context.dataset.label || ''
                    const value = context.parsed.y !== null ? context.parsed.y.toFixed(4) : 'N/A'
                    const quality = context.raw?.quality || 'N/A'
                    return [`${label}: ${value}`, `Качество: ${quality}`]
                  }
                }
              },
              zoom: {
                pan: { enabled: true, mode: 'x', modifierKey: 'ctrl' },
                zoom: { wheel: { enabled: true }, pinch: { enabled: true }, mode: 'x' }
              }
            },
            scales: {
              x: {
                type: 'time',
                time: {
                  unit: timeFormat.unit,
                  displayFormats: timeFormat.displayFormats,
                  tooltipFormat: 'yy.MM.dd HH:mm:ss.SSS'
                },
                adapters: { date: { locale: ru } },
                title: { display: showLabels.value, text: 'Время' },
                ticks: { maxRotation: 45, minRotation: 30, autoSkip: true, maxTicksLimit: 10 }
              },
              y: {
                type: 'linear',
                title: { display: showLabels.value, text: 'Значение' },
                ticks: { callback: value => typeof value === 'number' ? value.toFixed(2) : value }
              }
            },
            animation: { duration: 0 },
            elements: { point: { radius: 0, hoverRadius: 3 } }
          }
        })

        isChartInitialized.value = true
      } catch (error) {
        console.error('❌ Error initializing chart:', error)
        isChartInitialized.value = false
      }
    }

    const prepareChartData = () => {
      return chartData.value
        .filter(series => series.data && series.data.length > 0)
        .map((series, index) => {
          const validData = series.data
            .filter(point => point && typeof point.x === 'number' && typeof point.y === 'number' && !isNaN(point.x) && !isNaN(point.y))
            .sort((a, b) => a.x - b.x)

          return {
            label: showLabels.value ? `ID: ${series.id_obj}` : '',
            data: validData,
            borderColor: series.color,
            backgroundColor: series.color + '20',
            borderWidth: 2,
            borderDash: validData.some(p => p.quality !== 192) ? [5, 5] : [],
            pointRadius: 0,
            pointHoverRadius: 3,
            pointBackgroundColor: series.color,
            fill: false,
            tension: 0.1,
            spanGaps: true
          }
        })
    }

    const determineTimeFormat = () => {
      if (!chartData.value.length) return { unit: 'minute', displayFormats: { minute: 'HH:mm' } }
      
      const allTimes = chartData.value
        .flatMap(series => series.data ? series.data.map(point => point.x) : [])
        .filter(time => typeof time === 'number' && !isNaN(time))
        .sort((a, b) => a - b)
      
      if (allTimes.length < 2) return { unit: 'minute', displayFormats: { minute: 'HH:mm' } }
      
      const range = allTimes[allTimes.length - 1] - allTimes[0]
      const oneDay = 24 * 60 * 60 * 1000
      
      return range <= oneDay ? {
        unit: 'minute', displayFormats: { minute: 'HH:mm', hour: 'HH:mm' }
      } : range <= 7 * oneDay ? {
        unit: 'day', displayFormats: { minute: 'dd.MM HH:mm', hour: 'dd.MM HH:mm', day: 'dd.MM' }
      } : {
        unit: 'day', displayFormats: { day: 'dd.MM.yy', week: 'dd.MM.yy' }
      }
    }

    const updateChart = () => {
      if (chartInstance.value && hasChartData.value && isChartInitialized.value) {
        try {
          chartInstance.value.data.datasets = prepareChartData()
          chartInstance.value.options.plugins.legend.display = showLabels.value
          chartInstance.value.options.scales.x.title.display = showLabels.value
          chartInstance.value.options.scales.y.title.display = showLabels.value
          chartInstance.value.options.interaction.mode = interactionMode.value
          chartInstance.value.options.interaction.intersect = interactionMode.value === 'index'
          chartInstance.value.update('none')
        } catch (error) {
          console.error('❌ Error updating chart:', error)
          destroyChart()
          nextTick(initChart)
        }
      } else if (!hasChartData.value && chartInstance.value) {
        destroyChart()
      } else if (hasChartData.value && !isChartInitialized.value) {
        nextTick(initChart)
      }
    }

    // Watchers
    watch(chartData, () => nextTick(updateChart), { deep: true })
    watch(hasChartData, (newVal) => {
      if (newVal && !isChartInitialized.value) nextTick(initChart)
      else if (!newVal && isChartInitialized.value) destroyChart()
    })

    // Инициализация
    onMounted(() => {
      const defaultFilters = getDefaultFilters()
      setFilters(defaultFilters)
      filters.value = getFilters()
      dialogFilters.value = { ...filters.value }
      
      nextTick(() => { if (hasChartData.value) initChart() })
    })

    onUnmounted(destroyChart)

    return {
      showControlDialog, showFilterDialog, chartCanvas, filters, dialogFilters,
      duplicateIdError, showLabels, activeTrends, activeTrendsCount, totalPoints,
      hasChartData, activeFiltersText, hasActiveFilters, hasDuplicateIds, maxTrends,
      refreshData, addIdObject, removeIdObject, validateIdObject, removeTrend,
      toggleTrendVisibility, clearAllTrends, exportChartData, applyFilters,
      cancelFilters, clearAllFilters, getTrendStats, getTrendColor, resetZoom, toggleLabels,interactionMode,
      toggleInteractionMode
    }
  }
}
</script>

<style scoped>
@import '@/assets/styles/screen-trends-hist.css';
</style>