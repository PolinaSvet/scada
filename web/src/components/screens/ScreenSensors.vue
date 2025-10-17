<template>
  <div class="screen-sensors">
    <div class="screen-header">
      <h1>📊 Мониторинг датчиков</h1>
      <p>Реальное время состояние всех объектов</p>
    </div>
    
    <div class="controls">
      <button class="btn primary">Добавить датчик</button>
      <button class="btn secondary">Фильтр</button>
      <button class="btn secondary">Экспорт</button>
    </div>
    
    <div class="sensors-content">
      <p>Здесь будет отображаться список всех датчиков системы с их текущими значениями и статусами.</p>
      <p>Компонент виртуализированного списка будет интегрирован на этот экран.</p>
    </div>
  </div>
</template>

<script>
export default {
  name: 'ScreenSensors',
  setup() {
    return {
      // Логика для экрана датчиков
    }
  }
}
</script>

<style scoped>
.screen-sensors {
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

.controls {
  display: flex;
  gap: 10px;
  margin-bottom: 20px;
}

.btn {
  padding: 10px 20px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.2s ease;
}

.btn.primary {
  background: #42b883;
  color: white;
}

.btn.secondary {
  background: #34495e;
  color: white;
  border: 1px solid #42b883;
}

.btn:hover {
  opacity: 0.9;
  transform: translateY(-1px);
}

.sensors-content {
  background: #2c3e50;
  padding: 30px;
  border-radius: 8px;
  text-align: center;
}

.sensors-content p {
  margin: 0 0 15px 0;
  color: #ccc;
  font-size: 16px;
}

.sensors-content p:last-child {
  margin-bottom: 0;
}
</style>