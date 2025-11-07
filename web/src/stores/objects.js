// stores/objects.js
import { defineStore } from 'pinia'
import { decode, encode } from '@msgpack/msgpack'
import { ref, computed } from 'vue'
import { addToAlarmStore } from '@/stores/alarmStore.js'

export const useObjectsStore = defineStore('objects', () => {
  // === WebSocket Connections ===
  const dataConnection = ref(null)
  const controlConnection = ref(null)
  const testConnection = ref(null)
  
  // === Connection Status ===
  const connectionStatus = ref({
    data: 'disconnected',
    control: 'disconnected', 
    test: 'disconnected'
  })
  
  // === Objects State (ОСНОВНАЯ ЛОГИКА) ===
  const objects = ref({}) // Все объекты: { [objectId]: objectData }
  const activeControl = ref(null) // Активный объект для управления
  const activeSubscriptions = ref(new Set()) // Подписки на обновления
  const messagesReceived = ref(0)
  const messagesSent = ref(0)
  
  // === Флаг инициализации (для совместимости) ===
  const isWebSocketInitialized = ref(false)
  
  // === Test Data (просто для консоли) ===
  const testMessages = ref([])
  
  // === Config ===
  const config = ref({
    dataPort: 8081,
    controlPort: 8082,
    testPort: 8083,
    clientId: 'client_' + Math.random().toString(36).substr(2, 9),
    userId: 'user',
    clientType: 'vue_client'
  })

  // === Getters ===
  const isConnected = computed(() => {
    return connectionStatus.value.data === 'connected' && 
           connectionStatus.value.control === 'connected'
  })
  
  // ОБЪЕКТЫ ПО ПОДПИСКАМ (основной геттер для компонентов)
  const subscribedObjects = computed(() => {
    const result = {}
    activeSubscriptions.value.forEach(id => {
      if (objects.value[id]) {
        result[id] = objects.value[id]
      }
    })
    return result
  })
  
  // Последние тестовые сообщения (только для отладки)
  const recentTestMessages = computed(() => {
    return testMessages.value.slice(0, 10)
  })

  // === WebSocket Connection ===
  const connectWebSocket = (type) => {
    const portMap = {
      data: config.value.dataPort,
      control: config.value.controlPort,
      test: config.value.testPort
    }
    
    const url = `ws://localhost:${portMap[type]}/ws?clientId=${config.value.clientId}&userId=${config.value.userId}&type=${config.value.clientType}`
    
    const connection = new WebSocket(url)
    connection.binaryType = 'arraybuffer'
    
    connection.onopen = () => {
      connectionStatus.value[type] = 'connected'
      console.log(`✅ ${type} WebSocket connected`)
      
      if (type === 'data') dataConnection.value = connection
      else if (type === 'control') controlConnection.value = connection
      else if (type === 'test') testConnection.value = connection
    }
    
    connection.onmessage = (event) => {
      handleMessage(type, event.data)
    }
    
    connection.onclose = () => {
      connectionStatus.value[type] = 'disconnected'
      console.log(`❌ ${type} WebSocket disconnected`)
    }
    
    connection.onerror = (error) => {
      connectionStatus.value[type] = 'error'
      console.error(`❌ ${type} WebSocket error:`, error)
    }
  }
  
  // === Message Handling ===
  const handleMessage = async (type, data) => {
    try {
      messagesReceived.value++
      
      let message
      if (typeof data === 'string') {
        message = JSON.parse(data)
      } else {
        const uint8Array = new Uint8Array(data)
        message = decode(uint8Array)
      }
      
      processDecodedMessage(type, message)
      //console.log('XXXXXXXXX',type, message)
      
    } catch (error) {
      console.error('❌ Error processing message:', error)
    }
  }
  
  // === ОСНОВНАЯ ЛОГИКА ОБРАБОТКИ СООБЩЕНИЙ ===
  const processDecodedMessage = (type, message) => {
    switch (type) {
      case 'data':
        // ДАННЫЕ ОБЪЕКТОВ - основная логика
        if (message.type === 'data_batch' || message.type === 'updateObjectsBatch') {
          const objectsArray = message.data
         
          if (Array.isArray(objectsArray)) {
            updateObjectsBatch(objectsArray)
            //console.log(`📦 Received ${objectsArray.length} objects, subscriptions: ${activeSubscriptions.value.size}`)
          }
        }
        break
      
      case 'test':
          // ДАННЫЕ ОБЪЕКТОВ - основная логика
          if (message.type === 'mess_batch' ) {
            const objectsArray = message.data
  
            //console.log('mess_batch',objectsArray)
            
            if (Array.isArray(objectsArray)) {
              updateMessBatch(objectsArray)
              //console.log(`📦 mess_batch ${objectsArray.length}`)
            }
          }
          break  
        
      case 'control':
        // ОТВЕТЫ НА КОМАНДЫ - логируем
        console.log('🎛️ Control response:', message)
        break
        
      default:
        console.log('❓ Unknown message type:', type, message)
    }
  }

  // === ИНИЦИАЛИЗАЦИЯ WebSocket (для совместимости) ===
  const initializeWebSocket = async () => {
    // Если WebSocket уже инициализирован, не делаем повторную инициализацию
    if (isWebSocketInitialized.value) {
      console.log('🔌 WebSocket already initialized')
      return true
    }

    try {
      console.log('🔌 Initializing WebSocket...')
      connectAll()
      isWebSocketInitialized.value = true
      console.log('✅ WebSocket initialized successfully')
      return true
    } catch (error) {
      console.error('❌ Failed to initialize WebSocket:', error)
      return false
    }
  }

  // Отключение WebSocket (вызывать при закрытии приложения)
  const disconnectWebSocket = () => {
    if (isWebSocketInitialized.value) {
      disconnectAll()
      isWebSocketInitialized.value = false
      activeSubscriptions.value.clear()
      console.log('🔌 WebSocket disconnected')
    }
  }

  // === ОСНОВНЫЕ МЕТОДЫ ДЛЯ РАБОТЫ С СООБЩЕНИЯМИ ===

  // Обновление объектов (фильтруем по подпискам)
  const updateMessBatch = (objectsArray) => {
      objectsArray.forEach(obj => {
        addToAlarmStore(obj)
        /*obj.forEach(objItem => {
          addToAlarmStore(objItem)
        })*/

      })
  }

  // === ОСНОВНЫЕ МЕТОДЫ ДЛЯ РАБОТЫ С ОБЪЕКТАМИ ===
  
  // Обновление объектов (фильтруем по подпискам)
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

  // ПОДПИСКИ (основная функциональность)
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

  // УПРАВЛЕНИЕ ОБЪЕКТАМИ
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

  // Функция для кодирования сообщения
  const encodeMessage = (type, data, source = 'vue-client') => {

    //const serializedData = JSON.stringify(data)
    const serializedData = data

    const message = {
      id: generateMessageId(), // Нужно реализовать генерацию ID
      type: type,
      data: serializedData,
      init_dt: new Date().toISOString(),
      update_dt: new Date().toISOString(),
      source: source,
      clientId: config.value.clientId
    }
    return message
  }

  // Генерация ID сообщения
  const generateMessageId = () => {
    return `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`
  }

  // ОТПРАВКА КОМАНД (через control WebSocket)
  const sendCommand = async (objectId, command, data = {}) => {
    if (connectionStatus.value.control !== 'connected') {
      console.error('❌ Control WebSocket not connected')
      return { success: false, error: 'Control channel not connected' }
    }
    
    const commandMessage = {
      command: command,
      objectId: objectId,
      clientId: config.value.clientId,
      userId: config.value.userId,
      data: data,
      time: new Date().toISOString()
    }
    
    try {

      // Упаковываем в структуру Message
      const message = encodeMessage('command', commandMessage)

      const encodedMessage = encode(message)
      controlConnection.value.send(encodedMessage)
      messagesSent.value++
      console.log(`📤 Command sent: ${command} to ${objectId}`, message)
      return { success: true }
    } catch (error) {
      console.error('❌ Error sending command:', error)
      return { success: false, error: error.message }
    }
  }

  // УПРАВЛЕНИЕ СОЕДИНЕНИЕМ
  const connectAll = () => {
    console.log('🔌 Connecting all WebSockets...')
    connectWebSocket('data')
    connectWebSocket('control')
    connectWebSocket('test')
  }

  const disconnectAll = () => {
    const connections = [
      dataConnection.value,
      controlConnection.value,
      testConnection.value
    ]
    
    connections.forEach(conn => {
      if (conn && conn.readyState === WebSocket.OPEN) {
        conn.close()
      }
    })
    
    dataConnection.value = null
    controlConnection.value = null
    testConnection.value = null
    
    Object.keys(connectionStatus.value).forEach(key => {
      connectionStatus.value[key] = 'disconnected'
    })
    
    console.log('🔌 All WebSockets disconnected')
  }

  const updateConfig = (newConfig) => {
    config.value = { ...config.value, ...newConfig }
  }

  // Очистка при уничтожении компонента
  const cleanup = () => {
    disconnectWebSocket()
  }

  return {
    // State
    objects,
    activeControl,
    activeSubscriptions,
    connectionStatus,
    testMessages,
    config,
    messagesReceived,
    messagesSent,
    isWebSocketInitialized, // ← ДОБАВЛЕНО для совместимости
    
    // Getters  
    isConnected,
    subscribedObjects, // ОСНОВНОЙ - объекты по подпискам
    recentTestMessages,
    
    // Actions
    // Инициализация (для совместимости)
    initializeWebSocket,
    disconnectWebSocket,
    
    // Подписки
    subscribe,
    unsubscribe,
    subscribeMultiple,
    unsubscribeMultiple,
    clearAllSubscriptions,
    
    // Управление
    openControl,
    closeControl,
    sendCommand,
    
    // Соединение
    connectAll,
    disconnectAll,
    updateConfig,
    cleanup
  }
})

/*
// В компоненте - ВСЕ РАБОТАЕТ КАК РАНЬШЕ
import { useObjectsStore } from '@/stores/objects'

const objectsStore = useObjectsStore()

// Старый код работает:
onMounted(async () => {
  // Инициализация как раньше
  await objectsStore.initializeWebSocket()
  objectsStore.subscribe(props.id)
})

// Геттеры работают:
const objData = objectsStore.objects[props.id] // прямой доступ к objects
// или
const objData = objectsStore.subscribedObjects[props.id] // через подписки
*/