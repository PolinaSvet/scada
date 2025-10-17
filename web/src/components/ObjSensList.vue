<template>
  <div 
    class="virtual-container" 
    ref="containerRef"
    @scroll="handleScroll"
  >
    <div 
      class="virtual-content" 
      :style="{ height: totalHeight + 'px' }"
    >
      <div
        v-for="sensor in visibleSensors"
        :key="sensor.id"
        class="virtual-item"
        :style="{ transform: `translateY(${sensor.offset}px)` }"
      >
        <ObjSens
          :id="sensor.id"
          :x="sensor.x"
          :y="sensor.y"
        />
      </div>
    </div>
  </div>
</template>

<script>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useObjectsStore } from '@/stores/objects'
import ObjSens from './ObjSens.vue'

export default {
  name: 'ObjSensList',
  components: {
    ObjSens
  },
  props: {
    sensors: {
      type: Array,
      required: true
    }
  },
  setup(props) {
    const objectsStore = useObjectsStore()
    const containerRef = ref(null)
    const scrollTop = ref(0)
    const containerHeight = ref(600)
    const itemHeight = 100 // Высота одного элемента (80px + отступы)

    // Вычисляем видимые сенсоры
    const visibleSensors = computed(() => {
      const startIndex = Math.max(0, Math.floor(scrollTop.value / itemHeight) - 3) // Буфер сверху
      const visibleCount = Math.ceil(containerHeight.value / itemHeight) + 6 // Буфер снизу
      const endIndex = Math.min(startIndex + visibleCount, props.sensors.length)

      return props.sensors
        .slice(startIndex, endIndex)
        .map((sensor, index) => ({
          ...sensor,
          offset: (startIndex + index) * itemHeight
        }))
    })

    const totalHeight = computed(() => props.sensors.length * itemHeight)

    const handleScroll = (event) => {
      scrollTop.value = event.target.scrollTop
    }

    const updateContainerSize = () => {
      if (containerRef.value) {
        containerHeight.value = containerRef.value.clientHeight
      }
    }

    // Ресайз обработчик
    const handleResize = () => {
      updateContainerSize()
    }

    onMounted(() => {
      updateContainerSize()
      window.addEventListener('resize', handleResize)
    })

    onUnmounted(() => {
      window.removeEventListener('resize', handleResize)
    })

    // Обновляем размер при изменении контейнера
    watch(containerRef, () => {
      updateContainerSize()
    })

    return {
      containerRef,
      visibleSensors,
      totalHeight,
      handleScroll
    }
  }
}
</script>

<style scoped>
.virtual-container {
  height: 600px;
  overflow-y: auto;
  border: 2px solid #ccc;
  background: #f0f0f0;
  position: relative;
}

.virtual-content {
  position: relative;
}

.virtual-item {
  position: absolute;
  left: 0;
  right: 0;
  height: 100px;
  transition: transform 0.1s ease;
}

/* Стили для скроллбара */
.virtual-container::-webkit-scrollbar {
  width: 8px;
}

.virtual-container::-webkit-scrollbar-track {
  background: #f1f1f1;
}

.virtual-container::-webkit-scrollbar-thumb {
  background: #c1c1c1;
  border-radius: 4px;
}

.virtual-container::-webkit-scrollbar-thumb:hover {
  background: #a8a8a8;
}
</style>