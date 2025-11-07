<template>
  <div class="screen-alarms">
    <div class="screen-header">
      <h1>🚨 СОБЫТИЯ СИСТЕМЫ</h1>
      <p>Группировка сообщений по категориям и типам</p>
      <div class="header-stats">
        <span class="total-messages">Всего сообщений: {{ totalMessages.total }}</span>
        <span class="alarm-count" v-if="totalMessages.alarms > 0">Аварии: {{ totalMessages.alarms }}</span>
      </div>
    </div>
    
    <!-- Отображение по категориям -->
    <div class="categories-section">
      <h2>📊 Статистика по категориям</h2>
      <div class="categories-grid">
        <div 
          v-for="category in alarmCategories" 
          :key="category.category" 
          class="category-card"
          :style="{ 
            backgroundColor: category.color,
            borderLeftColor: darkenColor(category.color, 0.2)
          }"
        >
          <div class="category-header">
            <span class="category-icon">{{ category.icon }}</span>
            <span class="category-title">{{ category.name }}</span>
            <span class="category-count">{{ category.count }}</span>
          </div>
          
          <!-- Детали по типам внутри категории -->
          <div class="category-types" v-if="Object.keys(category.types).length > 0">
            <div 
              v-for="type in Object.values(category.types)" 
              :key="type.messType" 
              class="type-item"
            >
              <span class="type-name">{{ type.name }} ({{ type.messType }})</span>
              <span class="type-count">{{ type.count }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
    
    <!-- Детальное отображение по типам -->
    <div class="types-section">
      <h2>🔍 Детали по типам сообщений</h2>
      <div class="alarms-grid">
        <div 
          v-for="group in alarmGroups" 
          :key="group.messType" 
          class="alarm-card"
          :style="{ 
            backgroundColor: group.messColor,
            borderLeftColor: darkenColor(group.messColor, 0.1)
          }"
        >
          <div class="alarm-header">
            <span class="alarm-icon">{{ group.icon }}</span>
            <div class="alarm-title-section">
              <span class="alarm-title">{{ group.name }}</span>
              <span class="alarm-category">{{ group.categoryName }}</span>
            </div>
          </div>
          <div class="alarm-content">
            <p><strong>Код:</strong> {{ group.messType }}</p>
            <p><strong>Количество:</strong> {{ group.count }}</p>
            <p><strong>Категория:</strong> {{ group.categoryName }}</p>
          </div>
        </div>
        
        <!-- Сообщение если нет данных -->
        <div v-if="alarmGroups.length === 0" class="no-data-message">
          <p>Нет сообщений для отображения</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { getAlarmStatsInfo, getAlarmStats, getAlarmStatsByCategory } from '@/stores/alarmStore.js'

export default {
  name: 'ScreenAlarms',
  setup() {
    // Функция для затемнения цвета (для border)
    const darkenColor = (color, factor) => {
      // Простая реализация затемнения цвета
      if (color.startsWith('#')) {
        let r = parseInt(color.slice(1, 3), 16)
        let g = parseInt(color.slice(3, 5), 16)
        let b = parseInt(color.slice(5, 7), 16)
        
        r = Math.floor(r * (1 - factor))
        g = Math.floor(g * (1 - factor))
        b = Math.floor(b * (1 - factor))
        
        return `#${r.toString(16).padStart(2, '0')}${g.toString(16).padStart(2, '0')}${b.toString(16).padStart(2, '0')}`
      }
      return color
    }
    
    return {
      alarmGroups: getAlarmStatsInfo,
      alarmCategories: getAlarmStatsByCategory,
      totalMessages: getAlarmStats,
      darkenColor
    }
  }
}
</script>

<style scoped>
.screen-alarms {
  padding: 20px;
}

.screen-header {
  margin-bottom: 30px;
}

.screen-header h1 {
  margin: 0 0 10px 0;
  color: #42b883;
}

.screen-header p {
  margin: 0;
  color: #ccc;
  font-size: 16px;
}

.header-stats {
  margin-top: 10px;
  display: flex;
  gap: 15px;
  font-size: 14px;
}

.total-messages {
  color: #42b883;
  font-weight: bold;
}

.alarm-count {
  color: #ff4444;
  font-weight: bold;
}

.error-count {
  color: #ff00ff;
  font-weight: bold;
}

.categories-section,
.types-section {
  margin-bottom: 40px;
}

.categories-section h2,
.types-section h2 {
  color: #42b883;
  margin-bottom: 20px;
  font-size: 20px;
}

.categories-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(350px, 1fr));
  gap: 20px;
  margin-bottom: 30px;
}

.category-card {
  padding: 20px;
  border-radius: 8px;
  border-left: 4px solid;
  color: white;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
}

.category-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 15px;
}

.category-icon {
  font-size: 24px;
}

.category-title {
  font-weight: bold;
  font-size: 16px;
  text-transform: uppercase;
  flex-grow: 1;
}

.category-count {
  font-size: 24px;
  font-weight: bold;
  background: rgba(255, 255, 255, 0.2);
  padding: 4px 12px;
  border-radius: 20px;
}

.category-types {
  border-top: 1px solid rgba(255, 255, 255, 0.3);
  padding-top: 10px;
}

.type-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 6px 0;
  font-size: 13px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.type-item:last-child {
  border-bottom: none;
}

.type-name {
  flex-grow: 1;
}

.type-count {
  background: rgba(255, 255, 255, 0.2);
  padding: 2px 8px;
  border-radius: 12px;
  font-weight: bold;
}

.alarms-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 20px;
}

.alarm-card {
  padding: 20px;
  border-radius: 8px;
  border-left: 4px solid;
  color: white;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
}

.alarm-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 15px;
}

.alarm-icon {
  font-size: 24px;
}

.alarm-title-section {
  flex-grow: 1;
}

.alarm-title {
  display: block;
  font-weight: bold;
  font-size: 14px;
  text-transform: uppercase;
}

.alarm-category {
  display: block;
  font-size: 12px;
  opacity: 0.9;
  margin-top: 2px;
}

.alarm-content p {
  margin: 0 0 8px 0;
  font-size: 14px;
}

.no-data-message {
  grid-column: 1 / -1;
  text-align: center;
  padding: 40px;
  color: #ccc;
  font-size: 18px;
}
</style>