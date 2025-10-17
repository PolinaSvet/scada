<template>
  <div id="app">
    <NavbarTop />
    <NavbarLeft />
    <MainContainer />
    <NavbarBottom />
    
    <!-- Глобальные компоненты управления -->
    <!-- <CtrlConfirmation /> -->
  </div>
</template>

<script>
import NavbarTop from '@/components/layout/NavbarTop.vue'
import NavbarLeft from '@/components/layout/NavbarLeft.vue'
import MainContainer from '@/components/layout/MainContainer.vue'
import NavbarBottom from '@/components/layout/NavbarBottom.vue'
//import CtrlConfirmation from '@/components/CtrlConfirmation.vue'

export default {
  name: 'App',
  components: {
    NavbarTop,
    NavbarLeft,
    MainContainer,
    NavbarBottom,
    //CtrlConfirmation
  },
  setup() {
    console.log('SCADA System initialized')
    
    return {
    }
  }
}
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
  background: #1a1a1a;
  color: white;
  overflow: hidden;
}

#app {
  width: 100vw;
  height: 100vh;
  position: relative;
}
</style>