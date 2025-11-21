// stores/websocketConnection.js
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { decode, encode } from '@msgpack/msgpack'

export const useWebSocketStore = defineStore('websocket', () => {
  // === WebSocket Connections ===
  const dataConnection = ref(null)
  const controlConnection = ref(null)
  const testConnection = ref(null)
  
  // === Connection Status with timestamps ===
  const connectionStatus = ref({
    data: { status: 'disconnected', connectedAt: null, disconnectedAt: null, errorAt: null },
    control: { status: 'disconnected', connectedAt: null, disconnectedAt: null, errorAt: null },
    test: { status: 'disconnected', connectedAt: null, disconnectedAt: null, errorAt: null }
  })

  // === Connection Statistics ===
  const connectionStats = ref({
    messagesReceived: 0,
    messagesSent: 0,
    startTime: new Date().toISOString(),
    lastActivity: new Date().toISOString()
  })

  // === Config ===
  const config = ref({
    dataPort: 8081,
    controlPort: 8082,
    testPort: 8083,
    clientId: 'client_' + Math.random().toString(36).substring(2, 9),
    userId: 'user',
    clientType: 'vue_client'
  })

  // === Getters ===
  const isConnected = computed(() => {
    return connectionStatus.value.data.status === 'connected' && 
           connectionStatus.value.control.status === 'connected'
  })

  // Overall status для NavbarTop.vue
  const overallStatus = computed(() => {
    const statuses = Object.values(connectionStatus.value).map(s => s.status)
    
    let status, color, text
    if (statuses.every(s => s === 'connected')) {
      status = 'connected'
      color = '#00ff00'
      text = '✅ Связь установлена'
    } else if (statuses.some(s => s === 'error')) {
      status = 'error'
      color = '#ffff00'
      text = '⚠️ Ошибка связи'
    } else if (statuses.some(s => s === 'connected')) {
      status = 'partial'
      color = '#ffa500'
      text = '🟡 Частичное соединение'
    } else {
      status = 'disconnected'
      color = '#ff0000'
      text = '❌ Нет связи'
    }
    
    return {
      status,
      color,
      text
    }
  })

  // Статистика для NavbarTop.vue
  const stats = computed(() => {
    return {
      messagesReceived: connectionStats.value.messagesReceived,
      messagesSent: connectionStats.value.messagesSent,
      startTime: connectionStats.value.startTime,
      lastActivity: connectionStats.value.lastActivity,
      connectionStatus: { ...connectionStatus.value },
      config: { ...config.value },
      overallStatus: { ...overallStatus.value }
    }
  })

  // === WebSocket Connection ===
  const connectWebSocket = (type) => {
    const portMap = {
      data: config.value.dataPort,
      control: config.value.controlPort,
      test: config.value.testPort
    }
    
    const url = `ws://localhost:${portMap[type]}/ws?clientId=${config.value.clientId}&userId=${config.value.userId}&type=${config.value.clientType}`
    
    // Update status before connecting
    updateConnectionStatus(type, 'connecting')
    
    const connection = new WebSocket(url)
    connection.binaryType = 'arraybuffer'
    
    connection.onopen = () => {
      updateConnectionStatus(type, 'connected')
      console.log(`✅ ${type} WebSocket connected`)
      
      if (type === 'data') dataConnection.value = connection
      else if (type === 'control') controlConnection.value = connection
      else if (type === 'test') testConnection.value = connection
    }
    
    connection.onmessage = (event) => {
      connectionStats.value.lastActivity = new Date().toISOString()
      connectionStats.value.messagesReceived++
      
      // Принудительное обновление реактивности
      connectionStats.value = { ...connectionStats.value }
      
      // Emit event for objects store to handle
      const eventDetail = { type, data: event.data }
      window.dispatchEvent(new CustomEvent('websocketMessage', { detail: eventDetail }))
    }
    
    connection.onclose = () => {
      updateConnectionStatus(type, 'disconnected')
      console.log(`❌ ${type} WebSocket disconnected`)
    }
    
    connection.onerror = (error) => {
      updateConnectionStatus(type, 'error')
      console.error(`❌ ${type} WebSocket error:`, error)
    }
  }

  // === Connection Status Management ===
  const updateConnectionStatus = (type, status) => {
    const now = new Date().toISOString()
    connectionStatus.value[type] = {
      ...connectionStatus.value[type],
      status,
      ...(status === 'connected' && { connectedAt: now }),
      ...(status === 'disconnected' && { disconnectedAt: now }),
      ...(status === 'error' && { errorAt: now })
    }
    
    // Принудительное обновление реактивности
    connectionStatus.value = { ...connectionStatus.value }
  }

  // === Message Encoding ===
  const encodeMessage = (type, data, source) => {
    const message = {
      id: generateMessageId(),
      type: type,
      data: data,
      init_dt: new Date().toISOString(),
      update_dt: new Date().toISOString(),
      source: source,
      clientId: config.value.clientId
    }
    return message
  }

  const generateMessageId = () => {
    return `${Date.now()}-${Math.random().toString(36).substring(2, 9)}`
  }

  // === Send Message ===
  const sendMessage = (type, messageData, useBinary = true) => {
    const connectionMap = {
      data: dataConnection.value,
      control: controlConnection.value,
      test: testConnection.value
    }
    
    const connection = connectionMap[type]
    if (!connection || connection.readyState !== WebSocket.OPEN) {
      console.error(`❌ ${type} WebSocket not connected`)
      return { success: false, error: `${type} channel not connected` }
    }
    
    try {
      let messageToSend
      if (useBinary) {
        const encodedMessage = encode(messageData)
        connection.send(encodedMessage)
      } else {
        connection.send(JSON.stringify(messageData))
      }
      
      connectionStats.value.messagesSent++
      connectionStats.value.lastActivity = new Date().toISOString()
      
      // Принудительное обновление реактивности
      connectionStats.value = { ...connectionStats.value }
      
      console.log(`✅ Message sent via ${type}:`, messageData)
      return { success: true }
    } catch (error) {
      console.error('❌ Error sending message:', error)
      return { success: false, error: error.message }
    }
  }

  // === Send Command (convenience method) ===
  const sendCommand = async (objectId, command, command_source, data = {}) => {
    const commandMessage = {
      command: command,
      objectId: objectId,
      clientId: config.value.clientId,
      userId: config.value.userId,
      data: data,
      time: new Date().toISOString()
    }
    
    const message = encodeMessage(command, commandMessage, command_source)
    return sendMessage('control', message)
  }

  // === Send Message from HMI ===
  const sendMessageFromHMI = async (MessName, MessState, Color) => {
    const now = new Date()
    
    const commandData = [{
      IdObj: 100000,
      TypeObj: 100000,
      Code: 100000,
      Dt: now.getTime(),
      DtTxt: now.toISOString().replace('T', ' ').replace('Z', ''),
      Tag: 'HMI',
      MessFull: `${MessName}: ${MessState}`,
      MessName: MessName,
      MessState: MessState,
      UsoID: 100000,
      UsoTxt: '-',
      Users: `${config.value.userId}: ${config.value.clientId}(${config.value.clientType})`,
      Severity: 100000,
      Opermess: 1,
      Color: Color
    }]
    
    return sendCommand('alarms_system', 'command', 'alarms_set_mess', commandData)
  }

  // === Connection Management ===
  const connectAll = () => {
    console.log('🔌 Connecting all WebSockets...')
    connectWebSocket('data')
    connectWebSocket('control')
    connectWebSocket('test')
  }

  const disconnectAll = () => {
    const connections = [
      { conn: dataConnection.value, type: 'data' },
      { conn: controlConnection.value, type: 'control' },
      { conn: testConnection.value, type: 'test' }
    ]
    
    connections.forEach(({ conn, type }) => {
      if (conn && conn.readyState === WebSocket.OPEN) {
        conn.close()
        updateConnectionStatus(type, 'disconnected')
      }
    })
    
    dataConnection.value = null
    controlConnection.value = null
    testConnection.value = null
    
    console.log('🔌 All WebSockets disconnected')
  }

  const updateConfig = (newConfig) => {
    config.value = { ...config.value, ...newConfig }
    // Принудительное обновление реактивности
    config.value = { ...config.value }
  }

  return {
    // State
    connectionStatus,
    connectionStats,
    config,
    
    // Getters
    isConnected,
    overallStatus,
    stats,
    
    // Actions
    connectWebSocket,
    sendMessage,
    sendCommand,
    sendMessageFromHMI,
    connectAll,
    disconnectAll,
    updateConfig
  }
})