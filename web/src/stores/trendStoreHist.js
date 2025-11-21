import { ref, computed } from 'vue'

// Хранилище для исторических данных трендов
const trendStoreHist = ref([])
const currentPeriod = ref({ dt_start: null, dt_end: null })
const filters = ref({
  dt_start: '',
  dt_end: '',
  id_objects: [1],
  type: 0
})

// Константы
export const maxTrends = 10

// Цвета для трендов
const trendColors = [
  '#FF6384', '#36A2EB', '#FFCE56', '#4BC0C0', '#9966FF',
  '#FF9F40', '#8AC926', '#1982C4', '#6A4C93', '#FF595E'
]

// Функции для работы с фильтрами
export const getFilters = () => filters.value
export const setFilters = (newFilters) => {
  filters.value = { ...newFilters }
}

export const getDefaultFilters = () => {
  const todayStart = new Date()
  todayStart.setHours(0, 0, 0, 0)
  const todayEnd = new Date()
  todayEnd.setHours(23, 59, 0, 0)
  
  return {
    dt_start: getLocalDateTimeString(todayStart),
    dt_end: getLocalDateTimeString(todayEnd),
    id_objects: [1],
    type: "0"
  }
}

const getLocalDateTimeString = (date) => {
  return date.toISOString().slice(0, 16)
}

export const addTrendsHistBatch = (objectsArray) => {
  if (!Array.isArray(objectsArray) || objectsArray.length === 0) return
  
  const firstPoint = objectsArray[0]
  const targetIdObj = firstPoint.id_obj
  
  const pointsWithIds = objectsArray.map(point => ({
    ...point,
    uniqueId: `trend_${point.id_obj}_${point.dt}_${point.id}_${Math.random().toString(36).substr(2, 9)}`
  }))
  
  const existingIndex = trendStoreHist.value.findIndex(item => item.id_obj === targetIdObj)
  
  if (existingIndex !== -1) {
    trendStoreHist.value[existingIndex].points = pointsWithIds
    console.log(`🔄 Trend ${targetIdObj} updated: ${pointsWithIds.length} points`)
  } else {
    if (trendStoreHist.value.length < maxTrends) {
      trendStoreHist.value.push({
        id_obj: targetIdObj,
        points: pointsWithIds,
        visible: true
      })
      console.log(`✅ Trend ${targetIdObj} added: ${pointsWithIds.length} points`)
    } else {
      console.warn(`❌ Cannot add trend ${targetIdObj}: maximum ${maxTrends} trends reached`)
    }
  }
}

export const clearTrendsHistStore = () => {
  trendStoreHist.value = []
  currentPeriod.value = { dt_start: null, dt_end: null }
  console.log('🗑️ Trend history store cleared')
}

export const setCurrentPeriod = (dt_start, dt_end) => {
  if (currentPeriod.value.dt_start !== dt_start || currentPeriod.value.dt_end !== dt_end) {
    clearTrendsHistStore()
  }
  currentPeriod.value = { dt_start, dt_end }
}

export const removeTrendByIdObj = (id_obj) => {
  const index = trendStoreHist.value.findIndex(item => item.id_obj === id_obj)
  if (index !== -1) {
    trendStoreHist.value.splice(index, 1)
    console.log(`🗑️ Trend with id_obj ${id_obj} removed`)
  }
}

export const toggleTrendVisibility = (id_obj) => {
  const trend = trendStoreHist.value.find(item => item.id_obj === id_obj)
  if (trend) {
    trend.visible = !trend.visible
    console.log(`👁️ Trend ${id_obj} visibility: ${trend.visible}`)
  }
}

export const toggleAllTrendsVisibility = () => {
  const allVisible = trendStoreHist.value.every(trend => trend.visible)
  trendStoreHist.value.forEach(trend => {
    trend.visible = !allVisible
  })
  console.log(`👁️ All trends visibility: ${!allVisible}`)
}

export const getActiveTrends = computed(() => {
  return trendStoreHist.value.filter(trend => trend.visible !== false)
})

export const getAllTrends = computed(() => {
  return trendStoreHist.value.map(trend => ({
    id_obj: trend.id_obj,
    point_count: trend.points ? trend.points.length : 0,
    visible: trend.visible !== false
  }))
})

export const getChartData = computed(() => {
  const chartData = []
  
  trendStoreHist.value.forEach((trend, index) => {
    if (trend.visible !== false && trend.points) {
      const series = {
        id_obj: trend.id_obj,
        data: trend.points.map(point => ({
          x: point.dt,
          y: point.value,
          quality: point.quality
        })).sort((a, b) => a.x - b.x),
        visible: true,
        color: trendColors[index % trendColors.length]
      }
      chartData.push(series)
    }
  })
  
  return chartData
})

export const getTrendStats = (id_obj) => {
  const trend = trendStoreHist.value.find(t => t.id_obj === id_obj)
  if (!trend || !trend.points || trend.points.length === 0) return null
  
  const values = trend.points.map(p => p.value).filter(v => v != null)
  if (values.length === 0) return null
  
  return {
    min: Math.min(...values),
    max: Math.max(...values),
    avg: values.reduce((a, b) => a + b, 0) / values.length
  }
}

export const hasTrend = (id_obj) => {
  return trendStoreHist.value.some(trend => trend.id_obj === id_obj)
}

export const getTrendColor = (index) => {
  return trendColors[index % trendColors.length]
}

export const getCurrentPeriod = computed(() => currentPeriod.value)
export const getActiveTrendsCount = computed(() => getActiveTrends.value.length)
export const getTotalPoints = computed(() => {
  return getActiveTrends.value.reduce((total, trend) => total + (trend.points ? trend.points.length : 0), 0)
})

export const getTypeText = (type) => {
  switch (type) {
    case 0, "0": return 'Первые N записей'
    case 1, "1": return 'Среднее: берем каждую N-ю запись'
    case 2, "2": return 'Мин-Макс: в каждом интервале берем мин и макс'
    case 3, "3": return 'Минимум: в каждом интервале берем только минимум'
    case 4, "4": return 'Максимум: в каждом интервале берем только максимум'
    default: return 'неизвестно'
  }
}

// Функция для экспорта данных в HTML - отдельный файл для каждого тренда
export const exportChartDataToHTML = (chartData,activeFiltersText) => {
  if (!chartData || chartData.length === 0) return []
  
  const files = []
  
  chartData.forEach((trend) => {
    if (!trend.data || trend.data.length === 0) return
    
    const points = [...trend.data].sort((a, b) => a.x - b.x)
    const halfLength = Math.ceil(points.length / 2)
    
    let html = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Экспорт данных тренда ID: ${trend.id_obj}</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .trend-table { 
            border: 1px solid #ddd; 
            border-collapse: collapse; 
            margin-bottom: 20px;
            font-size: 12px;
        }
        .trend-table th, .trend-table td { 
            border: 1px solid #ddd; 
            padding: 4px 8px; 
            text-align: left;
        }
        .trend-table th { 
            background-color: #f5f5f5; 
            font-weight: bold;
        }
        .trend-title { 
            background: #e9ecef; 
            padding: 8px; 
            font-weight: bold;
            margin-bottom: 5px;
        }
        @media print {
            body { margin: 0; }
        }
    </style>
</head>
<body>
    <h1>Экспорт данных тренда ID: ${trend.id_obj}</h1>
    <div class="trend-title">Всего точек: ${points.length}</div>
    <div>${activeFiltersText}</div>
    <table class="trend-table">
        <thead>
            <tr>
                <th>№</th>
                <th>Время</th>
                <th>Значение</th>
                <th>Качество</th>
                <th>№</th>
                <th>Время</th>
                <th>Значение</th>
                <th>Качество</th>
            </tr>
        </thead>
        <tbody>
    `
    
    for (let i = 0; i < halfLength; i++) {
      const leftPoint = points[i]
      const rightPoint = points[i + halfLength]
      
      const leftTime = leftPoint ? new Date(leftPoint.x).toLocaleString('ru-RU', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
        fractionalSecondDigits: 3
      }).replace(/(\d+)\.(\d+)\.(\d+),?/, '$3.$2.$1') : ''
      
      const rightTime = rightPoint ? new Date(rightPoint.x).toLocaleString('ru-RU', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
        fractionalSecondDigits: 3
      }).replace(/(\d+)\.(\d+)\.(\d+),?/, '$3.$2.$1') : ''
      
      html += `
            <tr>
                <td>${i + 1}</td>
                <td>${leftTime}</td>
                <td>${leftPoint ? leftPoint.y.toFixed(4) : ''}</td>
                <td>${leftPoint ? leftPoint.quality : ''}</td>
                <td>${i + halfLength + 1}</td>
                <td>${rightTime}</td>
                <td>${rightPoint ? rightPoint.y.toFixed(4) : ''}</td>
                <td>${rightPoint ? rightPoint.quality : ''}</td>
            </tr>
      `
    }
    
    html += `
        </tbody>
    </table>
</body>
</html>
    `
    
    files.push({
      filename: `trend_${trend.id_obj}_${new Date().toISOString().slice(0, 19).replace(/:/g, '-')}.html`,
      content: html
    })
  })
  
  return files
}


export { trendStoreHist }