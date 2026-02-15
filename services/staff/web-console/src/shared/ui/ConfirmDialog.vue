<template>
  <VDialog :model-value="modelValue" max-width="520" @update:model-value="emit('update:modelValue', $event)">
    <VCard>
      <VCardTitle class="text-subtitle-1">{{ title }}</VCardTitle>
      <VCardText v-if="message" class="text-body-2 text-medium-emphasis">
        {{ message }}
      </VCardText>
      <VCardActions>
        <VSpacer />
        <VBtn variant="text" @click="onCancel">
          {{ cancelText }}
        </VBtn>
        <VBtn :color="danger ? 'error' : 'primary'" variant="tonal" @click="onConfirm">
          {{ confirmText }}
        </VBtn>
      </VCardActions>
    </VCard>
  </VDialog>
</template>

<script setup lang="ts">
const props = withDefaults(
  defineProps<{
    modelValue: boolean;
    title: string;
    message?: string;
    confirmText: string;
    cancelText: string;
    danger?: boolean;
  }>(),
  { danger: true },
);

const emit = defineEmits<{
  (e: "update:modelValue", v: boolean): void;
  (e: "confirm"): void;
  (e: "cancel"): void;
}>();

function onCancel(): void {
  emit("update:modelValue", false);
  emit("cancel");
}

function onConfirm(): void {
  emit("update:modelValue", false);
  emit("confirm");
}
</script>

