// stores/layout.js
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { SCREENS_CONFIG, getScreenConfig } from '@/config/screens'

export const useLayoutStore = defineStore('layout', () => {
  const currentScreen = ref('dashboard')
  const leftPanelState = ref('expanded')
  const bottomPanelState = ref('minimized')
  const screenViewMode = ref('autoScale')
  const screenBaseSize = ref({ width: 1632, height: 622 })
  //const screenBaseSize = ref({ width: 2000, height: 1500 })
  
  const setCurrentScreen = (screenName) => {
    if (SCREENS_CONFIG[screenName]) {
      currentScreen.value = screenName
    } else {
      console.warn(`Экран "${screenName}" не найден в конфигурации`)
      currentScreen.value = 'dashboard'
    }
  }

  const setLeftPanelState = (state) => {
    leftPanelState.value = state
  }

  const setBottomPanelState = (state) => {
    bottomPanelState.value = state
  }

  const toggleScreenViewMode = () => {
    screenViewMode.value = screenViewMode.value === 'autoScale' ? 'scroll' : 'autoScale'
  }

  // Текущая конфигурация экрана
  const currentScreenConfig = computed(() => getScreenConfig(currentScreen.value))

  const leftPanelWidth = computed(() => {
    switch (leftPanelState.value) {
      case 'minimized': return 60
      case 'expanded': return 200
      case 'maximized': return 400
      default: return 200
    }
  })

  const bottomPanelHeight = computed(() => {
    switch (bottomPanelState.value) {
      case 'minimized': return 40
      case 'expanded200': return 150
      case 'maximized': return 400
      default: return 40
    }
  })

  return {
    currentScreen,
    currentScreenConfig,
    leftPanelState,
    bottomPanelState,
    screenViewMode,
    screenBaseSize,
    leftPanelWidth,
    bottomPanelHeight,
    setCurrentScreen,
    setLeftPanelState,
    setBottomPanelState,
    toggleScreenViewMode
  }
})