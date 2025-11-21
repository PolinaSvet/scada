<template>
  <header class="navbar-top">
    <div class="navbar-content">
      <!-- Левая часть - логотип и бренд -->
      <div class="brand">
        <div class="logo">⚙️</div>
        <span class="brand-text">SCADA SYSTEM</span>
      </div>

      <!-- Центральная часть - навигация/статус -->
      <div class="navbar-center">
        <div class="screen-title">{{ currentScreenTitle }}</div>
      </div>

      <!-- Правая часть - состояние связи -->
      <div class="navbar-right">
        <div 
          class="data-indicator"
          :style="indicatorStyle"
          @click="showConnectionModal"
          :title="currentStatus.text"
        ></div>
      </div>

      <!-- Модальное окно с информацией о подключении -->
      <div v-if="showModal" class="modal-overlay" @click="hideModal">
        <div class="modal-content" @click.stop>
          <div class="modal-header">
            <h3>Информация о подключении</h3>
            <button class="close-btn" @click="hideModal">×</button>
          </div>
          <div class="modal-body">
            <div class="connection-info">
              <h4>Статус подключения: {{ currentStatus.text }}</h4>
              <div class="info-grid">
                <div class="info-item">
                  <label>Время запуска:</label>
                  <span>{{ formatTime(currentStats.startTime) }}</span>
                </div>
                <div class="info-item">
                  <label>Client ID:</label>
                  <span>{{ currentStats.config.clientId }}</span>
                </div>
                <div class="info-item">
                  <label>User ID:</label>
                  <span>{{ currentStats.config.userId }}</span>
                </div>
                <div class="info-item">
                  <label>Client Type:</label>
                  <span>{{ currentStats.config.clientType }}</span>
                </div>
                <div class="info-item">
                  <label>Получено сообщений:</label>
                  <span>{{ currentStats.messagesReceived }}</span>
                </div>
                <div class="info-item">
                  <label>Отправлено сообщений:</label>
                  <span>{{ currentStats.messagesSent }}</span>
                </div>
                <div class="info-item">
                  <label>Последняя активность:</label>
                  <span>{{ formatTime(currentStats.lastActivity) }}</span>
                </div>
              </div>
              
              <h4>Статус WebSocket соединений:</h4>
              <div class="connections-grid">
                <div 
                  v-for="(conn, type) in currentStats.connectionStatus" 
                  :key="type" 
                  class="connection-item"
                >
                  <div class="conn-type">{{ type.toUpperCase() }}</div>
                  <div class="conn-status" :class="conn.status">
                    {{ getStatusText(conn.status) }}
                  </div>
                  <div class="conn-times">
                    <div>Подключен: {{ formatTime(conn.connectedAt) }}</div>
                    <div>Отключен: {{ formatTime(conn.disconnectedAt) }}</div>
                    <div>Ошибка: {{ formatTime(conn.errorAt) }}</div>
                  </div>
                </div>
              </div>
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn btn-primary" @click="hideModal">Закрыть</button>
            <button class="btn btn-secondary" @click="forceUpdate">Обновить</button>
          </div>
        </div>
      </div>
    </div>
  </header>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { useLayoutStore } from '@/stores/layout'
import { useWebSocketStore  } from '@/stores/websocketConnection'

const layoutStore = useLayoutStore()
const objectsStore = useWebSocketStore()
const showModal = ref(false)

// Локальные реактивные переменные
const currentStatus = ref({ status: 'disconnected', color: '#ff0000', text: '❌ Нет связи' })
const currentStats = ref({
  messagesReceived: 0,
  messagesSent: 0,
  startTime: new Date().toISOString(),
  lastActivity: null,
  connectionStatus: {},
  config: {},
  overallStatus: {}
})

// Используем центральную конфигурацию для заголовка
const currentScreenTitle = computed(() => 
  layoutStore.currentScreenConfig?.title || 'SCADA System'
)

// Watcher для отслеживания изменений в store
watch(
  () => objectsStore.overallStatus,
  (newStatus) => {
    currentStatus.value = newStatus
  },
  { immediate: true, deep: true }
)

watch(
  () => objectsStore.stats,
  (newStats) => {
    currentStats.value = newStats
  },
  { immediate: true, deep: true }
)

// Стили индикатора
const indicatorStyle = computed(() => ({
  background: currentStatus.value.color,
  boxShadow: `0 0 10px ${currentStatus.value.color}50`
}))

// Принудительное обновление
const forceUpdate = () => {
  currentStatus.value = { ...objectsStore.overallStatus }
  currentStats.value = { ...objectsStore.stats }
  console.log('🔄 Forced update')
}

// Методы для модального окна
const showConnectionModal = () => {
  // Обновляем данные перед показом модального окна
  forceUpdate()
  showModal.value = true
}

const hideModal = () => {
  showModal.value = false
}

// Вспомогательные функции
const formatTime = (timestamp) => {
  if (!timestamp) return '-'
  return new Date(timestamp).toLocaleString('ru-RU')
}

const getStatusText = (status) => {
  const statusMap = {
    connected: 'Подключен',
    disconnected: 'Отключен',
    error: 'Ошибка',
    connecting: 'Подключается'
  }
  return statusMap[status] || status
}


</script>

<style scoped>
@import '@/assets/styles/navbar-top.css';
</style>