<template>
  <div class="ctrl-confirmation-overlay" v-if="isOpen">
    <div class="ctrl-confirmation-dialog">
      <p class="ctrl-confirmation-message">{{ message }}</p>
      <p class="ctrl-confirmation-question">{{ messageQuestion }}</p>
      <div class="ctrl-confirmation-buttons">
        <button class="ctrl-confirmation-btn confirm" @click="handleYes">ДА</button>
        <button class="ctrl-confirmation-btn" @click="handleNo">НЕТ</button>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  props: {
    isOpen: {
      type: Boolean,
      required: true
    },
    message: {
      type: String,
      required: true
    },
    messageQuestion: {
      type: String,
      required: true
    }
  },
  methods: {
    handleYes() {
      this.$emit('confirm')
    },
    handleNo() {
      this.$emit('close')
    }
  }
}
</script>

<style scoped>
/* Стили теперь в ctrl-base.css */
</style>