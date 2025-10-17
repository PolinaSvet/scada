// utils/screenDebug.js
export class ScreenSizeHelper {
  static getContainerSize() {
    const container = document.querySelector('.main-container')
    if (!container) return null
    
    const rect = container.getBoundingClientRect()
    return {
      width: rect.width,
      height: rect.height
    }
  }
  
  static calculateIdealSize(padding = 40) {
    const containerSize = this.getContainerSize()
    if (!containerSize) return null
    
    return {
      width: Math.floor(containerSize.width - padding),
      height: Math.floor(containerSize.height - padding)
    }
  }
  
  static logRecommendation() {
    const ideal = this.calculateIdealSize()
    if (!ideal) return
    
    console.log('🎯 Рекомендация по размерам экрана:')
    console.log('📏 Текущий контейнер:', this.getContainerSize())
    console.log('💡 Идеальный размер:', ideal)
    console.log('⚙️  Для применения в stores/layout.js:')
    console.log(`screenBaseSize: ref({ width: ${ideal.width}, height: ${ideal.height} })`)
  }
}

// Делаем доступной глобально для отладки
if (typeof window !== 'undefined') {
  window.ScreenSizeHelper = ScreenSizeHelper
}

/*
!!! Где задаются и просматриваются размеры окон
Размеры задаются в двух местах:
1. В stores/layout.js - базовый размер всех экранов:

javascript
// stores/layout.js
const screenBaseSize = ref({ width: 2000, height: 1500 })
2. В каждом экране (ScreenSensors1.vue, ScreenSensors2.vue) - CSS размеры:

css
// screen-base.css 
.screen-sensors {
  width: 2000px;
  height: 1500px;
  ...
}
*/