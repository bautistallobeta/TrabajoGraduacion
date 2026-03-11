<script setup>
import AppNavbar from './AppNavbar.vue'
</script>

<template>
  <div class="app-layout">
    <AppNavbar />
    <main class="main-content">
      <div class="content-wrapper">
        <router-view />
      </div>
    </main>
  </div>
</template>

<style scoped>
.app-layout {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
}

.main-content {
  flex: 1;
  padding: 2rem 0;
}

.content-wrapper {
  max-width: 1400px;
  margin: 0 auto;
  padding: 0 1.5rem;
}
</style>
