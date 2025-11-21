// stores/objects.js
import { defineStore } from 'pinia'
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { decode } from '@msgpack/msgpack'
import { addToAlarmStore } from '@/stores/alarmStore.js'
import { addMessHistBatch, clearAlarmHistStore } from '@/stores/alarmStoreHist.js'
import { addTrendsHistBatch, clearTrendsHistStore } from '@/stores/trendStoreHist.js'

export const useObjectsStore = defineStore('objects', () => {
  // === Objects State ===
  const objects = ref({})
  const activeControl = ref(null)
  const activeSubscriptions = ref(new Set())
  const testMessages = ref([])
  
  // === Message Statistics ===
  const messageStats = ref({
    data_batch: 0,
    mess_batch: 0,
    alarms_set_data: 0,
    trends_set_data: 0,
    unknown_messages: 0,
    sendCommand: 0
  })

  // === Флаг инициализации ===
  const isWebSocketInitialized = ref(false)

  // === Getters ===
  const subscribedObjects = computed(() => {
    const result = {}
    activeSubscriptions.value.forEach(id => {
      if (objects.value[id]) {
        result[id] = objects.value[id]
      }
    })
    return result
  })

  const recentTestMessages = computed(() => {
    return testMessages.value.slice(-10)
  })

  // === Message Handling ===
  const handleWebSocketMessage = async (event) => {
    const { type, data } = event.detail
    
    try {
      let message
      if (typeof data === 'string') {
        message = JSON.parse(data)
      } else {
        const uint8Array = new Uint8Array(data)
        message = decode(uint8Array)
      }
      
      processDecodedMessage(type, message)
      
    } catch (error) {
      console.error('❌ Error processing message:', error)
      messageStats.value.unknown_messages++
    }
  }

  // === Main Message Processing Logic ===
  const processDecodedMessage = (type, message) => {
    switch (type) {
      case 'data':
        if (message.type === 'data_batch' || message.type === 'updateObjectsBatch') {
          const objectsArray = message.data
          if (Array.isArray(objectsArray)) {
            updateObjectsBatch(objectsArray)
            messageStats.value.data_batch++
          }
        }
        break
      
      case 'test':
        if (message.type === 'mess_batch') {
          const objectsArray = message.data
          if (Array.isArray(objectsArray)) {
            updateMessBatch(objectsArray)
            messageStats.value.mess_batch++
          }
        }
        else if (message.type === 'alarms_set_data') {
          const objectsArray = message.data
          if (Array.isArray(objectsArray)) {
            updateMessHistBatch(objectsArray)
            messageStats.value.alarms_set_data++
          } else {
            clearAlarmHistStore()
          }
        }
        else if (message.type === 'trends_set_data') {
          const objectsArray = message.data
          if (Array.isArray(objectsArray)) {
            updateTrendsHistBatch(objectsArray)
            messageStats.value.trends_set_data++
          } else {
            clearTrendsHistStore()
          }
        }
        break
        
      case 'control':
        console.log('🎛️ Control response:', message)
        break
        
      default:
        console.log('❓ Unknown message type:', type, message)
        messageStats.value.unknown_messages++
    }
  }

  // === Object Management Methods ===
  const updateObjectsBatch = (objectsArray) => {
    const updatedObjects = { ...objects.value }
    let updatedCount = 0
    
    objectsArray.forEach(obj => {
      if (obj?.id && activeSubscriptions.value.has(obj.id)) {
        updatedObjects[obj.id] = {
          ...updatedObjects[obj.id],
          ...obj,
          lastUpdate: new Date().toISOString()
        }
        updatedCount++
      }
    })
    
    if (updatedCount > 0) {
      objects.value = updatedObjects
    }
  }

  const updateMessBatch = (objectsArray) => {
    objectsArray.forEach(obj => {
      addToAlarmStore(obj)
    })
  }

  const updateMessHistBatch = (objectsArray) => {
    addMessHistBatch(objectsArray)
  }

  const updateTrendsHistBatch = (objectsArray) => {
    addTrendsHistBatch(objectsArray)
  }

  // === Subscription Management ===
  const subscribe = (objectId) => {
    if (!activeSubscriptions.value.has(objectId)) {
      activeSubscriptions.value.add(objectId)
      console.log(`✅ Subscribed to: ${objectId}, total: ${activeSubscriptions.value.size}`)
    }
  }

  const unsubscribe = (objectId) => {
    if (activeSubscriptions.value.has(objectId)) {
      activeSubscriptions.value.delete(objectId)
      console.log(`❌ Unsubscribed from: ${objectId}, total: ${activeSubscriptions.value.size}`)
    }
  }

  const subscribeMultiple = (objectIds) => {
    objectIds.forEach(id => {
      if (!activeSubscriptions.value.has(id)) {
        activeSubscriptions.value.add(id)
      }
    })
    console.log(`✅ Subscribed to ${objectIds.length} objects, total: ${activeSubscriptions.value.size}`)
  }

  const unsubscribeMultiple = (objectIds) => {
    objectIds.forEach(id => activeSubscriptions.value.delete(id))
    console.log(`❌ Unsubscribed from ${objectIds.length} objects, total: ${activeSubscriptions.value.size}`)
  }

  const clearAllSubscriptions = () => {
    const count = activeSubscriptions.value.size
    activeSubscriptions.value.clear()
    console.log(`🗑️ Cleared all subscriptions: ${count}`)
  }

  // === Control Management ===
  const openControl = (objectId) => {
    const obj = objects.value[objectId]
    if (obj?.objInfo?.ctrlEnable) {
      activeControl.value = objectId
      console.log(`🎮 Control opened for: ${objectId}`)
    } else {
      console.log(`⛔ Control not available for: ${objectId}`)
    }
  }

  const closeControl = () => {
    activeControl.value = null
    console.log('🎮 Control closed')
  }

  // === ИНИЦИАЛИЗАЦИЯ WebSocket (остается здесь) ===
  const initializeWebSocket = async () => {
    if (isWebSocketInitialized.value) {
      console.log('🔌 WebSocket already initialized')
      return true
    }

    try {
      console.log('🔌 Initializing WebSocket...')
      
      // Импортируем websocket store динамически чтобы избежать циклических зависимостей
      const { useWebSocketStore } = await import('@/stores/websocketConnection')
      const websocketStore = useWebSocketStore()
      
      websocketStore.connectAll()
      isWebSocketInitialized.value = true
      
      // Настраиваем обработчики сообщений
      setupEventListeners()
      
      console.log('✅ WebSocket initialized successfully')
      return true
    } catch (error) {
      console.error('❌ Failed to initialize WebSocket:', error)
      return false
    }
  }

  // Отключение WebSocket
  const disconnectWebSocket = () => {
    if (isWebSocketInitialized.value) {
      // Импортируем websocket store динамически
      import('@/stores/websocketConnection').then(({ useWebSocketStore }) => {
        const websocketStore = useWebSocketStore()
        websocketStore.disconnectAll()
      })
      
      isWebSocketInitialized.value = false
      activeSubscriptions.value.clear()
      removeEventListeners()
      console.log('🔌 WebSocket disconnected')
    }
  }

  // === Event Listener Setup ===
  const setupEventListeners = () => {
    window.addEventListener('websocketMessage', handleWebSocketMessage)
  }

  const removeEventListeners = () => {
    window.removeEventListener('websocketMessage', handleWebSocketMessage)
  }

  // === Send Command (использует websocket store) ===
  const sendCommand = async (objectId, command, command_source, data = {}) => {
    try {
      const { useWebSocketStore } = await import('@/stores/websocketConnection')
      const websocketStore = useWebSocketStore()
      
      return await websocketStore.sendCommand(objectId, command, command_source, data)
    } catch (error) {
      console.error('❌ Error sending command:', error)
      return { success: false, error: error.message }
    }
  }

  // === Send Message from HMI ===
  const sendMessageFromHMI = async (MessName, MessState, Color) => {
    try {
      const { useWebSocketStore } = await import('@/stores/websocketConnection')
      const websocketStore = useWebSocketStore()
      
      return await websocketStore.sendMessageFromHMI(MessName, MessState, Color)
    } catch (error) {
      console.error('❌ Error sending HMI message:', error)
      return { success: false, error: error.message }
    }
  }

  // Очистка
  const cleanup = () => {
    disconnectWebSocket()
  }

  // Автоматическая настройка обработчиков при создании store
  setupEventListeners()

  return {
    // State
    objects,
    activeControl,
    activeSubscriptions,
    testMessages,
    messageStats,
    isWebSocketInitialized,
    
    // Getters
    subscribedObjects,
    recentTestMessages,
    
    // Actions
    initializeWebSocket,
    disconnectWebSocket,
    subscribe,
    unsubscribe,
    subscribeMultiple,
    unsubscribeMultiple,
    clearAllSubscriptions,
    openControl,
    closeControl,
    sendCommand,
    sendMessageFromHMI,
    cleanup
  }
})