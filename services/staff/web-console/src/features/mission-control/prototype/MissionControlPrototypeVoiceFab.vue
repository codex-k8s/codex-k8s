<template>
  <button type="button" class="mission-voice-fab" aria-label="Голосовая команда" @click="$emit('click')">
    <VIcon icon="mdi-microphone" size="26" />
  </button>
</template>

<script setup lang="ts">
defineEmits<{
  (event: "click"): void;
}>();
</script>

<style scoped>
.mission-voice-fab {
  position: fixed;
  right: 28px;
  bottom: 26px;
  z-index: 40;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 62px;
  height: 62px;
  border-radius: 50%;
  border: 1px solid rgba(255, 137, 110, 0.56);
  background: linear-gradient(135deg, rgb(255, 127, 102), rgb(255, 103, 88));
  color: white;
  box-shadow: 0 22px 44px rgba(255, 107, 84, 0.34);
  backdrop-filter: blur(12px);
}

.mission-voice-fab:hover {
  transform: translateY(-1px);
}

@media (max-width: 720px) {
  .mission-voice-fab {
    bottom: 18px;
    right: 18px;
  }
}
</style>
