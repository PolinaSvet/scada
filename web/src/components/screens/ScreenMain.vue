<template>
  <div class="screen-sensors screen-1">
    <div class="sensors-area" ref="sensorsArea">
      <div class="svg-container" ref="svgContainer"></div>
      <div class="sensors-container" ref="sensorsContainer"></div>
    </div>
    
    <component
      v-for="control in activeControls"
      :key="control.id"
      :is="control.component"
      :id="control.id"
      :isOpen="true"
      @close="closeControl"
    />
  </div>
</template>

<script>
import { computed, onMounted, onUnmounted, ref, watch, shallowRef } from 'vue'
import { useObjectsStore } from '@/stores/objects'
import { svgScreenLoader } from '@/utils/svgScreenLoader'
import { getScreenSvgSchema } from '@/config/screens'

export default {
  name: 'ScreenSensors1',
  setup() {
    const objectsStore = useObjectsStore()
    const svgContainer = ref(null)
    const sensorsContainer = ref(null)
    
    const screenObjects = ref([])
    const controlComponents = shallowRef({})
    const activeControls = shallowRef([])

    const activeControl = computed(() => objectsStore.activeControl)

    const closeControl = () => {
      objectsStore.closeControl()
    }

    const updateActiveControls = () => {
      activeControls.value = []
      
      if (activeControl.value) {
        const objectId = activeControl.value
        const controlComponent = svgScreenLoader.getControlComponentForObject(objectId)
        
        if (controlComponent) {
          activeControls.value.push({
            id: objectId,
            component: controlComponent
          })
        } else {
          console.warn(`⚠️ No control component found for object: ${objectId}`)
        }
      }
    }

    onMounted(async () => {
      console.log('🖥️ ScreenMain mounted')
      
      try {
        const svgSchema = getScreenSvgSchema('ScreenMain')
        
        if (svgSchema && svgContainer.value && sensorsContainer.value) {
          // Загружаем SVG и создаем компоненты
          screenObjects.value = await svgScreenLoader.loadSvgScreen(
            svgSchema,
            svgContainer.value,
            sensorsContainer.value
          )
          
          // Получаем созданные компоненты управления
          const controls = await svgScreenLoader.createControlComponents()
          controlComponents.value = Object.fromEntries(controls)
          
          // Подписываемся на объекты (проверяем дубликаты)
          if (screenObjects.value.length > 0) {
            console.log('📝 Subscribing to objects:', screenObjects.value)
            objectsStore.subscribeMultiple(screenObjects.value)
          }
        } else {
          console.warn('⚠️ No SVG schema found for ScreenSensors1')
        }
        
        // Инициализируем WebSocket (если еще не инициализирован)
        if (!objectsStore.isWebSocketInitialized) {
          await objectsStore.initializeWebSocket()
        }
        
      } catch (error) {
        console.error('❌ Error initializing screen:', error)
      }
    })

    const stopWatch = watch(activeControl, updateActiveControls, { immediate: true })

    onUnmounted(() => {
      // Останавливаем watcher
      stopWatch()
      
      // Отписываемся от объектов
      if (screenObjects.value.length > 0) {
        console.log('📝 Unsubscribing from objects:', screenObjects.value)
        objectsStore.unsubscribeMultiple(screenObjects.value)
      }
      
      // Очищаем загрузчик
      svgScreenLoader.cleanup()
      
      console.log('🖥️ ScreenSensors1 unmounted')
    })

    return {
      activeControl,
      activeControls,
      closeControl,
      svgContainer,
      sensorsContainer
    }
  }
}
</script>

<style scoped>
@import '@/assets/styles/screen-base.css';
</style>