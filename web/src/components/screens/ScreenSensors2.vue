<template>
  <div class="screen-sensors screen-1">
    <div class="sensors-area" ref="sensorsArea">
      <!-- SVG фон будет загружен динамически -->
      <div class="svg-container" ref="svgContainer"></div>
      
      <!-- Контейнер для динамических объектов -->
      <div class="sensors-container" ref="sensorsContainer"></div>
    </div>
    
    <CtrlObjSens 
      v-if="activeControl" 
      :id="activeControl" 
      :isOpen="!!activeControl" 
      @close="closeControl" 
    />
  </div>
</template>

<script>
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useObjectsStore } from '@/stores/objects'
import ObjSens from '@/components/ObjSens.vue'
import CtrlObjSens from '@/components/CtrlObjSens.vue'
import { h, createApp } from 'vue'

// Импортируем SVG как raw текст
import mnemoSchemaSvg from '@/assets/svg/screens/MnemoSchema1.svg?raw'

export default {
  name: 'ScreenSensors1',
  components: {
    ObjSens,
    CtrlObjSens
  },
  setup() {
    const objectsStore = useObjectsStore()
    const svgContainer = ref(null)
    const sensorsContainer = ref(null)
    const sensorsArea = ref(null)
    
    const screenObjects = ref([]) // Динамически собираемый список объектов
    const activeControl = computed(() => objectsStore.activeControl)

    const closeControl = () => {
      objectsStore.closeControl()
    }

    // Функция для загрузки и парсинга SVG
    const loadSvgSchema = async () => {
      try {
        // Вставляем SVG в контейнер
        if (svgContainer.value) {
          svgContainer.value.innerHTML = mnemoSchemaSvg
          
          // Парсим якоря и создаем компоненты
          await parseAnchorsAndCreateComponents()
        }
      } catch (error) {
        console.error('❌ Error loading SVG schema:', error)
      }
    }

    // Парсинг якорей и создание компонентов
    const parseAnchorsAndCreateComponents = () => {
      return new Promise((resolve) => {
        // Даем время SVG отрендериться
        setTimeout(() => {
          const svgElement = svgContainer.value?.querySelector('svg')
          if (!svgElement) {
            console.error('❌ SVG element not found')
            resolve()
            return
          }

          // Ищем все якоря с data-type="obj-sensor-anchor"
          const anchors = svgElement.querySelectorAll('[data-type="obj-sensor-anchor"]')
          console.log(`🔍 Found ${anchors.length} anchors in SVG`)

          screenObjects.value = [] // Очищаем список

          anchors.forEach((anchor, index) => {
            const objectId = anchor.getAttribute('data-id')
            const objectType = anchor.getAttribute('data-type')
            
            if (!objectId) {
              console.warn('⚠️ Anchor without data-id found:', anchor)
              return
            }

            console.log(`📍 Processing anchor ${index + 1}:`, { objectId, objectType })

            // Получаем координаты и размеры якоря
            const rect = anchor.getBoundingClientRect()
            const svgRect = svgElement.getBoundingClientRect()
            
            // Относительные координаты внутри sensors-area
            const x = rect.left - svgRect.left
            const y = rect.top - svgRect.top
            const width = rect.width
            const height = rect.height

            console.log(`📐 Anchor dimensions:`, { x, y, width, height })

            // Создаем компонент в зависимости от типа
            if (objectType === 'obj-sensor-anchor') {
              createObjSensComponent(objectId, x, y, width, height)
            } else {
              console.warn(`⚠️ Unknown object type: ${objectType}`)
            }

            // Добавляем в список для подписки
            screenObjects.value.push(objectId)
          })

          console.log(`✅ Created ${screenObjects.value.length} dynamic components`)
          resolve()
        }, 100)
      })
    }

    // Создание компонента ObjSens
    const createObjSensComponent = (objectId, x, y, width, height) => {
      if (!sensorsContainer.value) return

      // Создаем контейнер для компонента
      const componentContainer = document.createElement('div')
      componentContainer.className = 'dynamic-component'
      componentContainer.style.position = 'absolute'
      componentContainer.style.left = `${x}px`
      componentContainer.style.top = `${y}px`
      componentContainer.style.width = `${width}px`
      componentContainer.style.height = `${height}px`
      componentContainer.style.pointerEvents = 'auto'

      sensorsContainer.value.appendChild(componentContainer)

      // Создаем и монтируем компонент
      const Component = {
        render() {
          return h(ObjSens, {
            id: objectId,
            x: 0,
            y: 0,
            w: width,
            h: height
          })
        }
      }

      createApp(Component).mount(componentContainer)
      console.log(`✅ Created ObjSens: ${objectId} at [${x}, ${y}] ${width}x${height}`)
    }

    // Очистка динамически созданных компонентов
    const cleanupDynamicComponents = () => {
      if (sensorsContainer.value) {
        const components = sensorsContainer.value.querySelectorAll('.dynamic-component')
        components.forEach(component => {
          component.remove()
        })
        console.log(`🧹 Removed ${components.length} dynamic components`)
      }
    }

    onMounted(async () => {
      console.log('🖥️ ScreenSensors1 mounted')
      
      // Загружаем SVG схему
      await loadSvgSchema()
      
      // Подписываемся на объекты этого экрана
      if (screenObjects.value.length > 0) {
        objectsStore.subscribeMultiple(screenObjects.value)
      }
      
      // Инициализируем WebSocket
      await objectsStore.initializeWebSocket()
    })

    onUnmounted(() => {
      // Отписываемся от объектов при уничтожении экрана
      if (screenObjects.value.length > 0) {
        objectsStore.unsubscribeMultiple(screenObjects.value)
      }
      
      // Очищаем динамически созданные компоненты
      cleanupDynamicComponents()
      
      console.log('🖥️ ScreenSensors1 unmounted')
    })

    return {
      activeControl,
      closeControl,
      svgContainer,
      sensorsContainer,
      sensorsArea
    }
  }
}
</script>

<style scoped>
@import '@/assets/styles/screen-base.css';
/*
.sensors-area {
  position: relative;
  width: 100%;
  height: 100%;
  overflow: hidden;
}

.svg-container {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  pointer-events: none;
}

.svg-container :deep(svg) {
  width: 100%;
  height: 100%;
}

.sensors-container {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  pointer-events: none;
}

.sensors-container > * {
  pointer-events: auto;
}*/
</style>