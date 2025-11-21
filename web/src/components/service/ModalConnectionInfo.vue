<template>
  <div v-if="modelValue" class="modal-overlay" @click="close">
    <div class="modal-content" @click.stop>
      <div class="modal-header">
        <h3>{{ title }}</h3>
        <button class="close-btn" @click="close" title="Закрыть">×</button>
      </div>
      
      <div class="modal-body">
        <!-- Информация о клиенте -->
        <section class="info-section">
          <h4>Информация о клиенте</h4>
          <div class="info-grid">
            <div class="info-item">
              <label>Время запуска:</label>
              <span>{{ clientInfo.startTime }}</span>
            </div>
            <div class="info-item">
              <label>Client ID:</label>
              <span>{{ clientInfo.clientId }}</span>
            </div>
            <div class="info-item">
              <label>User ID:</label>
              <span>{{ clientInfo.userId }}</span>
            </div>
            <div class="info-item">
              <label>Client Type:</label>
              <span>{{ clientInfo.clientType }}</span>
            </div>
          </div>
        </section>

        <!-- Статус WebSocket подключений -->
        <section class="info-section">
          <h4>Статус WebSocket подключений</h4>
          <div class="connection-grid">
            <div 
              v-for="(conn, type) in connectionStatus" 
              :key="type" 
              class="connection-item"
            >
              <div class="conn-type">{{ type.toUpperCase() }}</div>
              <div class="conn-status" :class="conn.status">
                {{ getStatusText(conn.status) }}
              </div>
              <div class="conn-times">
                <div><strong>Подключен:</strong> {{ formatTime(conn.connectedAt) }}</div>
                <div><strong>Отключен:</strong> {{ formatTime(conn.disconnectedAt) }}</div>
                <div><strong>Ошибка:</strong> {{ formatTime(conn.errorAt) }}</div>
              </div>
            </div>
          </div>
        </section>

        <!-- Статистика сообщений -->
        <section class="info-section">
          <h4>Статистика сообщений (последние 10 мин)</h4>
          <div class="stats-grid">
            <div v-for="(count, type) in messageStats" :key="type" class="stat-item">
              <label>{{ formatMessageType(type) }}:</label>
              <span>{{ count }}</span>
            </div>
          </div>
        </section>

        <!-- Дополнительная информация -->
        <section v-if="additionalInfo && Object.keys(additionalInfo).length" class="info-section">
          <h4>Дополнительная информация</h4>
          <div class="info-grid">
            <div 
              v-for="(value, key) in additionalInfo" 
              :key="key" 
              class="info-item"
            >
              <label>{{ formatKey(key) }}:</label>
              <span>{{ value }}</span>
            </div>
          </div>
        </section>
      </div>

      <div class="modal-footer">
        <button class="btn btn-secondary" @click="close">Закрыть</button>
        <button v-if="showRefresh" class="btn btn-primary" @click="refresh">
          Обновить
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useWebSocketStore } from '@/stores/websocketConnection'
import { useObjectsStore } from '@/stores/objects'

// Props
const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  title: {
    type: String,
    default: 'Информация о подключении'
  },
  showRefresh: {
    type: Boolean,
    default: true
  },
  additionalInfo: {
    type: Object,
    default: () => ({})
  }
})

// Emits
const emit = defineEmits(['update:modelValue', 'close', 'refresh'])

// Stores
const websocketStore = useWebSocketStore()
const objectsStore = useObjectsStore()

// Computed
const clientInfo = computed(() => {
  const stats = websocketStore.getStats()
  return {
    startTime: formatTime(stats.startTime),
    clientId: stats.config.clientId,
    userId: stats.config.userId,
    clientType: stats.config.clientType
  }
})

const connectionStatus = computed(() => websocketStore.connectionStatus)
const messageStats = computed(() => objectsStore.getMessageStats())

// Methods
const close = () => {
  emit('update:modelValue', false)
  emit('close')
}

const refresh = () => {
  emit('refresh')
}

const formatTime = (timestamp) => {
  if (!timestamp) return '-'
  return new Date(timestamp).toLocaleString('ru-RU')
}

const formatMessageType = (type) => {
  const typeMap = {
    data_batch: 'Data Batch',
    mess_batch: 'Message Batch', 
    alarms_set_data: 'Alarms Data',
    trends_set_data: 'Trends Data',
    unknown_messages: 'Unknown Messages',
    sendCommand: 'Send Command'
  }
  return typeMap[type] || type
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

const formatKey = (key) => {
  return key.replace(/([A-Z])/g, ' $1').replace(/^./, str => str.toUpperCase())
}
</script>

<style scoped>
@import '@/assets/styles/modal-connection-info.css';
</style>