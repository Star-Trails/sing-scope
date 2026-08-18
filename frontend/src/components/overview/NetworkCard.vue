<template>
  <div class="network-card base-container w-full p-4">
    <div class="network-card-grid grid grid-cols-1 gap-3">
      <ConnectionStatus />
    </div>
  </div>
</template>

<script setup lang="ts">
import ConnectionStatus from '@/components/overview/ConnectionStatus.vue'
</script>

<style scoped>
.network-card {
  container-type: inline-size;
}
</style>
