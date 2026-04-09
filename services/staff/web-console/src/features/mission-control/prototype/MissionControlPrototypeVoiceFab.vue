<template>
  <button type="button" class="mission-voice-fab" @click="$emit('click')">
    <VIcon icon="mdi-microphone" size="22" />
    <span>Голосовая команда</span>
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
  left: 50%;
  bottom: 26px;
  z-index: 40;
  transform: translateX(-50%);
  display: inline-flex;
  align-items: center;
  gap: 10px;
  padding: 14px 20px;
  border-radius: 999px;
  border: 1px solid rgba(255, 137, 110, 0.44);
  background: linear-gradient(135deg, rgb(255, 127, 102), rgb(255, 103, 88));
  color: white;
  font-size: 0.96rem;
  font-weight: 700;
  box-shadow: 0 22px 44px rgba(255, 107, 84, 0.34);
  backdrop-filter: blur(12px);
}

.mission-voice-fab:hover {
  transform: translateX(-50%) translateY(-1px);
}

@media (max-width: 720px) {
  .mission-voice-fab {
    bottom: 18px;
    width: calc(100% - 32px);
    justify-content: center;
  }
}
</style>
