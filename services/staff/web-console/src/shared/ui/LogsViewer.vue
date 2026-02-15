<template>
  <VCard variant="outlined">
    <VCardTitle class="d-flex align-center justify-space-between ga-3 flex-wrap">
      <div class="d-flex align-center ga-2">
        <VIcon icon="mdi-text-long" />
        <span class="text-subtitle-1">{{ titleLabel }}</span>
      </div>

      <div class="d-flex align-center ga-2 flex-wrap justify-end">
        <VSelect
          v-model="tailLines"
          :items="tailLineOptions"
          density="compact"
          hide-details
          style="max-width: 120px"
        />
        <VSwitch v-model="followTail" :label="t('logs.follow')" hide-details density="compact" />
        <AdaptiveBtn
          variant="tonal"
          icon="mdi-refresh"
          :label="t('common.refresh')"
          :loading="loading"
          @click="emit('refresh', tailLines)"
        />
        <AdaptiveBtn
          variant="tonal"
          icon="mdi-content-copy"
          :label="t('common.copy')"
          @click="copyAll"
          :disabled="filteredLines.length === 0"
        />
        <AdaptiveBtn
          variant="tonal"
          icon="mdi-download"
          :label="t('logs.download')"
          @click="download"
          :disabled="filteredLines.length === 0"
        />
      </div>
    </VCardTitle>

    <VCardText class="pt-0">
      <div class="text-body-2 text-medium-emphasis mb-3 d-flex ga-3 flex-wrap">
        <span v-if="status">
          <strong>{{ t("logs.status") }}:</strong> {{ status }}
        </span>
        <span v-if="updatedAtLabel">
          <strong>{{ t("logs.updatedAt") }}:</strong> {{ updatedAtLabel }}
        </span>
        <span>
          <strong>{{ t("logs.lines") }}:</strong> {{ filteredLines.length }}
        </span>
      </div>

      <VTextField
        v-model.trim="search"
        :label="t('logs.search')"
        prepend-inner-icon="mdi-magnify"
        density="comfortable"
        hide-details
        clearable
      />

      <div ref="boxRef" class="logs-box mt-4">
        <div v-if="filteredLines.length === 0" class="py-8 text-medium-emphasis text-center">
          {{ t("logs.noLines") }}
        </div>
        <div
          v-for="(line, idx) in filteredLines"
          :key="`${idx}-${line}`"
          class="log-line"
          :class="{ match: searchLower && line.toLowerCase().includes(searchLower) }"
        >
          <span class="mono">{{ line }}</span>
        </div>
      </div>
    </VCardText>
  </VCard>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref, watch } from "vue";
import { useI18n } from "vue-i18n";

import AdaptiveBtn from "./AdaptiveBtn.vue";
import { useSnackbarStore } from "./feedback/snackbar-store";

const props = withDefaults(
  defineProps<{
    title?: string;
    lines: ReadonlyArray<string>;
    status?: string;
    updatedAtLabel?: string;
    loading?: boolean;
    fileName?: string;
  }>(),
  {
    title: "",
    status: "",
    updatedAtLabel: "",
    loading: false,
    fileName: "logs.txt",
  },
);

const emit = defineEmits<{
  (e: "refresh", tailLines: number): void;
}>();

const { t } = useI18n({ useScope: "global" });
const snackbar = useSnackbarStore();

const titleLabel = computed(() => props.title || t("logs.title"));

const boxRef = ref<HTMLElement | null>(null);
const search = ref("");
const followTail = ref(false);

const tailLineOptions = [100, 200, 500, 1000] as const;
const tailLines = ref<(typeof tailLineOptions)[number]>(200);

const searchLower = computed(() => search.value.trim().toLowerCase());
const filteredLines = computed(() => {
  const q = searchLower.value;
  if (!q) return [...props.lines];
  return props.lines.filter((l) => l.toLowerCase().includes(q));
});

let intervalId: number | null = null;

watch(
  followTail,
  (enabled) => {
    if (intervalId !== null) {
      window.clearInterval(intervalId);
      intervalId = null;
    }
    if (!enabled) return;
    intervalId = window.setInterval(() => emit("refresh", tailLines.value), 2000);
  },
  { immediate: true },
);

watch(
  [() => props.lines, followTail],
  async () => {
    if (!followTail.value) return;
    await nextTick();
    const el = boxRef.value;
    if (!el) return;
    el.scrollTop = el.scrollHeight;
  },
  { deep: false },
);

onBeforeUnmount(() => {
  if (intervalId !== null) {
    window.clearInterval(intervalId);
  }
});

async function copyAll(): Promise<void> {
  try {
    await navigator.clipboard.writeText(filteredLines.value.join("\n"));
    snackbar.success(t("common.copied"));
  } catch {
    snackbar.error(t("errors.copyFailed"));
  }
}

function download(): void {
  const blob = new Blob([filteredLines.value.join("\n") + "\n"], { type: "text/plain;charset=utf-8" });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = props.fileName;
  a.click();
  URL.revokeObjectURL(url);
}
</script>

<style scoped>
.logs-box {
  border: 1px solid rgba(17, 24, 39, 0.12);
  border-radius: 12px;
  background: rgba(17, 24, 39, 0.03);
  max-height: 420px;
  overflow: auto;
}
.log-line {
  padding: 6px 10px;
  border-bottom: 1px dashed rgba(17, 24, 39, 0.1);
}
.log-line:last-child {
  border-bottom: none;
}
.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 12px;
  white-space: pre-wrap;
  word-break: break-word;
}
.match {
  background: rgba(255, 234, 168, 0.55);
}
</style>
