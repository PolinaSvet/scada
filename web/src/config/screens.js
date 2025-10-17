// config/screens.js
import ScreenDashboard from '@/components/screens/ScreenDashboard.vue'
import ScreenSensors1 from '@/components/screens/ScreenSensors1.vue'
import ScreenSensors2 from '@/components/screens/ScreenSensors2.vue'
import ScreenAlarms from '@/components/screens/ScreenAlarms.vue'

// Импорты SVG
import MnemoSchema1 from '@/assets/svg/screens/MnemoSchema1.svg?raw'
// import MnemoSchema2 from '@/assets/svg/screens/MnemoSchema2.svg?raw'

export const SCREENS_CONFIG = {
  dashboard: {
    id: 'dashboard',
    label: 'Главная',
    icon: '🏠',
    title: 'Главная панель',
    component: ScreenDashboard
  },
  ScreenSensors1: {
    id: 'ScreenSensors1',
    label: 'Датчики 1',
    icon: '📊',
    title: 'Датчики 1',
    component: ScreenSensors1,
    svgSchema: MnemoSchema1,
    objectTypes: ['obj-sensor-anchor']
  },
  ScreenSensors2: {
    id: 'ScreenSensors2', 
    label: 'Датчики 2',
    icon: '📈',
    title: 'Датчики 2',
    component: ScreenSensors2,
    svgSchema: null, // Будет добавлен позже
    objectTypes: ['obj-sensor-anchor']
  },
  alarms: {
    id: 'alarms',
    label: 'Аварии',
    icon: '🚨',
    title: 'Аварийные события',
    component: ScreenAlarms
  }
}

// Вспомогательные функции
export const getScreenConfig = (screenId) => SCREENS_CONFIG[screenId] || SCREENS_CONFIG.dashboard

export const getScreenSvgSchema = (screenId) => {
  const config = getScreenConfig(screenId)
  return config.svgSchema || null
}

export const getScreenObjectTypes = (screenId) => {
  const config = getScreenConfig(screenId)
  return config.objectTypes || []
}

export const getPrimaryScreens = () => Object.values(SCREENS_CONFIG)

export const getScreenTitle = (screenId) => getScreenConfig(screenId).title

export const getScreenComponent = (screenId) => getScreenConfig(screenId).component

// Экспортируем все компоненты для MainContainer
export const screenComponents = Object.values(SCREENS_CONFIG).reduce((components, screen) => {
  components[screen.id] = screen.component
  return components
}, {})