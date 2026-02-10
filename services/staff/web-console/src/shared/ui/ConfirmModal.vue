<template>
  <div v-if="open" class="overlay" role="dialog" aria-modal="true" @click.self="emit('cancel')">
    <div class="modal">
      <div class="title">{{ title }}</div>
      <div v-if="message" class="msg">{{ message }}</div>
      <div class="actions">
        <button class="btn" type="button" @click="emit('cancel')">
          {{ cancelText }}
        </button>
        <button class="btn" :class="danger ? 'danger' : 'primary'" type="button" @click="emit('confirm')">
          {{ confirmText }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
withDefaults(
  defineProps<{
    open: boolean;
    title: string;
    message?: string;
    confirmText: string;
    cancelText: string;
    danger?: boolean;
  }>(),
  { danger: true },
);

const emit = defineEmits<{
  (e: "confirm"): void;
  (e: "cancel"): void;
}>();
</script>

<style scoped>
.overlay {
  position: fixed;
  inset: 0;
  background: rgba(17, 24, 39, 0.35);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 18px;
  z-index: 999;
}
.modal {
  width: min(520px, 100%);
  background: rgba(255, 255, 255, 0.92);
  border: 1px solid rgba(17, 24, 39, 0.12);
  border-radius: 14px;
  padding: 14px;
  backdrop-filter: blur(10px);
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.12);
}
.title {
  font-weight: 900;
  letter-spacing: -0.01em;
}
.msg {
  margin-top: 8px;
  opacity: 0.85;
  font-weight: 650;
  font-size: 13px;
  line-height: 1.35;
}
.actions {
  margin-top: 14px;
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}
</style>

