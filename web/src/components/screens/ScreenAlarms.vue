<template>
  <div class="screen-alarms">
    <div class="screen-header">
      <h1>⚠️ Аварийные события</h1>
      <p>История и текущие аварийные ситуации</p>
    </div>
    
    <div class="alarms-grid">
      <div class="alarm-card critical">
        <div class="alarm-header">
          <span class="alarm-icon">🔥</span>
          <span class="alarm-title">КРИТИЧЕСКИЙ УРОВЕНЬ</span>
        </div>
        <div class="alarm-content">
          <p>Датчик TEMP_001: Температура превысила 150°C</p>
          <span class="alarm-time">10:23:45</span>
        </div>
      </div>
      
      <div class="alarm-card warning">
        <div class="alarm-header">
          <span class="alarm-icon">⚠️</span>
          <span class="alarm-title">ПРЕДУПРЕЖДЕНИЕ</span>
        </div>
        <div class="alarm-content">
          <p>Датчик PRESS_012: Давление接近临界值</p>
          <span class="alarm-time">09:45:12</span>
        </div>
      </div>
      
      <div class="alarm-card info">
        <div class="alarm-header">
          <span class="alarm-icon">ℹ️</span>
          <span class="alarm-title">ИНФОРМАЦИЯ</span>
        </div>
        <div class="alarm-content">
          <p>Система: Плановое обслуживание запланировано</p>
          <span class="alarm-time">08:30:00</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'ScreenAlarms',
  setup() {
    return {
      // Логика для экрана аварий
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

.alarms-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 20px;
}

.alarm-card {
  padding: 20px;
  border-radius: 8px;
  border-left: 4px solid;
}

.alarm-card.critical {
  background: #ff4444;
  border-left-color: #cc0000;
}

.alarm-card.warning {
  background: #ffaa00;
  border-left-color: #cc8800;
  color: #000;
}

.alarm-card.info {
  background: #4287f5;
  border-left-color: #1a5fc9;
}

.alarm-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 15px;
}

.alarm-icon {
  font-size: 20px;
}

.alarm-title {
  font-weight: bold;
  font-size: 14px;
}

.alarm-content p {
  margin: 0 0 10px 0;
  font-size: 14px;
}

.alarm-time {
  font-family: 'Courier New', monospace;
  font-size: 12px;
  opacity: 0.8;
}
</style>