<template>
  <aside class="navbar-left" :class="panelClass">
    <!-- Панель управления -->
    <div class="left-controls">
      <div class="control-buttons" :class="buttonsLayout">
        <button class="control-btn icon-minimize" @click="minimizePanel" title="Свернуть">
        </button>
        <button class="control-btn icon-expand" @click="expandTo200" title="Развернуть до 200px">
        </button>
        <button class="control-btn icon-maximize" @click="maximizePanel" title="Максимально развернуть">
        </button>
        <button 
          class="control-btn icon-viewmode" 
          @click="toggleViewMode" 
          :title="viewModeTitle"
          :class="{ 'active': screenViewMode === 'scroll' }"
        >
        </button>
      </div>
    </div>

    <!-- Дата и время -->
    <div v-if="panelState !== 'minimized'" class="time-display" :class="panelState">
      {{ formattedTime }}
    </div>

    <!-- Основные кнопки навигации -->
    <nav class="screen-nav" :class="navLayout">
      <button
        v-for="screen in primaryScreens"
        :key="screen.id"
        class="nav-button"
        :class="[panelState, { active: currentScreen === screen.id }]"
        @click="switchScreen(screen.id)"
        :title="screen.label"
      >
        <span class="nav-icon">{{ screen.icon }}</span>
        <span class="nav-label">{{ screen.label }}</span>
      </button>
    </nav>

    <!-- Второй ряд кнопок -->
    <div v-if="panelState === 'maximized'" class="secondary-buttons">
      <button
        v-for="button in secondaryButtons"
        :key="button.id"
        class="secondary-button"
        @click="handleSecondaryClick(button.id)"
        :title="button.label"
      >
        {{ button.label }}
      </button>
    </div>
  </aside>
</template>

<script>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useLayoutStore } from '@/stores/layout'
import { getPrimaryScreens } from '@/config/screens'

export default {
  name: 'NavbarLeft',
  setup() {
    const layoutStore = useLayoutStore()
    
    const panelState = ref('expanded')
    const currentTime = ref(new Date())

    // Используем центральную конфигурацию
    const primaryScreens = getPrimaryScreens()
    const currentScreen = computed(() => layoutStore.currentScreen)
    const screenViewMode = computed(() => layoutStore.screenViewMode)

    const panelClass = computed(() => ({
      'minimized': panelState.value === 'minimized',
      'expanded': panelState.value === 'expanded',
      'maximized': panelState.value === 'maximized'
    }))

    const buttonsLayout = computed(() => 
      panelState.value === 'minimized' ? 'vertical' : 'horizontal'
    )

    const navLayout = computed(() => 
      panelState.value === 'maximized' ? 'grid-layout' : 'vertical'
    )

    const formattedTime = computed(() => {
      const now = currentTime.value
      return now.toLocaleDateString('ru-RU') + ' ' + now.toLocaleTimeString('ru-RU')
    })

    const viewModeTitle = computed(() => 
      screenViewMode.value === 'autoScale' ? 'Автомасштаб (переключить на скролл)' : 'Скроллбары (переключить на автомасштаб)'
    )

    const secondaryButtons = [
      { id: 'export', label: 'Экспорт' },
      { id: 'import', label: 'Импорт' },
      { id: 'backup', label: 'Резервная копия' },
      { id: 'restore', label: 'Восстановление' },
      { id: 'logs', label: 'Логи' },
      { id: 'help', label: 'Помощь' }
    ]

    const minimizePanel = () => {
      panelState.value = 'minimized'
      layoutStore.setLeftPanelState('minimized')
    }

    const expandTo200 = () => {
      panelState.value = 'expanded'
      layoutStore.setLeftPanelState('expanded')
    }

    const maximizePanel = () => {
      panelState.value = 'maximized'
      layoutStore.setLeftPanelState('maximized')
    }

    const toggleViewMode = () => {
      layoutStore.toggleScreenViewMode()
    }

    const switchScreen = (screenId) => {
      layoutStore.setCurrentScreen(screenId)
    }

    const handleSecondaryClick = (buttonId) => {
      console.log('Нажата кнопка:', buttonId)
    }

    const startTimeUpdates = () => {
      return setInterval(() => {
        currentTime.value = new Date()
      }, 1000)
    }

    onMounted(() => {
      const timeInterval = startTimeUpdates()
      onUnmounted(() => {
        clearInterval(timeInterval)
      })
    })

    return {
      panelState,
      currentScreen,
      screenViewMode,
      panelClass,
      buttonsLayout,
      navLayout,
      formattedTime,
      viewModeTitle,
      primaryScreens,
      secondaryButtons,
      minimizePanel,
      expandTo200,
      maximizePanel,
      toggleViewMode,
      switchScreen,
      handleSecondaryClick
    }
  }
}
</script>

<style scoped>
@import '@/assets/styles/navbar-left.css';
</style>