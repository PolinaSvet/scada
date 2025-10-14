import { defineStore } from 'pinia'
import { decode, encode } from '@msgpack/msgpack'
import { ref, computed } from 'vue'

export const useWebSocketStore = defineStore('websocket', () => {
  // State
  const dataConnection = ref(null)
  const controlConnection = ref(null)
  const testConnection = ref(null)
  
  const objectData = ref([])
  const testData = ref([])
  const connectionStatus = ref({
    data: 'disconnected',
    control: 'disconnected',
    test: 'disconnected'
  })
  
  const messagesReceived = ref(0)
  const messagesSent = ref(0)
  
  const config = ref({
    dataPort: 8081,
    controlPort: 8082,
    testPort: 8083,
    clientId: 'test_client_' + Math.random().toString(36).substr(2, 9),
    userId: 'test_user',
    clientType: 'full'
  })

  // Getters
  const isConnected = computed(() => {
    return Object.values(connectionStatus.value).every(status => status === 'connected')
  })
  
  const connectionStats = computed(() => {
    return {
      data: connectionStatus.value.data,
      control: connectionStatus.value.control,
      test: connectionStatus.value.test,
      messagesReceived: messagesReceived.value,
      messagesSent: messagesSent.value
    }
  })

  // Получаем только последние 5 объектов для отображения
  const lastFiveObjects = computed(() => {
    return objectData.value.slice(-5)
  })

  // Actions
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
      console.log(`${type} WebSocket connected`)
      
      // Сохраняем соединение
      if (type === 'data') dataConnection.value = connection
      else if (type === 'control') controlConnection.value = connection
      else if (type === 'test') testConnection.value = connection
    }
    
    connection.onmessage = (event) => {
      handleMessage(type, event.data)
    }
    
    connection.onclose = () => {
      connectionStatus.value[type] = 'disconnected'
      console.log(`${type} WebSocket disconnected`)
    }
    
    connection.onerror = (error) => {
      connectionStatus.value[type] = 'error'
      console.error(`${type} WebSocket error:`, error)
    }
  }
  
  const handleMessage = async (type, data) => {
    try {
      messagesReceived.value++
      
      let message
      if (typeof data === 'string') {
        // JSON сообщение
        message = JSON.parse(data)
      } else {
        // MessagePack сообщение - данные приходят как ArrayBuffer
        const uint8Array = new Uint8Array(data)
        message = decode(uint8Array)
      }
      
      console.log(`Received ${type} message:`, message)
      processDecodedMessage(type, message)
      
    } catch (error) {
      console.error('Error processing message:', error)
    }
  }
  
  const processDecodedMessage = (type, message) => {
    switch (type) {
      case 'data':
        if (message.type === 'data_batch' || message.type === 'updateObjectsBatch') {
          // Данные объектов
          let objectDataDecoded
          objectDataDecoded = message.data

          objectData.value.unshift({
            message: objectDataDecoded.ID,
            time: message.time || new Date().toISOString(),
            received: new Date().toLocaleTimeString()
          })
          
          // Ограничиваем количество сохраняемых тестовых данных
          if (objectData.value.length > 20) {
            objectData.value = objectData.value.slice(0, 20)
          }
          
          
        } else {
          console.log('Other data message type:', message.type)
        }
        break
        
      case 'test':
        if (message.type === 'test_data') {
          // Тестовые данные
          try {
            console.log('Processing test data')
            
            // message.data - это Uint8Array с тестовыми данными
            let testDataDecoded
            testDataDecoded = message.data
            
            testData.value.unshift({
              message: testDataDecoded,
              time: message.time || new Date().toISOString(),
              received: new Date().toLocaleTimeString()
            })
            
            // Ограничиваем количество сохраняемых тестовых данных
            if (testData.value.length > 20) {
              testData.value = testData.value.slice(0, 20)
            }
            
            console.log('Test data updated, count:', testData.value.length)
            console.log('Latest test data:', testData.value[0])
            
          } catch (error) {
            console.error('Error decoding test data:', error)
          }
        } else {
          console.log('Other test message type:', message.type)
        }
        break
        
      case 'control':
        // Обработка ответов на команды
        console.log('Control message:', message)
        break
    }
  }
  
  const sendCommand = async (objectId, command, data = {}) => {
    if (connectionStatus.value.control !== 'connected') {
      console.error('Control WebSocket not connected')
      return
    }
    
    const commandMessage = {
      command: command,
      objectId: objectId,
      userId: config.value.userId,
      data: data,
      time: new Date().toISOString()
    }
    
    try {
      // Упаковываем в MessagePack
      const encodedMessage = encode(commandMessage)
      controlConnection.value.send(encodedMessage)
      messagesSent.value++
      console.log('Command sent:', commandMessage)
    } catch (error) {
      console.error('Error sending command:', error)
    }
  }
  
  const connectAll = () => {
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
  }
  
  const updateConfig = (newConfig) => {
    config.value = { ...config.value, ...newConfig }
  }

  return {
    // State
    dataConnection,
    controlConnection,
    testConnection,
    objectData,
    testData,
    connectionStatus,
    messagesReceived,
    messagesSent,
    config,
    
    // Getters
    isConnected,
    connectionStats,
    lastFiveObjects, // Добавляем геттер для последних 5 объектов
    
    // Actions
    connectWebSocket,
    handleMessage,
    processDecodedMessage,
    sendCommand,
    connectAll,
    disconnectAll,
    updateConfig
  }
})