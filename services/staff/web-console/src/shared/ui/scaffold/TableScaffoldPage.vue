<template>
  <div>
    <PageHeader :title="t(titleKey)" :hint="hintKey ? t(hintKey) : undefined">
      <template #actions>
        <slot name="header-actions">
          <VBtn :disabled="loading" variant="tonal" prepend-icon="mdi-refresh" @click="refresh">
            {{ t("common.refresh") }}
          </VBtn>
        </slot>
      </template>
    </PageHeader>

    <slot name="below-header" />

    <VRow class="mt-4" density="compact">
      <VCol cols="12" sm="4">
        <VCard variant="tonal">
          <VCardText>
            <div class="text-caption text-medium-emphasis">{{ t("scaffold.metrics.total") }}</div>
            <div class="text-h6 font-weight-bold">{{ items.length }}</div>
          </VCardText>
        </VCard>
      </VCol>
      <VCol cols="12" sm="4">
        <VCard variant="tonal">
          <VCardText>
            <div class="text-caption text-medium-emphasis">{{ t("scaffold.metrics.filtered") }}</div>
            <div class="text-h6 font-weight-bold">{{ filteredItems.length }}</div>
          </VCardText>
        </VCard>
      </VCol>
      <VCol cols="12" sm="4">
        <VCard variant="tonal">
          <VCardText>
            <div class="text-caption text-medium-emphasis">{{ t("scaffold.metrics.updatedAt") }}</div>
            <div class="text-h6 font-weight-bold">{{ updatedAtLabel }}</div>
          </VCardText>
        </VCard>
      </VCol>
    </VRow>

    <VCard class="mt-4" variant="outlined">
      <VCardText>
        <VRow density="compact" class="align-end">
          <VCol cols="12" md="8">
            <VTextField
              v-model.trim="search"
              :label="t('scaffold.filters.search')"
              prepend-inner-icon="mdi-magnify"
              density="comfortable"
              hide-details
              clearable
            />
          </VCol>
          <VCol cols="12" md="4" class="d-flex justify-end ga-2 flex-wrap">
            <VMenu>
              <template #activator="{ props: menuProps }">
                <VBtn v-bind="menuProps" variant="tonal" prepend-icon="mdi-tune">
                  {{ t("scaffold.table.settings") }}
                </VBtn>
              </template>
              <VCard min-width="280">
                <VCardTitle class="text-subtitle-2">{{ t("scaffold.table.settings") }}</VCardTitle>
                <VCardText class="pt-0">
                  <div class="text-caption text-medium-emphasis">{{ t("scaffold.table.density") }}</div>
                  <VBtnToggle v-model="density" class="mt-2" divided density="compact" mandatory>
                    <VBtn value="default">{{ t("scaffold.table.densityDefault") }}</VBtn>
                    <VBtn value="comfortable">{{ t("scaffold.table.densityComfortable") }}</VBtn>
                    <VBtn value="compact">{{ t("scaffold.table.densityCompact") }}</VBtn>
                  </VBtnToggle>

                  <div class="text-caption text-medium-emphasis mt-4">{{ t("scaffold.table.columns") }}</div>
                  <div class="mt-2">
                    <VCheckbox
                      v-for="h in toggleableHeaders"
                      :key="h.key"
                      v-model="visibleHeaderKeys"
                      :value="h.key"
                      :label="h.title"
                      density="compact"
                      hide-details
                    />
                  </div>
                </VCardText>
              </VCard>
            </VMenu>
          </VCol>
        </VRow>
      </VCardText>
    </VCard>

    <VCard class="mt-4" variant="outlined">
      <VCardText>
        <VSkeletonLoader v-if="loading" type="table" />

        <VAlert v-else-if="errorMode" type="error" variant="tonal" :title="t('scaffold.states.errorTitle')">
          <div class="text-body-2">{{ t("scaffold.states.errorText") }}</div>
          <div class="mt-3 d-flex ga-2 flex-wrap">
            <VBtn variant="tonal" prepend-icon="mdi-refresh" @click="refresh">
              {{ t("scaffold.actions.retry") }}
            </VBtn>
          </div>
        </VAlert>

        <template v-else>
          <slot name="before-table" />

          <VAlert
            v-if="emptyMode"
            type="info"
            variant="tonal"
            :title="t('scaffold.states.emptyTitle')"
            class="mb-4"
          >
            <div class="text-body-2">{{ t("scaffold.states.emptyText") }}</div>
            <div class="mt-3 d-flex ga-2 flex-wrap">
              <VBtn variant="tonal" prepend-icon="mdi-refresh" @click="refresh">
                {{ t("scaffold.actions.retry") }}
              </VBtn>
              <VBtn variant="text" prepend-icon="mdi-close" @click="search = ''">
                {{ t("scaffold.actions.clearFilters") }}
              </VBtn>
            </div>
          </VAlert>

          <VDataTable
            :headers="effectiveHeaders"
            :items="filteredItems"
            :items-per-page="10"
            :density="density"
            :hover="true"
          >
            <template #item.actions="{ item }">
              <slot name="row-actions" :item="item">
                <VMenu>
                  <template #activator="{ props: actionProps }">
                    <VBtn v-bind="actionProps" icon="mdi-dots-horizontal" variant="text" />
                  </template>
                  <VList density="compact">
                    <VListItem prepend-icon="mdi-eye" :title="t('scaffold.rowActions.view')" />
                    <VListItem prepend-icon="mdi-pencil" :title="t('scaffold.rowActions.edit')" />
                    <VListItem prepend-icon="mdi-delete" :title="t('scaffold.rowActions.delete')" />
                  </VList>
                </VMenu>
              </slot>
            </template>

            <template #no-data>
              <div class="py-6 text-medium-emphasis">
                {{ t("scaffold.states.noRows") }}
              </div>
            </template>
          </VDataTable>

          <slot name="after-table" />
        </template>
      </VCardText>
    </VCard>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import { useRoute } from "vue-router";
import { useI18n } from "vue-i18n";

import PageHeader from "../PageHeader.vue";

type TableHeader = {
  title: string;
  key: string;
  align?: "start" | "center" | "end";
  sortable?: boolean;
  width?: string | number;
};

const props = defineProps<{
  titleKey: string;
  hintKey?: string;
  headers: ReadonlyArray<TableHeader>;
  items: ReadonlyArray<Record<string, unknown>>;
}>();

const { t } = useI18n({ useScope: "global" });
const route = useRoute();

const search = ref("");
const loading = ref(false);
const updatedAt = ref(new Date());

const density = ref<"default" | "comfortable" | "compact">("comfortable");

const toggleableHeaders = computed(() => props.headers.filter((h) => h.key !== "actions"));
const visibleHeaderKeys = ref<string[]>(toggleableHeaders.value.map((h) => h.key));

const effectiveHeaders = computed(() => {
  const visible = new Set(visibleHeaderKeys.value);
  return props.headers.filter((h) => h.key === "actions" || visible.has(h.key));
});

const filteredItems = computed(() => {
  const q = search.value.trim().toLowerCase();
  if (!q) return [...props.items];
  return props.items.filter((row) => JSON.stringify(row).toLowerCase().includes(q));
});

const emptyMode = computed(() => filteredItems.value.length === 0);
const errorMode = computed(() => String(route.query.error || "").toLowerCase() === "1");

const updatedAtLabel = computed(() => {
  const d = updatedAt.value;
  const hh = String(d.getHours()).padStart(2, "0");
  const mm = String(d.getMinutes()).padStart(2, "0");
  return `${hh}:${mm}`;
});

function refresh(): void {
  loading.value = true;
  window.setTimeout(() => {
    updatedAt.value = new Date();
    loading.value = false;
  }, 350);
}
</script>
