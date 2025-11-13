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
      console.log('🔄 Starting SVG screen load...')
      containerRef.innerHTML = svgRaw
      const screenObjects = await this.parseAnchorsAndCreateComponents(containerRef, sensorsContainerRef)
      
      // Ждем полной загрузки всех компонентов
      await this.waitForAllComponents()
      
      // Принудительно обновляем позиции и видимость
      setTimeout(() => {
        const svgElement = containerRef.querySelector('svg')
        if (svgElement) {
          this.updateComponentsPosition(svgElement, sensorsContainerRef)
          this.forceComponentsVisibility()
        }
      }, 200)
      
      // Проверяем невидимые компоненты и пересоздаем их при необходимости
      setTimeout(() => {
        this.debugComponents()
        const invisibleCount = this.dynamicComponents.filter(comp => 
          !comp.component.container || 
          comp.component.container.offsetWidth === 0 || 
          comp.component.container.offsetHeight === 0
        ).length
        
        if (invisibleCount > 0) {
          console.log(`🔄 ${invisibleCount} components are invisible, recreating...`)
          this.recreateInvisibleComponents(containerRef, sensorsContainerRef)
        }
      }, 500)
      
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

        // Ждем полного рендера SVG
        await this.waitForSvgRender(svgElement)
        
        this.setupResizeObserver(svgElement, sensorsContainerRef)

        const anchors = svgElement.querySelectorAll('[data-type]')
        console.log(`🔍 Found ${anchors.length} anchors in SVG`)

        const screenObjects = []

        // Создаем компоненты для ВСЕХ anchor элементов
        for (const anchor of anchors) {
          const objectId = anchor.getAttribute('data-id')
          const objectType = anchor.getAttribute('data-type')
          const anchorId = anchor.getAttribute('id')
          
          if (!objectId || !objectType) {
            console.warn('⚠️ Anchor missing data-id or data-type:', anchor)
            continue
          }

          try {
            // Добавляем небольшую задержку между созданием компонентов
            await new Promise(resolve => setTimeout(resolve, 10))
            
            await this.createComponentForAnchor(
              objectId, 
              objectType, 
              anchorId,
              anchor, 
              svgElement, 
              sensorsContainerRef
            )
            screenObjects.push({
              objectId,
              anchorId,
              objectType
            })
          } catch (error) {
            console.error(`❌ Error creating component for ${objectId}:`, error)
          }
        }

        await this.createControlComponents()

        console.log(`✅ Created ${this.dynamicComponents.length} dynamic components`)
        resolve(screenObjects)
      }, 150)
    })
  }

  // Ждем полного рендера SVG
  waitForSvgRender(svgElement) {
    return new Promise((resolve) => {
      const checkRender = () => {
        const rect = svgElement.getBoundingClientRect()
        if (rect.width > 0 && rect.height > 0) {
          console.log('✅ SVG fully rendered:', rect)
          resolve()
        } else {
          console.log('⏳ Waiting for SVG render...')
          setTimeout(checkRender, 50)
        }
      }
      checkRender()
    })
  }

  // Создание компонента для якоря
  async createComponentForAnchor(objectId, objectType, anchorId, anchor, svgElement, sensorsContainerRef) {
    // Ждем пока anchor будет иметь валидные размеры
    await this.waitForValidAnchorSize(anchor)
    
    const rect = anchor.getBoundingClientRect()
    const svgRect = svgElement.getBoundingClientRect()
    
    // Относительные координаты внутри sensors-area
    const x = rect.left - svgRect.left
    const y = rect.top - svgRect.top
    const width = Math.max(rect.width, 1)
    const height = Math.max(rect.height, 1)

    // Проверяем валидность координат
    if (isNaN(x) || isNaN(y) || width <= 0 || height <= 0) {
      console.warn(`⚠️ Invalid coordinates for anchor ${anchorId}:`, { x, y, width, height })
    }

    const componentId = anchorId || `${objectId}_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`

    console.log(`🎯 Creating ${objectType}: ${objectId} at [${x}, ${y}] ${width}x${height}`)

    const component = await this.createDynamicComponent(
      objectId, 
      objectType, 
      componentId,
      x, y, width, height, 
      sensorsContainerRef
    )

    if (component) {
      this.dynamicComponents.push({
        id: componentId,
        objectId: objectId,
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

  // Ждем пока anchor получит валидные размеры
  waitForValidAnchorSize(anchor) {
    return new Promise((resolve) => {
      const checkSize = () => {
        const rect = anchor.getBoundingClientRect()
        if (rect.width > 0 && rect.height > 0) {
          resolve()
        } else {
          console.log('⏳ Waiting for anchor size...', anchor.getAttribute('id'))
          setTimeout(checkSize, 30)
        }
      }
      checkSize()
    })
  }

  // Создание динамического компонента
  async createDynamicComponent(objectId, objectType, componentId, x, y, width, height, sensorsContainerRef) {
    if (!sensorsContainerRef) {
      console.error('❌ No sensors container ref')
      return null
    }

    const componentLoader = COMPONENT_MAP[objectType]
    if (!componentLoader) {
      console.warn(`⚠️ No component mapped for type: ${objectType}`)
      return null
    }

    try {
      const componentModule = await componentLoader()
      const Component = componentModule.default

      const componentContainer = document.createElement('div')
      componentContainer.className = `dynamic-component ${objectType}`
      componentContainer.dataset.componentId = componentId
      componentContainer.dataset.objectId = objectId
      componentContainer.style.position = 'absolute'
      componentContainer.style.left = `${x}px`
      componentContainer.style.top = `${y}px`
      componentContainer.style.width = `${width}px`
      componentContainer.style.height = `${height}px`
      componentContainer.style.pointerEvents = 'auto'
      componentContainer.style.zIndex = '0'
      componentContainer.style.background = 'transparent'
      componentContainer.style.display = 'block'
      componentContainer.style.visibility = 'visible'
      componentContainer.style.opacity = '1'
      componentContainer.style.overflow = 'visible'

      // Добавляем контейнер в DOM ДО монтирования Vue компонента
      sensorsContainerRef.appendChild(componentContainer)

      // Принудительный reflow
      void componentContainer.offsetHeight

      const AppComponent = {
        mounted() {
          console.log(`✅ Vue component mounted: ${componentId}`)
          // Дополнительная проверка после монтирования
          setTimeout(() => {
            const isVisible = componentContainer.offsetWidth > 0 && 
                             componentContainer.offsetHeight > 0 &&
                             componentContainer.style.display !== 'none'
            
            if (!isVisible) {
              console.warn(`⚠️ Component ${componentId} might be invisible after mount`)
              // Принудительно исправляем стили
              componentContainer.style.display = 'block'
              componentContainer.style.visibility = 'visible'
              componentContainer.style.opacity = '1'
            }
          }, 50)
        },
        render() {
          return h(Component, {
            id: objectId,
            componentId: componentId,
            x: 0,
            y: 0,
            w: width,
            h: height,
            key: componentId
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
    } catch (error) {
      console.error(`❌ Error creating dynamic component ${componentId}:`, error)
      return null
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
    if (!svgElement || !sensorsContainerRef) return
    
    const svgRect = svgElement.getBoundingClientRect()
    
    console.log(`🔄 Updating positions for ${this.dynamicComponents.length} components`)
    
    this.dynamicComponents.forEach((comp, index) => {
      if (!comp.anchorElement) {
        console.warn(`❌ No anchor element for component ${comp.id}`)
        return
      }
      
      try {
        const rect = comp.anchorElement.getBoundingClientRect()
        
        const x = rect.left - svgRect.left
        const y = rect.top - svgRect.top
        const width = Math.max(rect.width, 1)
        const height = Math.max(rect.height, 1)
        
        // Проверяем валидность координат
        if (isNaN(x) || isNaN(y) || width <= 0 || height <= 0) {
          console.warn(`⚠️ Invalid coordinates for component ${comp.id}:`, { x, y, width, height })
          return
        }
        
        if (comp.component.container) {
          comp.component.container.style.left = `${x}px`
          comp.component.container.style.top = `${y}px`
          comp.component.container.style.width = `${width}px`
          comp.component.container.style.height = `${height}px`
          
          // Принудительно показываем элемент
          comp.component.container.style.display = 'block'
          comp.component.container.style.visibility = 'visible'
          comp.component.container.style.opacity = '1'
        }
        
        comp.coords = { x, y, width, height }
        
      } catch (error) {
        console.error(`❌ Error updating position for component ${comp.id}:`, error)
      }
    })
  }

  // Ожидание загрузки всех компонентов
  waitForAllComponents() {
    return new Promise((resolve) => {
      const checkComponents = () => {
        const allLoaded = this.dynamicComponents.every(comp => {
          return comp.component?.container && comp.component.instance
        })
        
        if (allLoaded) {
          console.log('✅ All components loaded')
          resolve(true)
        } else {
          setTimeout(checkComponents, 50)
        }
      }
      
      checkComponents()
    })
  }

  // Принудительное обновление видимости компонентов
  forceComponentsVisibility() {
    console.log('🔧 Forcing components visibility...')
    
    this.dynamicComponents.forEach((comp, index) => {
      if (comp.component.container) {
        const container = comp.component.container
        
        // Принудительно устанавливаем стили
        container.style.display = 'block'
        container.style.visibility = 'visible'
        container.style.opacity = '1'
        container.style.zIndex = (1000 + index).toString()
        container.style.background = 'transparent'
        
        // Принудительный reflow
        void container.offsetHeight
        
        console.log(`✅ Component ${comp.id} styles:`, {
          display: container.style.display,
          visibility: container.style.visibility,
          opacity: container.style.opacity,
          zIndex: container.style.zIndex,
          coords: comp.coords
        })
      }
    })
    
    // Дополнительная проверка через небольшой интервал
    setTimeout(() => {
      const visibleCount = this.dynamicComponents.filter(comp => {
        return comp.component.container && 
               comp.component.container.offsetWidth > 0 && 
               comp.component.container.offsetHeight > 0
      }).length
      
      console.log(`📊 Visibility report: ${visibleCount}/${this.dynamicComponents.length} components visible`)
    }, 200)
  }

  // Пересоздание невидимых компонентов
  async recreateInvisibleComponents(containerRef, sensorsContainerRef) {
    const invisibleComponents = this.dynamicComponents.filter(comp => {
      return !comp.component.container || 
             comp.component.container.offsetWidth === 0 || 
             comp.component.container.offsetHeight === 0
    })
    
    console.log(`🔄 Recreating ${invisibleComponents.length} invisible components`)
    
    for (const comp of invisibleComponents) {
      // Удаляем старый компонент
      if (comp.component.app) {
        comp.component.app.unmount()
      }
      if (comp.component.container) {
        comp.component.container.remove()
      }
      
      // Удаляем из массива
      this.dynamicComponents = this.dynamicComponents.filter(c => c.id !== comp.id)
      
      // Пересоздаем компонент
      await new Promise(resolve => setTimeout(resolve, 50))
      
      try {
        await this.createComponentForAnchor(
          comp.objectId,
          comp.type,
          comp.id,
          comp.anchorElement,
          containerRef.querySelector('svg'),
          sensorsContainerRef
        )
      } catch (error) {
        console.error(`❌ Error recreating component ${comp.id}:`, error)
      }
    }
  }

  // Отладка компонентов
  debugComponents() {
    console.log('=== COMPONENTS DEBUG ===')
    console.log(`Total components in memory: ${this.dynamicComponents.length}`)
    
    const visibleComponents = this.dynamicComponents.filter(comp => {
      const container = comp.component?.container
      if (!container) return false
      
      const style = window.getComputedStyle(container)
      const isVisible = style.display !== 'none' && 
                       style.visibility !== 'hidden' && 
                       style.opacity !== '0' &&
                       container.offsetWidth > 0 &&
                       container.offsetHeight > 0
      
      return isVisible
    })
    
    console.log(`Visible components: ${visibleComponents.length}`)
    
    this.dynamicComponents.forEach((comp, index) => {
      const container = comp.component?.container
      const style = container ? window.getComputedStyle(container) : null
      
      console.log(`Component ${index}:`, {
        id: comp.id,
        objectId: comp.objectId,
        type: comp.type,
        containerExists: !!container,
        display: style?.display,
        visibility: style?.visibility,
        opacity: style?.opacity,
        width: container?.offsetWidth,
        height: container?.offsetHeight,
        position: comp.coords
      })
    })
  }

  // Получение всех компонентов по data-id
  getComponentsByObjectId(objectId) {
    return this.dynamicComponents.filter(comp => comp.objectId === objectId)
  }

  // Получение компонента по его уникальному идентификатору
  getComponentById(componentId) {
    return this.dynamicComponents.find(comp => comp.id === componentId)
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

  // Получение типа объекта по data-id
  getObjectType(objectId) {
    const components = this.getComponentsByObjectId(objectId)
    return components.length > 0 ? components[0].type : null
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