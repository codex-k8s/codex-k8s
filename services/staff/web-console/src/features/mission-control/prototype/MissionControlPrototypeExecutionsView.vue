<template>
  <div class="mission-executions">
    <section class="mission-executions__hero">
      <div>
        <div class="mission-executions__eyebrow">Исполнения</div>
        <h2 class="mission-executions__title">Диагностика исполнений по артефактам</h2>
        <p class="mission-executions__summary">
          Здесь живут только технические исполнения вокруг GitHub Issue и PR. На главном экране и в инициативе они скрыты
          за артефактами, чтобы не перегружать основной поток управления.
        </p>
      </div>
    </section>

    <section class="mission-executions__stats">
      <article class="mission-executions__stat">
        <span>Всего</span>
        <strong>{{ totalExecutions }}</strong>
      </article>
      <article class="mission-executions__stat">
        <span>Идут</span>
        <strong>{{ runningExecutions }}</strong>
      </article>
      <article class="mission-executions__stat">
        <span>Ожидают</span>
        <strong>{{ waitingExecutions }}</strong>
      </article>
      <article class="mission-executions__stat">
        <span>С ошибками</span>
        <strong>{{ failedExecutions }}</strong>
      </article>
    </section>

    <div class="mission-executions__filters">
      <VBtnToggle v-model="statusFilter" divided mandatory density="comfortable">
        <VBtn value="all">Все</VBtn>
        <VBtn value="running">Идут</VBtn>
        <VBtn value="waiting">Ожидают</VBtn>
        <VBtn value="failed">С ошибками</VBtn>
      </VBtnToggle>
    </div>

    <section class="mission-executions__groups">
      <article v-for="group in filteredGroups" :key="group.groupId" class="mission-executions__group">
        <div class="mission-executions__group-head">
          <div>
            <div class="mission-executions__group-title">{{ group.artifactTitle }}</div>
            <div class="mission-executions__group-subtitle">
              {{ group.initiativeTitle }} · {{ kindLabel(group.artifactKind) }}
            </div>
          </div>
          <VChip size="small" variant="outlined">{{ group.items.length }}</VChip>
        </div>

        <div class="mission-executions__group-summary">{{ group.summary }}</div>

        <div class="mission-executions__list">
          <article v-for="item in group.items" :key="item.executionId" class="mission-executions__item">
            <div class="mission-executions__item-status">
              <VChip size="x-small" :color="statusColor(item.status)" variant="tonal">{{ statusLabel(item.status) }}</VChip>
              <span>{{ item.agentRoleLabel }}</span>
              <span>{{ item.startedAtLabel }}</span>
              <span>{{ item.durationLabel }}</span>
            </div>
            <div class="mission-executions__item-title">{{ item.title }}</div>
            <div class="mission-executions__item-summary">{{ item.summary }}</div>
          </article>
        </div>
      </article>

      <article v-if="filteredGroups.length === 0" class="mission-executions__empty">
        По выбранному фильтру исполнений нет.
      </article>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";

import { missionArtifactKindLabel } from "./presenters";
import type { MissionExecutionGroup, MissionExecutionStatus } from "./types";

const props = defineProps<{
  groups: MissionExecutionGroup[];
}>();

const statusFilter = ref<"all" | MissionExecutionStatus>("all");

const allItems = computed(() => props.groups.flatMap((group) => group.items));
const totalExecutions = computed(() => allItems.value.length);
const runningExecutions = computed(() => allItems.value.filter((item) => item.status === "running").length);
const waitingExecutions = computed(() => allItems.value.filter((item) => item.status === "waiting").length);
const failedExecutions = computed(() => allItems.value.filter((item) => item.status === "failed").length);
const filteredGroups = computed(() => {
  if (statusFilter.value === "all") {
    return props.groups;
  }

  return props.groups
    .map((group) => ({
      ...group,
      items: group.items.filter((item) => item.status === statusFilter.value),
    }))
    .filter((group) => group.items.length > 0);
});

function kindLabel(kind: MissionExecutionGroup["artifactKind"]): string {
  return missionArtifactKindLabel(kind);
}

function statusColor(status: MissionExecutionStatus): string {
  switch (status) {
    case "running":
      return "info";
    case "waiting":
      return "warning";
    case "failed":
      return "error";
    case "done":
      return "success";
  }
}

function statusLabel(status: MissionExecutionStatus): string {
  switch (status) {
    case "running":
      return "Идет";
    case "waiting":
      return "Ожидает";
    case "failed":
      return "Ошибка";
    case "done":
      return "Завершено";
  }
}
</script>

<style scoped>
.mission-executions {
  display: grid;
  gap: 18px;
}

.mission-executions__hero,
.mission-executions__stat,
.mission-executions__group,
.mission-executions__empty {
  padding: 18px;
  border-radius: 24px;
  border: 1px solid rgba(223, 227, 233, 0.92);
  background: rgba(255, 255, 255, 0.94);
  box-shadow: 0 16px 34px rgba(26, 29, 35, 0.05);
}

.mission-executions__eyebrow {
  font-size: 0.78rem;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  font-weight: 700;
  color: rgb(79, 91, 112);
}

.mission-executions__title {
  margin: 8px 0 0;
  font-size: 1.5rem;
  line-height: 1.2;
  color: rgb(31, 36, 43);
}

.mission-executions__summary,
.mission-executions__group-subtitle,
.mission-executions__group-summary,
.mission-executions__item-summary,
.mission-executions__item-status {
  font-size: 0.86rem;
  line-height: 1.5;
  color: rgb(98, 107, 121);
}

.mission-executions__stats {
  display: grid;
  gap: 14px;
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.mission-executions__stat {
  display: grid;
  gap: 8px;
}

.mission-executions__stat span {
  font-size: 0.82rem;
  color: rgb(98, 107, 121);
}

.mission-executions__stat strong {
  font-size: 1.6rem;
  color: rgb(31, 36, 43);
}

.mission-executions__filters {
  display: flex;
  justify-content: flex-start;
}

.mission-executions__groups {
  display: grid;
  gap: 14px;
}

.mission-executions__group-head {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  align-items: flex-start;
}

.mission-executions__group-title,
.mission-executions__item-title {
  font-size: 1rem;
  font-weight: 700;
  color: rgb(31, 36, 43);
}

.mission-executions__group-summary {
  margin-top: 8px;
}

.mission-executions__list {
  display: grid;
  gap: 10px;
  margin-top: 14px;
}

.mission-executions__item {
  display: grid;
  gap: 6px;
  padding: 12px 14px;
  border-radius: 18px;
  background: rgba(248, 250, 252, 0.92);
}

.mission-executions__item-status {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  align-items: center;
}

.mission-executions__empty {
  color: rgb(91, 100, 114);
}

@media (max-width: 980px) {
  .mission-executions__stats {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 640px) {
  .mission-executions__stats {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
