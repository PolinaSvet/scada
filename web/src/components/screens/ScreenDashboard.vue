<template>
  <div class="screen-dashboard">
    <div class="screen-header">
      <h1>🏠 Главная панель</h1>
      <p>Обзор системы мониторинга</p>
    </div>
    
    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-icon">📊</div>
        <div class="stat-info">
          <h3>Всего объектов</h3>
          <span class="stat-value">156</span>
        </div>
      </div>
      
      <div class="stat-card">
        <div class="stat-icon">✅</div>
        <div class="stat-info">
          <h3>В норме</h3>
          <span class="stat-value" style="color: #00ff00">142</span>
        </div>
      </div>
      
      <div class="stat-card">
        <div class="stat-icon">⚠️</div>
        <div class="stat-info">
          <h3>Аварии</h3>
          <span class="stat-value" style="color: #ff4444">3</span>
        </div>
      </div>
      
      <div class="stat-card">
        <div class="stat-icon">🔧</div>
        <div class="stat-info">
          <h3>Обслуживание</h3>
          <span class="stat-value" style="color: #ffaa00">11</span>
        </div>
      </div>
    </div>
    
    <div class="recent-activity">
      <h2>Последние события</h2>
      <div class="activity-list">
        <div class="activity-item">
          <span class="time">10:23:45</span>
          <span class="message">Датчик TEMP_001: Превышение температуры</span>
          <span class="status error">АВАРИЯ</span>
        </div>
        <div class="activity-item">
          <span class="time">10:20:12</span>
          <span class="message">Датчик PRESS_045: Давление в норме</span>
          <span class="status normal">НОРМА</span>
        </div>
        <div class="activity-item">
          <span class="time">10:15:33</span>
          <span class="message">Система: Автоматическое обновление данных</span>
          <span class="status info">ИНФО</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'ScreenDashboard',
  setup() {
    return {
      // Логика для главной панели
    }
  }
}
</script>

<style scoped>
.screen-dashboard {
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

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 20px;
  margin-bottom: 30px;
}

.stat-card {
  background: #2c3e50;
  padding: 20px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  gap: 15px;
  border-left: 4px solid #42b883;
}

.stat-icon {
  font-size: 32px;
}

.stat-info h3 {
  margin: 0 0 8px 0;
  font-size: 14px;
  color: #ccc;
}

.stat-value {
  font-size: 24px;
  font-weight: bold;
}

.recent-activity h2 {
  margin-bottom: 15px;
  color: #42b883;
}

.activity-list {
  background: #2c3e50;
  border-radius: 8px;
  overflow: hidden;
}

.activity-item {
  display: flex;
  align-items: center;
  padding: 12px 15px;
  border-bottom: 1px solid #34495e;
  gap: 15px;
}

.activity-item:last-child {
  border-bottom: none;
}

.time {
  font-family: 'Courier New', monospace;
  color: #42b883;
  min-width: 80px;
}

.message {
  flex: 1;
}

.status {
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: bold;
}

.status.error {
  background: #ff4444;
  color: white;
}

.status.normal {
  background: #00ff00;
  color: #000;
}

.status.info {
  background: #4287f5;
  color: white;
}
</style>