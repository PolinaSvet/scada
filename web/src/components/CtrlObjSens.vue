<template>
  <div class="ctrl-dialog-overlay" v-if="isOpen">
    <div class="ctrl-dialog">
      <div class="ctrl-dialog-header">
        <span class="ctrl-dialog-title">{{ computedData.fullName }}</span>
        <button class="ctrl-close-button" @click="closeDialog">×</button>
      </div>
      
      <div class="ctrl-dialog-content">
        <!-- Область объекта -->
        <div class="ctrl-obj-section">
          <div class="ctrl-obj-grid">
            <div class="ctrl-obj-sens-container">
              <!-- Исправляем передачу props -->
              <ObjSens :id="id" :x="10" :y="5" :w="100" :h="100" />
            </div>
            <div class="ctrl-obj-status-container">
              <div class="ctrl-status-text" :style="{ backgroundColor: computedData.stateColor }">
                {{ computedData.stateTxt }}
              </div>
              <div class="ctrl-value-text">
                {{ computedData.inputValue }}
              </div>
            </div>
          </div>
        </div>
        
        <div class="ctrl-tab-titles">
          <div
            v-for="tab in tabs"
            :key="tab.name"
            :class="['ctrl-tab-title', { active: tab.name === activeTab }]"
            @click="activateTab(tab.name)"
          >
            {{ tab.label }}
          </div>
        </div>

        <div class="ctrl-tab-content">
          <!-- Вкладка управления -->
          <div v-if="activeTab === 'tab1'" class="tab-panel">
            <div class="ctrl-grid-layout ctrl-grid-cols-2">
              <!-- Секция имитации -->
              <div 
                class="ctrl-grid-section" :class="{ active: computedData.imitColor === CTRL_COLORS.stateActive }" 
              >
                <div class="ctrl-grid-section-header">
                  {{ CTRL_TEXT.sections.imitation }}
                </div>
                <div class="ctrl-grid-section-buttons">
                  <button class="ctrl-grid-btn" @click="openConfirmation('imit_on')">
                    {{ computedData.stateOnTxt }}
                  </button>
                  <button class="ctrl-grid-btn" @click="openConfirmation('imit_off')">
                    {{ computedData.stateOffTxt }}
                  </button>
                  <button class="ctrl-grid-btn" @click="openConfirmation('imit_clear')">
                    {{ CTRL_TEXT.buttons.imitateClear }}
                  </button>
                </div>
              </div>
              
              <!-- Секция маскирования -->
              <div 
                class="ctrl-grid-section" 
                :class="{ active: computedData.maskColor === CTRL_COLORS.stateActive }"
              >
                <div class="ctrl-grid-section-header">
                  {{ CTRL_TEXT.sections.masking }}
                </div>
                <div class="ctrl-grid-section-buttons">
                  <button class="ctrl-grid-btn" @click="openConfirmation('mask_on')">
                    {{ CTRL_TEXT.buttons.maskOn }}
                  </button>
                  <button class="ctrl-grid-btn" @click="openConfirmation('mask_off')">
                    {{ CTRL_TEXT.buttons.maskOff }}
                  </button>
                </div>
              </div>
              
              <!-- Секция квитирования -->
              <div 
                class="ctrl-grid-section-full" 
                :class="{ active: computedData.ackColor === CTRL_COLORS.stateActive }"
              >
                <div class="ctrl-grid-section-header">
                  {{ CTRL_TEXT.sections.acknowledgment }}
                </div>
                <div class="ctrl-grid-section-buttons">
                  <button class="ctrl-grid-btn-full" @click="openConfirmation('ack')">
                    {{ CTRL_TEXT.buttons.acknowledge }}
                  </button>
                </div>
              </div>
            </div>
          </div>
          
          <!-- Вкладка неисправностей -->
          <div v-if="activeTab === 'tab2'" class="tab-panel">
            <div class="ctrl-error-grid">
              <div 
                v-for="(error, index) in 16" 
                :key="index"
                class="ctrl-error-item" 
                :style="{ backgroundColor: computedData.errorColors[index] }"
              >
                {{ computedData.errorTexts[index] }}
              </div>
            </div>
          </div>
        </div>

        <CtrlConfirmation
          v-if="isConfirmationOpen"
          :isOpen="isConfirmationOpen"
          :message="confirmationMessage"
          :messageQuestion="confirmationMessageQuestion"
          @confirm="handleConfirm"
          @close="closeConfirmation"
        />
      </div>
      
      <div class="ctrl-dialog-footer">
        <span>{{ computedData.uso }}</span>
      </div>
    </div>
  </div>
</template>

<script>
import { computed, ref, onMounted, onUnmounted } from 'vue'
import { useObjectsStore } from '@/stores/objects'
import ObjSens from './ObjSens.vue'
import CtrlConfirmation from './CtrlConfirmation.vue'
import { CTRL_COLORS } from '@/constants/ctrl-colors'
import { CTRL_TEXT } from '@/constants/ctrl-text'

import '@/assets/styles/ctrl-base.css'

export default {
  name: 'CtrlObjSens',
  components: {
    ObjSens,
    CtrlConfirmation
  },
  props: {
    isOpen: {
      type: Boolean,
      required: true
    },
    id: {
      type: String,
      required: true
    }
  },
  setup(props, { emit }) {
    const objectsStore = useObjectsStore()

    // Подписываемся при открытии окна
    onMounted(() => {
      objectsStore.subscribe(props.id)
    })

    // Отписываемся при закрытии окна
    onUnmounted(() => {
      objectsStore.unsubscribe(props.id)
    })

    const activeTab = ref('tab1')
    const isConfirmationOpen = ref(false)
    const confirmationMessage = ref('')
    const confirmationMessageQuestion = ref('')
    const pendingCommand = ref(null)

    const tabs = [
      { name: 'tab1', label: CTRL_TEXT.tabs.control },
      { name: 'tab2', label: CTRL_TEXT.tabs.errors }
    ]

    // Следим за изменениями объекта
    const computedData = computed(() => {
      const objData = objectsStore.objects[props.id]
      if (!objData) {
        return getDefaultData()
      }

      const objInfo = objData.objInfo || {}
      const objVue = objData.objVue || {}
      const stateInfo = objInfo.state || {}

      const error = objVue.error || 0
      const errorColors = Array.from({ length: 16 }, (_, index) => 
        (error >> index) & 1 ? CTRL_COLORS.errorActive : CTRL_COLORS.errorInactive
      )

      return {
        fullName: `${objInfo.info?.name || ''} (${objInfo.info?.tag || ''})`,
        stateColor: objVue.stateColor || CTRL_COLORS.stateUnknown,
        stateTxt: objVue.stateTxt || CTRL_TEXT.defaults.noData,
        inputValue: objVue.inputValue || '',
        stateOnTxt: stateInfo.txtOn || CTRL_TEXT.defaults.stateOn,
        stateOffTxt: stateInfo.txtOff || CTRL_TEXT.defaults.stateOff,
        imitColor: objVue.imit ? CTRL_COLORS.stateActive : CTRL_COLORS.stateInactive,
        maskColor: objVue.mask ? CTRL_COLORS.stateActive : CTRL_COLORS.stateInactive,
        ackColor: objVue.ack ? CTRL_COLORS.stateActive : CTRL_COLORS.stateInactive,
        uso: objInfo.uso?.txt || '',
        errorColors,
        errorTexts: 
          objData?.objInfo?.errType === 1 ? CTRL_TEXT.errors_type_1 :
          objData?.objInfo?.errType === 2 ? CTRL_TEXT.errors_type_2 :
          CTRL_TEXT.errors_type_0 
      }
    })

    // Обновляем сообщения подтверждения при изменении данных
    const commandMessages = computed(() => ({
      imit_on: `${CTRL_TEXT.confirmations.imitateOn} <${computedData.value.stateOnTxt}> СИГНАЛА?`,
      imit_off: `${CTRL_TEXT.confirmations.imitateOff} <${computedData.value.stateOffTxt}> СИГНАЛА?`,
      imit_clear: `${CTRL_TEXT.confirmations.imitateClear}?`,
      mask_on: `${CTRL_TEXT.confirmations.maskOn}?`,
      mask_off: `${CTRL_TEXT.confirmations.maskOff}?`,
      ack: `${CTRL_TEXT.confirmations.acknowledge}?`
    }))

    const getDefaultData = () => ({
      fullName: CTRL_TEXT.defaults.objectNotFound,
      stateColor: CTRL_COLORS.stateUnknown,
      stateTxt: CTRL_TEXT.defaults.noData,
      inputValue: '',
      stateOnTxt: CTRL_TEXT.defaults.stateOn,
      stateOffTxt: CTRL_TEXT.defaults.stateOff,
      imitColor: CTRL_COLORS.stateInactive,
      maskColor: CTRL_COLORS.stateInactive,
      ackColor: CTRL_COLORS.stateInactive,
      uso: '',
      errorColors: Array(16).fill(CTRL_COLORS.errorInactive),
      errorTexts: CTRL_TEXT.errors_type_0
    })

    const activateTab = (tabName) => {
      activeTab.value = tabName
    }

    const closeDialog = () => {
      emit('close')
    }

    const openConfirmation = (commandType) => {
      const objData = objectsStore.objects[props.id]
      if (!objData?.objInfo?.ctrlEnable) {
        console.log('❌ Управление отключено для объекта:', props.id)
        return
      }

      confirmationMessage.value = computedData.value.fullName
      confirmationMessageQuestion.value = commandMessages.value[commandType] || 'Подтвердите действие'
      //pendingCommand.value = commandType
  
      const commandMap = {
        'imit_on': 1,
        'imit_off': 2,
        'imit_clear':3,
        'mask_on': 4,
        'mask_off': 5,
        'set_settings': 6,
        'imit_atten': 7,
        'ack':15
      }

      const objInfo = objData.objInfo || {}
      const stateInfo = objInfo.cmd || {}
      const codeCmd = commandMap[commandType] || 0
      const idObj = objInfo.id || 0
      
      const command = (codeCmd << 12) | idObj
      
      pendingCommand.value = {
        cmdTag: stateInfo, 
        cmdValue: command, 
        cmdType: commandType,
        cmdMess:  confirmationMessage.value,
        cmdMessQuestion: confirmationMessageQuestion.value,
        objId: props.id,
        objType: 'sensor'
      }
      
      isConfirmationOpen.value = true
    }

    const handleConfirm = () => {
      if (pendingCommand.value) {
        objectsStore.sendCommand(props.id,'sendCommand', pendingCommand.value)
        //alert(JSON.stringify(pendingCommand.value, null, 2))
      }
      closeConfirmation()
    }

    const closeConfirmation = () => {
      isConfirmationOpen.value = false
      confirmationMessage.value = ''
      confirmationMessageQuestion.value = ''
      pendingCommand.value = null
    }

    return {
      activeTab,
      tabs,
      isConfirmationOpen,
      confirmationMessage,
      confirmationMessageQuestion,
      computedData,
      CTRL_COLORS,
      CTRL_TEXT,
      activateTab,
      closeDialog,
      openConfirmation,
      handleConfirm,
      closeConfirmation
    }
  }
}
</script>

<style scoped>
.ctrl-dialog {
  width: 500px !important;
  height: 580px !important;
}
</style>