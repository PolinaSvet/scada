<template>
  <div class="ctrl-dialog-overlay" v-if="isOpen">
    <div class="ctrl-dialog" :style="{ width: '500px', height: '550px' }">
      <div class="ctrl-dialog-header">
        <span class="ctrl-dialog-title">Управление устройством</span>
        <button class="ctrl-close-button" @click="closeDialog">×</button>
      </div>
      
      <div class="ctrl-dialog-content">
        <!-- Своя уникальная область -->
        <div class="ctrl-obj-section ctrl-mb-3">
          <!-- Уникальный контент -->
        </div>
        
        <!-- Вкладки -->
        <div class="ctrl-tab-titles">
          <div class="ctrl-tab-title active">Основные</div>
          <div class="ctrl-tab-title">Настройки</div>
          <div class="ctrl-tab-title">Диагностика</div>
        </div>

        <div class="ctrl-tab-content">
          <!-- Своя сетка -->
          <div class="ctrl-grid-layout ctrl-grid-cols-3">
            <div class="ctrl-grid-section">
              <div class="ctrl-grid-section-header">Секция 1</div>
              <div class="ctrl-grid-section-buttons">
                <button class="ctrl-grid-btn">Кнопка 1</button>
                <button class="ctrl-grid-btn">Кнопка 2</button>
              </div>
            </div>
            <!-- ... -->
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import '@/assets/styles/ctrl-base.css'

export default {
  name: 'CtrlAnotherDevice',
  // ... логика компонента
}
</script>