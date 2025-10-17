// utils/svgScreenLoader.js
import { h, createApp } from 'vue'

// Карта компонентов по типам
const COMPONENT_MAP = {
  'obj-sensor-anchor': () => import('@/components/ObjSens.vue'),
  // 'obj-valve-anchor': () => import('@/components/ObjValve.vue'),
  // 'obj-pump-anchor': () => import('@/components/ObjPump.vue'),
}

// Карта окон управления по типам объектов
const CONTROL_MAP = {
  'obj-sensor-anchor': () => import('@/components/CtrlObjSens.vue'),
  // 'obj-valve-anchor': () => import('@/components/CtrlObjValve.vue'),
  // 'obj-pump-anchor': () => import('@/components/CtrlObjPump.vue'),
}

export class SvgScreenLoader {
  constructor(options = {}) {
    this.options = {
      showAnchorPlaceholder: true,
      ...options
    }
    this.dynamicComponents = []
    this.resizeObserver = null
    this.controlComponents = new Map()
  }

  // Основной метод загрузки SVG и создания компонентов
  async loadSvgScreen(svgRaw, containerRef, sensorsContainerRef) {
    try {
      containerRef.innerHTML = svgRaw
      const screenObjects = await this.parseAnchorsAndCreateComponents(containerRef, sensorsContainerRef)
      return screenObjects
    } catch (error) {
      console.error('❌ Error loading SVG screen:', error)
      throw error
    }
  }

  // Парсинг якорей и создание компонентов
  async parseAnchorsAndCreateComponents(containerRef, sensorsContainerRef) {
    return new Promise((resolve) => {
      setTimeout(async () => {
        const svgElement = containerRef?.querySelector('svg')
        if (!svgElement) {
          console.error('❌ SVG element not found')
          resolve([])
          return
        }

        this.setupResizeObserver(svgElement, sensorsContainerRef)

        const anchors = svgElement.querySelectorAll('[data-type]')
        console.log(`🔍 Found ${anchors.length} anchors in SVG`)

        const screenObjects = []

        for (const anchor of anchors) {
          const objectId = anchor.getAttribute('data-id')
          const objectType = anchor.getAttribute('data-type')
          
          if (!objectId || !objectType) {
            console.warn('⚠️ Anchor missing data-id or data-type:', anchor)
            continue
          }

          try {
            await this.createComponentForAnchor(
              objectId, 
              objectType, 
              anchor, 
              svgElement, 
              sensorsContainerRef
            )
            screenObjects.push(objectId)
          } catch (error) {
            console.error(`❌ Error creating component for ${objectId}:`, error)
          }
        }

        // Создаем компоненты управления после создания всех объектов
        await this.createControlComponents()

        console.log(`✅ Created ${this.dynamicComponents.length} dynamic components`)
        resolve(screenObjects)
      }, 100)
    })
  }

  // Создание компонента для якоря
  async createComponentForAnchor(objectId, objectType, anchor, svgElement, sensorsContainerRef) {
    // Получаем координаты и размеры якоря через getBoundingClientRect()
    const rect = anchor.getBoundingClientRect()
    const svgRect = svgElement.getBoundingClientRect()
    
    // Относительные координаты внутри sensors-area
    const x = rect.left - svgRect.left
    const y = rect.top - svgRect.top
    const width = rect.width
    const height = rect.height

    console.log(`📍 Created ${objectType}: ${objectId} at [${x}, ${y}] ${width}x${height}`)

    const component = await this.createDynamicComponent(
      objectId, 
      objectType, 
      x, y, width, height, 
      sensorsContainerRef
    )

    if (component) {
      this.dynamicComponents.push({
        id: objectId,
        type: objectType,
        coords: { x, y, width, height },
        component,
        anchorElement: anchor
      })

      if (this.options.showAnchorPlaceholder) {
        this.replaceAnchorWithPlaceholder(anchor)
      }
    }

    return component
  }

  // Создание динамического компонента
  async createDynamicComponent(objectId, objectType, x, y, width, height, sensorsContainerRef) {
    if (!sensorsContainerRef) return null

    const componentLoader = COMPONENT_MAP[objectType]
    if (!componentLoader) {
      console.warn(`⚠️ No component mapped for type: ${objectType}`)
      return null
    }

    const componentModule = await componentLoader()
    const Component = componentModule.default

    const componentContainer = document.createElement('div')
    componentContainer.className = `dynamic-component ${objectType}`
    componentContainer.style.position = 'absolute'
    componentContainer.style.left = `${x}px`
    componentContainer.style.top = `${y}px`
    componentContainer.style.width = `${width}px`
    componentContainer.style.height = `${height}px`
    componentContainer.style.pointerEvents = 'auto'

    sensorsContainerRef.appendChild(componentContainer)

    const AppComponent = {
      render() {
        return h(Component, {
          id: objectId,
          x: 0,
          y: 0,
          w: width,
          h: height
        })
      }
    }

    const app = createApp(AppComponent)
    const instance = app.mount(componentContainer)

    return {
      container: componentContainer,
      instance: instance,
      app: app
    }
  }

  // Замена якоря прозрачным прямоугольником
  replaceAnchorWithPlaceholder(anchor) {
    anchor.style.fill = 'transparent'
    anchor.style.stroke = 'transparent'
    anchor.style.pointerEvents = 'none'
  }

  // Настройка отслеживания изменения размеров
  setupResizeObserver(svgElement, sensorsContainerRef) {
    if (this.resizeObserver) {
      this.resizeObserver.disconnect()
    }

    this.resizeObserver = new ResizeObserver((entries) => {
      this.updateComponentsPosition(svgElement, sensorsContainerRef)
    })

    this.resizeObserver.observe(svgElement)
  }

  // Обновление позиций компонентов при изменении размеров
  updateComponentsPosition(svgElement, sensorsContainerRef) {
    this.dynamicComponents.forEach(comp => {
      // Находим соответствующий anchor
      const anchors = svgElement.querySelectorAll(`[data-id="${comp.id}"]`)
      if (anchors.length === 0) return
      
      const anchor = anchors[0]
      const rect = anchor.getBoundingClientRect()
      const svgRect = svgElement.getBoundingClientRect()
      
      const x = rect.left - svgRect.left
      const y = rect.top - svgRect.top
      const width = rect.width
      const height = rect.height

      // Обновляем позицию контейнера
      if (comp.component.container) {
        comp.component.container.style.left = `${x}px`
        comp.component.container.style.top = `${y}px`
        comp.component.container.style.width = `${width}px`
        comp.component.container.style.height = `${height}px`
      }

      // Обновляем координаты
      comp.coords = { x, y, width, height }
    })
  }

  // Получение компонента управления по типу объекта
  async getControlComponent(objectType) {
    const controlLoader = CONTROL_MAP[objectType]
    if (!controlLoader) {
      console.warn(`⚠️ No control component mapped for type: ${objectType}`)
      return null
    }
    
    const controlModule = await controlLoader()
    return controlModule.default
  }

  // Получение всех типов управления на экране
  getControlTypes() {
    const types = new Set()
    this.dynamicComponents.forEach(comp => {
      if (CONTROL_MAP[comp.type]) {
        types.add(comp.type)
      }
    })
    return Array.from(types)
  }

  // Создание компонентов управления для экрана
  async createControlComponents() {
    const controlTypes = this.getControlTypes()
    
    for (const type of controlTypes) {
      const ControlComponent = await this.getControlComponent(type)
      if (ControlComponent) {
        this.controlComponents.set(type, ControlComponent)
      }
    }
    
    return this.controlComponents
  }

  // Получение типа объекта по ID
  getObjectType(objectId) {
    const component = this.dynamicComponents.find(comp => comp.id === objectId)
    return component ? component.type : null
  }

  // Получение компонента управления для объекта
  getControlComponentForObject(objectId) {
    const objectType = this.getObjectType(objectId)
    return objectType ? this.controlComponents.get(objectType) : null
  }

  // Очистка
  cleanup() {
    if (this.resizeObserver) {
      this.resizeObserver.disconnect()
      this.resizeObserver = null
    }

    this.dynamicComponents.forEach(comp => {
      if (comp.component.app) {
        comp.component.app.unmount()
      }
      if (comp.component.container) {
        comp.component.container.remove()
      }
    })

    this.dynamicComponents = []
    this.controlComponents.clear()
    console.log('🧹 SVG Screen Loader cleaned up')
  }
}

// Создание экземпляра загрузчика
export const svgScreenLoader = new SvgScreenLoader()