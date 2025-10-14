<template>
  <div class="client-test">
    <h1>VueWay Test Client</h1>
    
    <!-- Статус подключения -->
    <div class="connection-status">
      <h2>Connection Status</h2>
      <div class="status-grid">
        <div 
          v-for="(status, type) in connectionStats" 
          :key="type"
          :class="['status-item', status]"
        >
          <span class="status-type">{{ type }}:</span>
          <span class="status-value">{{ status }}</span>
        </div>
      </div>
      
      <div class="stats">
        <p>Messages Received: {{ connectionStats.messagesReceived }}</p>
        <p>Messages Sent: {{ connectionStats.messagesSent }}</p>
      </div>
      
      <div class="controls">
        <button @click="connectAll" :disabled="isConnected">Connect All</button>
        <button @click="disconnectAll" :disabled="!isConnected">Disconnect All</button>
        <button @click="updateClientId">New Client ID</button>
      </div>
    </div>
    
    <!-- Данные объектов 
    <div class="object-data">
      <h2>Object Data (from Data WebSocket)</h2>
      <div v-if="objectData.length === 0" class="no-data">
        No object data received
      </div>
      <div v-else class="objects-grid">
        <div 
          v-for="obj in objectData" 
          :key="obj.id"
          class="object-item"
        >
          <h3>Object {{ obj.id }}</h3>
          <p><strong>Name:</strong> {{ obj.info?.name || 'N/A' }}</p>
          <p><strong>Type:</strong> {{ obj.info?.type || 'N/A' }}</p>
          <p><strong>State:</strong> {{ obj.state?.txtOn || 'N/A' }}</p>
          <div class="object-actions">
            <button @click="sendCommand(obj.id, 'ON')">ON</button>
            <button @click="sendCommand(obj.id, 'OFF')">OFF</button>
            <button @click="sendCommand(obj.id, 'RESET')">RESET</button>
          </div>
        </div>
      </div>
    </div>-->

    <!-- Данные объектов -->
    <div class="test-data">
      <h2>Object Data (from Data WebSocket)</h2>
      <div v-if="objectData.length === 0" class="no-data">
        No test data received
      </div>
      <div v-else class="test-list">
        <div 
          v-for="(test, index) in objectData" 
          :key="index"
          class="test-item"
        >
          <p><strong>Time:</strong> {{ test.current_time }}</p>
          <p><strong>Message:</strong> {{ test.message }}</p>
          <p><strong>Received:</strong> {{ test.received }}</p>
        </div>
      </div>
    </div>
    
    <!-- Тестовые данные -->
    <div class="test-data">
      <h2>Test Data (from Test WebSocket)</h2>
      <div v-if="testData.length === 0" class="no-data">
        No test data received
      </div>
      <div v-else class="test-list">
        <div 
          v-for="(test, index) in testData" 
          :key="index"
          class="test-item"
        >
          <p><strong>Time:</strong> {{ test.current_time }}</p>
          <p><strong>Message:</strong> {{ test.message }}</p>
          <p><strong>Received:</strong> {{ test.received }}</p>
        </div>
      </div>
    </div>
    
    <!-- Отправка команды -->
    <div class="command-panel">
      <h2>Send Command</h2>
      <div class="command-form">
        <input 
          v-model="commandObjectId" 
          placeholder="Object ID"
          type="text"
        >
        <input 
          v-model="commandName" 
          placeholder="Command"
          type="text"
        >
        <button 
          @click="sendCustomCommand"
          :disabled="connectionStats.control !== 'connected'"
        >
          Send Command
        </button>
      </div>
    </div>

    <!-- Информация о клиенте -->
    <div class="client-info">
      <h2>Client Information</h2>
      <p><strong>Client ID:</strong> {{ config.clientId }}</p>
      <p><strong>User ID:</strong> {{ config.userId }}</p>
      <p><strong>Type:</strong> {{ config.clientType }}</p>
    </div>
  </div>
</template>

<script setup>
import { useWebSocketStore } from '@/stores/websocket'
import { storeToRefs } from 'pinia'
import { ref } from 'vue'

const websocketStore = useWebSocketStore()

const {
  objectData,
  testData,
  connectionStats,
  isConnected,
  config
} = storeToRefs(websocketStore)

const commandObjectId = ref('')
const commandName = ref('')

const connectAll = () => {
  websocketStore.connectAll()
}

const disconnectAll = () => {
  websocketStore.disconnectAll()
}

const sendCommand = (objectId, command) => {
  websocketStore.sendCommand(objectId, command)
}

const sendCustomCommand = () => {
  if (commandObjectId.value && commandName.value) {
    websocketStore.sendCommand(commandObjectId.value, commandName.value)
    commandObjectId.value = ''
    commandName.value = ''
  }
}

const updateClientId = () => {
  const newClientId = 'test_client_' + Math.random().toString(36).substr(2, 9)
  websocketStore.updateConfig({ clientId: newClientId })
  disconnectAll()
}
</script>

<style scoped>
.client-test {
  padding: 20px;
  max-width: 1200px;
  margin: 0 auto;
}

.connection-status {
  background: #f5f5f5;
  padding: 20px;
  border-radius: 8px;
  margin-bottom: 20px;
}

.status-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 10px;
  margin-bottom: 15px;
}

.status-item {
  padding: 10px;
  border-radius: 4px;
  text-align: center;
}

.status-item.connected {
  background: #d4edda;
  color: #155724;
}

.status-item.disconnected {
  background: #f8d7da;
  color: #721c24;
}

.status-item.error {
  background: #fff3cd;
  color: #856404;
}

.controls {
  margin-top: 15px;
}

.controls button {
  margin-right: 10px;
  padding: 8px 16px;
  border: 1px solid #ccc;
  border-radius: 4px;
  background: white;
  cursor: pointer;
}

.controls button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.controls button:hover:not(:disabled) {
  background: #e9e9e9;
}

.objects-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 15px;
  margin-top: 15px;
}

.object-item {
  border: 1px solid #ddd;
  padding: 15px;
  border-radius: 8px;
  background: white;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

.object-actions {
  margin-top: 10px;
}

.object-actions button {
  margin-right: 5px;
  padding: 5px 10px;
  border: 1px solid #007bff;
  background: #007bff;
  color: white;
  border-radius: 3px;
  cursor: pointer;
}

.object-actions button:hover {
  background: #0056b3;
}

.test-list {
  max-height: 300px;
  overflow-y: auto;
  margin-top: 15px;
}

.test-item {
  border: 1px solid #eee;
  padding: 10px;
  margin-bottom: 10px;
  border-radius: 4px;
  background: white;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
}

.command-form {
  display: flex;
  gap: 10px;
  align-items: center;
  margin-top: 15px;
}

.command-form input {
  padding: 8px;
  border: 1px solid #ddd;
  border-radius: 4px;
  flex: 1;
}

.command-form button {
  padding: 8px 16px;
  border: 1px solid #28a745;
  background: #28a745;
  color: white;
  border-radius: 4px;
  cursor: pointer;
}

.command-form button:disabled {
  background: #6c757d;
  border-color: #6c757d;
  cursor: not-allowed;
}

.command-form button:hover:not(:disabled) {
  background: #218838;
}

.no-data {
  text-align: center;
  color: #666;
  padding: 20px;
  font-style: italic;
}

.client-info {
  background: #e9ecef;
  padding: 15px;
  border-radius: 8px;
  margin-top: 20px;
}

.client-info h2 {
  margin-bottom: 10px;
}
</style>