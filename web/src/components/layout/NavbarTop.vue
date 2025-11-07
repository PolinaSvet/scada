<template>
  <header class="navbar-top">
    <div class="navbar-content">
      <!-- Левая часть - логотип и бренд -->
      <div class="brand">
        <div class="logo">⚙️</div>
        <span class="brand-text">SCADA SYSTEM</span>
      </div>

      <!-- Центральная часть - навигация/статус -->
      <div class="navbar-center">
        <div class="screen-title">{{ currentScreenTitle }}</div>
      </div>

      <!-- Правая часть - пользователь и управление -->
      <div class="navbar-right">
        <button class="navbar-menu" @click="toggleMenu" title="Меню">
          ⚙️
        </button>
        <div class="user-info">
          <span class="user-name">Operator</span>
          <div class="data-indicator" :class="connectionStatus"></div>
        </div>
      </div>
    </div>
  </header>
</template>

<script>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useLayoutStore } from '@/stores/layout'

export default {
  name: 'NavbarTop',
  setup() {
    const layoutStore = useLayoutStore()
    
    const connectionStatus = ref('online')

    // Используем центральную конфигурацию для заголовка
    const currentScreenTitle = computed(() => 
      layoutStore.currentScreenConfig?.title || 'SCADA System'
    )

    const toggleMenu = () => {
      console.log('Открыто меню системы')
    }

    const simulateConnectionChanges = () => {
      return setInterval(() => {
        const statuses = ['online', 'connecting', 'online', 'online', 'offline']
        const randomStatus = statuses[Math.floor(Math.random() * statuses.length)]
        connectionStatus.value = randomStatus
        
      }, 10000)
    }

    onMounted(() => {
      const connectionInterval = simulateConnectionChanges()
      onUnmounted(() => {
        clearInterval(connectionInterval)
      })
    })

    return {
      connectionStatus,
      currentScreenTitle,
      toggleMenu
    }
  }
}
</script>

<style scoped>
@import '@/assets/styles/navbar-top.css';
</style>