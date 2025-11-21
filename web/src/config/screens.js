// config/screens.js
import ScreenDashboard from '@/components/screens/ScreenDashboard.vue'
import ScreenMain from '@/components/screens/ScreenMain.vue'
import ScreenSensors1 from '@/components/screens/ScreenSensors1.vue'
import ScreenSensors2 from '@/components/screens/ScreenSensors2.vue'
import ScreenSensors3 from '@/components/screens/ScreenSensors3.vue'
import ScreenAlarms from '@/components/screens/ScreenAlarms.vue'
import ScreenAlarmsHist from '@/components/screens/ScreenAlarmsHist.vue'
import ScreenTrendsHist from '@/components/screens/ScreenTrendsHist.vue'

// Импорты SVG
import MnemoMain from '@/assets/svg/screens/MnemoMain.svg?raw'
import MnemoSchema1 from '@/assets/svg/screens/MnemoSchema1.svg?raw'
import MnemoSchema2 from '@/assets/svg/screens/MnemoSchema2.svg?raw'
import MnemoSchema3 from '@/assets/svg/screens/MnemoSchema3.svg?raw'

export const SCREENS_CONFIG = {
  ScreenMain: {
    id: 'ScreenMain',
    label: 'ОСНОВНОЙ ЭКРАН',
    icon: '🏠',
    title: 'ОСНОВНОЙ ЭКРАН',
    component: ScreenMain,
    svgSchema: MnemoMain,
    objectTypes: ['obj-sensor-anchor']
  },
  ScreenSensors1: {
    id: 'ScreenSensors1',
    label: 'ОПЕРАТОРНАЯ',
    icon: '023',
    title: 'МЕСТНАЯ ОПЕРАТОРНАЯ (023)',
    component: ScreenSensors1,
    svgSchema: MnemoSchema1,
    objectTypes: ['obj-sensor-anchor']
  },
  ScreenSensors2: {
    id: 'ScreenSensors2', 
    label: 'СКЛАД',
    icon: '101',
    title: 'СКЛАДСКОЙ КОМПЛЕКС (101)',
    component: ScreenSensors2,
    svgSchema: MnemoSchema2,
    objectTypes: ['obj-sensor-anchor']
  },
  ScreenSensors3: {
    id: 'ScreenSensors3', 
    label: 'СКЛАД',
    icon: '102',
    title: 'СКЛАДСКОЙ КОМПЛЕКС (102)',
    component: ScreenSensors3,
    svgSchema: MnemoSchema3,
    objectTypes: ['obj-sensor-anchor']
  },
  alarms: {
    id: 'alarms',
    label: 'СОБЫТИЯ',
    icon: '🚨',
    title: 'СОБЫТИЯ СИСТЕМЫ',
    component: ScreenAlarms
  },
  alarmsHist: {
    id: 'alarmsHist',
    label: 'ИСТОРИЯ',
    icon: '📋',
    title: 'ЖУРНАЛ ИСТОРИЧЕСКИХ СОБЫТИЙ',
    component: ScreenAlarmsHist
  },
  trendsHist: {
    id: 'trendsHist',
    label: 'ТРЕНДЫ',
    icon: '📊',
    title: 'ТРЕНДЫ',
    component: ScreenTrendsHist
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