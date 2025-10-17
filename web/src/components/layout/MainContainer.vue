<template>
  <main class="main-container" :style="containerStyle" ref="mainContainerRef">
    <!-- Индикатор режима -->
    <div class="mode-indicator">
      {{ isAutoScaleMode ? `Автомасштаб: ${Math.round(zoomLevel * 100)}%` : 'Скроллбары' }}
    </div>

    <transition name="fade" mode="out-in">
      <div 
        class="screen-wrapper"
        :class="wrapperClass"
        ref="screenWrapperRef"
      >
        <div 
          class="screen-content"
          :style="screenContentStyle"
        >
          <component 
            :is="currentScreenComponent" 
            :key="currentScreen"
          />
        </div>
      </div>
    </transition>
  </main>
</template>

<script>
import { computed, onMounted, ref, onUnmounted, nextTick } from 'vue'
import { useLayoutStore } from '@/stores/layout'
import { useObjectsStore } from '@/stores/objects'
import { screenComponents } from '@/config/screens'

export default {
  name: 'MainContainer',
  setup() {
    const layoutStore = useLayoutStore()
    const objectsStore = useObjectsStore()
    
    const mainContainerRef = ref(null)
    const screenWrapperRef = ref(null)
    const zoomLevel = ref(1)
    
    let resizeObserver = null

    const isAutoScaleMode = computed(() => layoutStore.screenViewMode === 'autoScale')

    // Стили контейнера
    const containerStyle = computed(() => {
      const leftMargin = layoutStore.leftPanelWidth
      const bottomMargin = layoutStore.bottomPanelHeight
      const topMargin = 60

        
      return {
        marginTop: `${topMargin}px`,
        marginLeft: `${leftMargin}px`,
        marginBottom: `${bottomMargin}px`,
        padding: '10px',
        height: `calc(100vh - ${topMargin + bottomMargin}px)`,
      }
    })

    // Класс обертки
    const wrapperClass = computed(() => 
      isAutoScaleMode.value ? 'screen-wrapper auto-scale' : 'screen-wrapper scroll-mode'
    )

    // Стили контента экрана
    const screenContentStyle = computed(() => {
      const baseStyle = {
        width: `${layoutStore.screenBaseSize.width}px`,
        height: `${layoutStore.screenBaseSize.height}px`,
      }

      if (isAutoScaleMode.value) {
        return {
          ...baseStyle,
          transform: `scale(${zoomLevel.value})`,
          transformOrigin: 'top left'
        }
      } else {
        return {
          ...baseStyle,
          minWidth: `${layoutStore.screenBaseSize.width}px`,
          minHeight: `${layoutStore.screenBaseSize.height}px`
        }
      }
    })

    // Функция расчета масштаба
    const calculateZoom = () => {
      if (!screenWrapperRef.value || !isAutoScaleMode.value) return
      
      const wrapper = screenWrapperRef.value
      const wrapperWidth = wrapper.clientWidth
      const wrapperHeight = wrapper.clientHeight
      
      const scaleX = wrapperWidth / layoutStore.screenBaseSize.width
      const scaleY = wrapperHeight / layoutStore.screenBaseSize.height
      
      zoomLevel.value = Math.min(scaleX, scaleY)* 1//0.95
    }

    // Инициализация ResizeObserver
    const initResizeObserver = () => {
      if (screenWrapperRef.value) {
        resizeObserver = new ResizeObserver(() => {
          nextTick(() => {
            calculateZoom()
          })
        })
        resizeObserver.observe(screenWrapperRef.value)
      }
    }

    const currentScreenComponent = computed(() => 
      screenComponents[layoutStore.currentScreen] || screenComponents.dashboard
    )

    onMounted(() => {
      objectsStore.initializeWebSocket()
      //objectsStore.initializeScreens()

      nextTick(() => {
        initResizeObserver()
        calculateZoom()
      })
    })

    onUnmounted(() => {
      if (resizeObserver && screenWrapperRef.value) {
        resizeObserver.unobserve(screenWrapperRef.value)
      }

      // Полная очистка при закрытии приложения
      objectsStore.cleanup()
    })

    return {
      currentScreen: computed(() => layoutStore.currentScreen),
      currentScreenComponent,
      containerStyle,
      wrapperClass,
      screenContentStyle,
      isAutoScaleMode,
      zoomLevel,
      mainContainerRef,
      screenWrapperRef
    }
  }
}
</script>

<style scoped>
@import '@/assets/styles/main-container.css';
</style>