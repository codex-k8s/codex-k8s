<template>
  <VBtn
    v-if="showText"
    v-bind="attrs"
    :prepend-icon="icon"
    :title="titleValue"
    :aria-label="titleValue"
  >
    <slot>
      {{ label }}
    </slot>
  </VBtn>
  <VBtn
    v-else
    v-bind="attrs"
    icon
    :title="titleValue"
    :aria-label="titleValue"
  >
    <VIcon :icon="icon" />
  </VBtn>
</template>

<script setup lang="ts">
import { computed, useAttrs } from "vue";
import { useDisplay } from "vuetify";

defineOptions({ inheritAttrs: false });

const props = withDefaults(
  defineProps<{
    icon: string;
    label: string;
    title?: string;
    iconOnlyAt?: number;
  }>(),
  { title: "", iconOnlyAt: 1600 },
);

const attrs = useAttrs();
const display = useDisplay();

const showText = computed(() => display.width.value > props.iconOnlyAt);
const titleValue = computed(() => props.title || props.label);
</script>
