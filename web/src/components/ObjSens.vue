<template>
  <div class="obj-sensor" :style="sensorStyle" @click="handleClick">
    <svg :viewBox="viewBox" :width="props.w" :height="props.h" preserveAspectRatio="xMidYMid meet">
      <g id="mainGroup" transform="translate(0, 0)">
        
        <foreignObject x="5" y="0" width="90" height="27">
          <div xmlns="http://www.w3.org/1999/xhtml" class="text-wrap" 
              :style="formattedTextData.textStyle" 
              :title="props.id">
            {{ formattedTextData.formattedText }}
          </div>
        </foreignObject>
        
        <rect x="35" y="30" width="30" height="30" stroke-width="2" rx="3"
          :fill="computedState.stateColor"     
          :stroke="computedState.maskColor" 
          :class="{ blinking: computedState.ackBlink }" />
        
        <polyline points="15,5 10,15 20,15 15,25"
            fill="none" stroke="#000000" stroke-width="2" transform="translate(35,30)" />
      
        <text x="50" y="65" dominant-baseline="hanging" text-anchor="middle" fill="#000000" font-size="12" font-weight="normal">
          {{ computedState.inputValue }}
        </text>
        
        <g transform="translate(65,60)" :class="{ 'fade-transition': true }" 
           :opacity="computedState.imitVisible ? 1 : 0">
          <circle r="7" cx="0" cy="0" fill="#FF00FF" stroke="#000" stroke-width="1"/>
          <text x="0" y="0" dominant-baseline="middle" text-anchor="middle" fill="#000000" font-size="12" font-weight="bold">и</text>
        </g>
      </g>
    </svg>
  </div>
</template>

<script>
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useObjectsStore } from '@/stores/objects'
import { createFormattedText } from '@/utils/textFormatter' // Импорт утилиты
import '@/assets/styles/obj-base.css'
import { CTRL_COLORS } from '@/constants/ctrl-colors'
import { CTRL_TEXT } from '@/constants/ctrl-text'

export default {
  name: 'ObjSens',
  props: {
    id: String,
    x: Number,
    y: Number,
    w: Number,
    h: Number
  },
  setup(props) {
    const objectsStore = useObjectsStore()
    
    const viewBox = ref('0 0 100 100')

    // Подписываемся при монтировании компонента
    onMounted(() => {
      objectsStore.subscribe(props.id)
    })

    // Отписываемся при уничтожении компонента
    onUnmounted(() => {
      objectsStore.unsubscribe(props.id)
    })
    
    // следим за обновлением статуса
    const computedState = computed(() => {
      const objData = objectsStore.objects[props.id]
      if (!objData) {
        return {
          stateColor: CTRL_COLORS.stateUnknown,
          maskColor: CTRL_COLORS.stateMaskOff,
          inputValue: CTRL_TEXT.defaults.noValue,
          ackBlink: false,
          imitVisible: true
        }
      }

      const objVue = objData.objVue || {}
     
      return {
        stateColor: objVue.stateColor || CTRL_COLORS.stateUnknown,
        maskColor: objVue.mask ? CTRL_COLORS.stateMaskOn : CTRL_COLORS.stateMaskOff,
        inputValue: objVue.inputValue || CTRL_TEXT.defaults.noValue,
        ackBlink: objVue.ack,
        imitVisible: objVue.imit
      }
    })

    // следим за обновлением имени тэга
    let cachedResult = null
    let lastTagName = null

    const formattedTextData = computed(() => {
      const objData = objectsStore.objects[props.id]
      
      if (!objData || !objData.objInfo) {
        return createFormattedText(CTRL_TEXT.defaults.noTag, {
          maxLength: 12,
          fontSize: '12px',
          color: '#000000',
          textAlign: 'center',
          verticalAlign: 'bottom',
          fontWeight: 'bold'
        })
      }

      const objInfo = objData.objInfo || {}
      const tagName = objInfo.info?.tag || CTRL_TEXT.defaults.noTag
      
      if (cachedResult && lastTagName === tagName) {
        return cachedResult
      }
      
      cachedResult = createFormattedText(tagName, {
        maxLength: 12,
        fontSize: '12px',
        color: '#000000',
        textAlign: 'center',
        verticalAlign: 'bottom',
        fontWeight: 'bold'
      })
      
      lastTagName = tagName
      return cachedResult
    })

    // следим за обновлением расположения элемента на мнемосхеме
    const sensorStyle = computed(() => ({
      position: 'absolute',
      left: `${props.x}px`,
      top: `${props.y}px`,
      width: `${props.w}px`,
      height: `${props.h}px`,
      cursor: 'pointer',
      zIndex: 1000
    }))

    
    // следим за вызовом окна управления
    const handleClick = () => {
      objectsStore.openControl(props.id)
    }

    return {
      props,
      computedState,
      sensorStyle,
      handleClick,
      viewBox,
      formattedTextData
    }
  }
}
</script>

<style scoped>

</style>