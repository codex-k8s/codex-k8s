<template>
  <VAlert v-if="visible" type="warning" variant="tonal" :title="title">
    <div class="text-body-2">
      {{ text }}
    </div>

    <div class="mt-3 d-flex align-center justify-space-between ga-2 flex-wrap">
      <VCheckbox v-model="dontShowAgain" :label="t('common.doNotShowAgain')" density="compact" hide-details />

      <VBtn variant="text" icon="mdi-close" :title="t('common.close')" @click="visible = false" />
    </div>
  </VAlert>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { useI18n } from "vue-i18n";

import { getCookie, setCookie } from "../lib/cookies";

const props = defineProps<{
  alertId: string;
  title: string;
  text: string;
}>();

const { t } = useI18n({ useScope: "global" });

const cookieKey = computed(() => `codexk8s_alert_hide_${props.alertId}`);
const visible = ref(getCookie(cookieKey.value) !== "1");
const dontShowAgain = ref(false);

watch(
  dontShowAgain,
  (v) => {
    if (!v) return;
    // TODO: Persist dismissed alerts in backend user params (cookies are a temporary UX shortcut).
    setCookie(cookieKey.value, "1", { maxAgeDays: 365, sameSite: "Lax" });
    visible.value = false;
  },
  { immediate: false },
);
</script>

