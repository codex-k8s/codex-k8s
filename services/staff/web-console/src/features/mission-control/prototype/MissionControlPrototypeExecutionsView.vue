<template>
  <div class="mission-executions">
    <section class="mission-executions__hero">
      <div>
        <div class="mission-executions__eyebrow">Executions</div>
        <h2 class="mission-executions__title">Диагностика исполнений по артефактам</h2>
        <p class="mission-executions__summary">
          Здесь живут технические исполнения. На Home и Workspace они скрыты за карточками задач и PR, чтобы не
          перегружать основной поток управления.
        </p>
      </div>
    </section>

    <section class="mission-executions__groups">
      <article v-for="group in groups" :key="group.groupId" class="mission-executions__group">
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
            <div class="mission-executions__item-topline">
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
    </section>
  </div>
</template>

<script setup lang="ts">
import { missionArtifactKindLabel } from "./presenters";
import type { MissionExecutionGroup, MissionExecutionStatus } from "./types";

defineProps<{
  groups: MissionExecutionGroup[];
}>();

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
.mission-executions__group {
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

.mission-executions__summary {
  margin: 8px 0 0;
  max-width: 840px;
  font-size: 0.95rem;
  line-height: 1.55;
  color: rgb(93, 102, 116);
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

.mission-executions__group-title {
  font-size: 1rem;
  font-weight: 700;
  color: rgb(31, 36, 43);
}

.mission-executions__group-subtitle,
.mission-executions__group-summary,
.mission-executions__item-summary,
.mission-executions__item-topline {
  font-size: 0.86rem;
  line-height: 1.5;
  color: rgb(98, 107, 121);
}

.mission-executions__group-summary {
  margin-top: 8px;
}

.mission-executions__list {
  display: grid;
  gap: 12px;
  margin-top: 14px;
}

.mission-executions__item {
  display: grid;
  gap: 8px;
  padding: 14px;
  border-radius: 18px;
  background: rgba(248, 250, 252, 0.92);
}

.mission-executions__item-topline {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  align-items: center;
}

.mission-executions__item-title {
  font-size: 0.95rem;
  font-weight: 700;
  color: rgb(31, 36, 43);
}
</style>
