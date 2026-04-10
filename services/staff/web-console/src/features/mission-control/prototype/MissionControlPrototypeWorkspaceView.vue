<template>
  <div v-if="initiative && workflow" class="mission-workspace">
    <section class="mission-workspace__hero">
      <div class="mission-workspace__eyebrow">Инициатива</div>
      <h2 class="mission-workspace__title">{{ initiative.title }}</h2>
      <p class="mission-workspace__summary">{{ initiative.summary }}</p>

      <div class="mission-workspace__chips">
        <VChip size="small" variant="tonal" color="primary">{{ workflow.title }}</VChip>
        <VChip size="small" variant="tonal" :color="toneColor(initiative.attentionTone)">{{ initiative.attentionLabel }}</VChip>
      </div>

      <div class="mission-workspace__metrics">
        <article class="mission-workspace__metric">
          <span>Следующий шаг</span>
          <strong>{{ initiative.nextAction }}</strong>
        </article>
        <article class="mission-workspace__metric">
          <span>Текущий статус</span>
          <strong>{{ initiative.statusLabel }}</strong>
        </article>
        <article class="mission-workspace__metric">
          <span>Исполнений за инициативой</span>
          <strong>{{ initiative.runSummary.total }}</strong>
        </article>
      </div>
    </section>

    <div class="mission-workspace__modebar">
      <VBtnToggle divided mandatory :model-value="workspaceView" @update:model-value="onUpdateView">
        <VBtn value="overview">Обзор</VBtn>
        <VBtn value="flow">Поток</VBtn>
        <VBtn value="artifacts">Issue и PR</VBtn>
        <VBtn value="activity">Активность</VBtn>
      </VBtnToggle>

      <VSpacer />

      <VBtn variant="text" prepend-icon="mdi-shape-outline" @click="$emit('open-studio')">Редактор workflow</VBtn>
      <VBtn variant="text" prepend-icon="mdi-timeline-outline" @click="$emit('open-executions')">Исполнения</VBtn>
    </div>

    <section v-if="workspaceView === 'overview'" class="mission-workspace__overview">
      <div class="mission-workspace__overview-grid">
        <article class="mission-workspace__card mission-workspace__card--primary">
          <div class="mission-workspace__section-label">Сейчас в работе</div>
          <div class="mission-workspace__section-title">{{ currentStage?.label || "Этап не определен" }}</div>
          <p class="mission-workspace__section-copy">{{ currentStage?.summary || initiative.summary }}</p>

          <div class="mission-workspace__action-block">
            <span>Что должен сделать оператор</span>
            <strong>{{ initiative.nextAction }}</strong>
          </div>

          <div class="mission-workspace__object-grid">
            <button
              type="button"
              class="mission-workspace__object-card"
              :disabled="!currentIssue"
              @click="focusArtifact(currentIssue?.artifactId)"
            >
              <span>Текущий issue</span>
              <strong>{{ currentIssue?.title || "Issue еще не создан" }}</strong>
            </button>
            <button
              type="button"
              class="mission-workspace__object-card"
              :disabled="!currentPr"
              @click="focusArtifact(currentPr?.artifactId)"
            >
              <span>Текущий PR</span>
              <strong>{{ currentPr?.title || "PR появится на этапе разработки" }}</strong>
            </button>
            <button
              type="button"
              class="mission-workspace__object-card"
              :disabled="!nextFollowUp"
              @click="focusArtifact(nextFollowUp?.artifactId)"
            >
              <span>Следующий follow-up</span>
              <strong>{{ nextFollowUp?.title || "Следующий issue будет создан позже" }}</strong>
            </button>
          </div>

          <div class="mission-workspace__quick-actions">
            <VBtn
              v-if="currentIssue"
              size="small"
              color="primary"
              variant="tonal"
              @click="focusArtifact(currentIssue.artifactId)"
            >
              К issue
            </VBtn>
            <VBtn v-if="currentPr" size="small" variant="tonal" @click="focusArtifact(currentPr.artifactId)">К PR</VBtn>
            <VBtn size="small" variant="text" @click="emit('update:view', 'flow')">Показать поток</VBtn>
          </div>
        </article>

        <article class="mission-workspace__card">
          <div class="mission-workspace__section-label">Как это будет исполняться</div>
          <div class="mission-workspace__section-title">Issue и PR создаются через gh</div>
          <div class="mission-workspace__policy-list">
            <div class="mission-workspace__policy-item">
              Workflow добавляет в prompt этап, ожидаемый артефакт и следующий follow-up.
            </div>
            <div class="mission-workspace__policy-item">
              Агент открывает GitHub Issue и PR через <code>gh</code>, а не через MCP.
            </div>
            <div class="mission-workspace__policy-item">
              При приемке машина проверяет watermark-блок и обязательные поля в body.
            </div>
          </div>
        </article>
      </div>

      <article class="mission-workspace__card">
        <div class="mission-workspace__section-label">Лента этапов</div>
        <div class="mission-workspace__stage-strip">
          <button
            v-for="stage in stageViews"
            :key="stage.stageKey"
            type="button"
            class="mission-workspace__stage-pill"
            :class="`mission-workspace__stage-pill--${stage.status}`"
            @click="focusStage(stage.stageKey)"
          >
            <div class="mission-workspace__stage-pill-title">{{ stage.label }}</div>
            <div class="mission-workspace__stage-pill-summary">{{ stage.exitLabel }}</div>
          </button>
        </div>
      </article>

      <div class="mission-workspace__overview-grid">
        <article class="mission-workspace__card">
          <div class="mission-workspace__section-label">Текущие Issue и PR</div>
          <div class="mission-workspace__artifact-list">
            <button
              v-for="artifact in currentStageArtifacts"
              :key="artifact.artifactId"
              type="button"
              class="mission-workspace__artifact-card"
              :class="{ 'mission-workspace__artifact-card--selected': artifact.selected }"
              @click="$emit('select-artifact', artifact.artifactId)"
            >
              <div class="mission-workspace__artifact-topline">
                <VChip size="x-small" variant="outlined">{{ kindLabel(artifact.kind) }}</VChip>
                <VChip size="x-small" :color="artifactColor(artifact.status)" variant="tonal">
                  {{ artifactStatusLabel(artifact.status) }}
                </VChip>
              </div>
              <div class="mission-workspace__artifact-title">{{ artifact.title }}</div>
              <div class="mission-workspace__artifact-summary">{{ artifact.summary }}</div>
            </button>
          </div>
        </article>

        <article class="mission-workspace__card">
          <div class="mission-workspace__section-label">Последняя активность</div>
          <div class="mission-workspace__activity-list">
            <article v-for="item in recentActivity" :key="item.itemId" class="mission-workspace__activity-item">
              <div class="mission-workspace__activity-meta">
                <VChip size="x-small" :color="toneColor(item.tone)" variant="tonal">{{ item.actorLabel }}</VChip>
                <span>{{ item.happenedAtLabel }}</span>
              </div>
              <div class="mission-workspace__activity-title">{{ item.title }}</div>
              <div class="mission-workspace__activity-summary">{{ item.summary }}</div>
              <div class="mission-workspace__activity-target">
                {{ activityTargetLabel(item.targetKind) }}: {{ item.targetLabel }}
              </div>
            </article>
          </div>
        </article>
      </div>
    </section>

    <section v-else-if="workspaceView === 'flow'" class="mission-workspace__flow">
      <div class="mission-workspace__flow-toolbar">
        <div>
          <div class="mission-workspace__section-title">Карта этапов</div>
          <div class="mission-workspace__section-copy">
            Поток строится автоматически по этапам workflow. Здесь нет ручного drag-and-drop, чтобы связи не ломались.
          </div>
        </div>
        <VSwitch
          v-model="showRunSummary"
          density="compact"
          color="primary"
          inset
          hide-details
          label="Показывать summary executions"
        />
      </div>

      <div class="mission-workspace__flow-strip">
        <div
          v-for="(stage, index) in stageViews"
          :key="stage.stageKey"
          class="mission-workspace__flow-stage-wrap"
        >
          <button
            type="button"
            class="mission-workspace__flow-stage"
            :class="`mission-workspace__flow-stage--${stage.status}`"
            @click="focusStage(stage.stageKey)"
          >
            <div class="mission-workspace__flow-stage-head">
              <div class="mission-workspace__flow-stage-title">{{ stage.label }}</div>
              <VChip size="x-small" :color="statusColor(stage.status)" variant="tonal">{{ statusLabel(stage.status) }}</VChip>
            </div>
            <div class="mission-workspace__flow-stage-copy">{{ stage.summary }}</div>
            <div class="mission-workspace__flow-stage-meta">
              <span>Issue: {{ artifactCount(stage.stageKey, "issue") }}</span>
              <span>PR: {{ artifactCount(stage.stageKey, "pr") }}</span>
              <span v-if="showRunSummary">Runs: {{ stageRunCount(stage.stageKey) }}</span>
            </div>

            <div class="mission-workspace__flow-stage-artifacts">
              <button
                v-for="artifact in artifactsForStage(stage.stageKey)"
                :key="artifact.artifactId"
                type="button"
                class="mission-workspace__flow-artifact"
                @click.stop="$emit('select-artifact', artifact.artifactId)"
              >
                <span>{{ kindLabel(artifact.kind) }}</span>
                <strong>{{ artifact.title }}</strong>
              </button>
            </div>
          </button>

          <div v-if="index < stageViews.length - 1" class="mission-workspace__flow-arrow">
            <VIcon icon="mdi-arrow-right" size="18" />
          </div>
        </div>
      </div>

      <article v-if="selectedArtifact" class="mission-workspace__card mission-workspace__card--detail">
        <div class="mission-workspace__section-label">Выбранный артефакт</div>
        <div class="mission-workspace__section-title">{{ selectedArtifact.title }}</div>
        <div class="mission-workspace__section-copy">{{ selectedArtifact.summary }}</div>
        <div class="mission-workspace__detail-meta">
          <span>{{ kindLabel(selectedArtifact.kind) }}</span>
          <span>{{ artifactStatusLabel(selectedArtifact.status) }}</span>
          <span>{{ selectedArtifact.ownerLabel }}</span>
          <span>Исполнений: {{ selectedArtifact.runSummary.total }}</span>
        </div>
      </article>
    </section>

    <section v-else-if="workspaceView === 'artifacts'" class="mission-workspace__artifacts">
      <article v-for="stage in stageViews" :key="stage.stageKey" class="mission-workspace__artifact-section">
        <div class="mission-workspace__section-title">{{ stage.label }}</div>
        <div class="mission-workspace__section-copy">{{ stage.summary }}</div>

        <div class="mission-workspace__artifact-list">
          <button
            v-for="artifact in artifactsForStage(stage.stageKey)"
            :key="artifact.artifactId"
            type="button"
            class="mission-workspace__artifact-card"
            :class="{ 'mission-workspace__artifact-card--selected': artifact.selected }"
            @click="$emit('select-artifact', artifact.artifactId)"
          >
            <div class="mission-workspace__artifact-topline">
              <VChip size="x-small" variant="outlined">{{ kindLabel(artifact.kind) }}</VChip>
              <VChip size="x-small" :color="artifactColor(artifact.status)" variant="tonal">
                {{ artifactStatusLabel(artifact.status) }}
              </VChip>
            </div>
            <div class="mission-workspace__artifact-title">{{ artifact.title }}</div>
            <div class="mission-workspace__artifact-summary">{{ artifact.summary }}</div>
            <div class="mission-workspace__artifact-meta">
              <span>{{ artifact.ownerLabel }}</span>
              <span>{{ artifact.updatedAtLabel }}</span>
              <span>Исполнений: {{ artifact.runSummary.total }}</span>
            </div>
          </button>
        </div>
      </article>
    </section>

    <section v-else class="mission-workspace__activity">
      <article v-for="item in activity" :key="item.itemId" class="mission-workspace__activity-item">
        <div class="mission-workspace__activity-meta">
          <VChip size="x-small" :color="toneColor(item.tone)" variant="tonal">{{ item.actorLabel }}</VChip>
          <span>{{ item.happenedAtLabel }}</span>
          <span>{{ activityTargetLabel(item.targetKind) }}</span>
        </div>
        <div class="mission-workspace__activity-title">{{ item.title }}</div>
        <div class="mission-workspace__activity-summary">{{ item.summary }}</div>
        <div class="mission-workspace__activity-target">{{ item.targetLabel }}</div>
      </article>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";

import {
  missionArtifactKindLabel,
  missionArtifactStatusColor,
  missionAttentionToneColor,
} from "./presenters";
import type {
  MissionActivityItem,
  MissionArtifactStatus,
  MissionInitiative,
  MissionInitiativeWorkspaceView,
  MissionWorkspaceArtifactView,
  MissionWorkspaceStageView,
  MissionWorkflowStageStatus,
  MissionWorkflowTemplate,
} from "./types";

const props = defineProps<{
  initiative: MissionInitiative | null;
  workflow: MissionWorkflowTemplate | null;
  workspaceView: MissionInitiativeWorkspaceView;
  stageViews: MissionWorkspaceStageView[];
  artifactViews: MissionWorkspaceArtifactView[];
  selectedArtifact: MissionWorkspaceArtifactView | null;
  activity: MissionActivityItem[];
}>();

const emit = defineEmits<{
  (event: "update:view", nextView: MissionInitiativeWorkspaceView): void;
  (event: "select-artifact", artifactId: string): void;
  (event: "open-executions"): void;
  (event: "open-studio"): void;
}>();

const showRunSummary = ref(false);

const currentStage = computed(() =>
  props.stageViews.find((stage) => stage.status === "active" || stage.status === "blocked" || stage.status === "attention") ??
  props.stageViews[0] ??
  null,
);

const currentStageArtifacts = computed(() =>
  currentStage.value ? props.artifactViews.filter((artifact) => artifact.stageKey === currentStage.value?.stageKey) : [],
);

const currentIssue = computed(() => primaryArtifactForKind(currentStageArtifacts.value, props.artifactViews, "issue"));
const currentPr = computed(() => primaryArtifactForKind(currentStageArtifacts.value, props.artifactViews, "pr"));
const nextFollowUp = computed(
  () =>
    props.artifactViews.find((artifact) => artifact.status === "draft") ??
    props.artifactViews.find(
      (artifact) => artifact.kind === "issue" && artifact.stageKey !== currentStage.value?.stageKey && artifact.status !== "done",
    ) ??
    null,
);
const recentActivity = computed(() => props.activity.slice(0, 4));

function onUpdateView(nextValue: string): void {
  if (nextValue === "overview" || nextValue === "flow" || nextValue === "artifacts" || nextValue === "activity") {
    emit("update:view", nextValue);
  }
}

function toneColor(tone: MissionActivityItem["tone"] | MissionInitiative["attentionTone"]): string {
  return missionAttentionToneColor(tone);
}

function artifactColor(status: MissionArtifactStatus): string {
  return missionArtifactStatusColor(status);
}

function kindLabel(kind: MissionWorkspaceArtifactView["kind"]): string {
  return missionArtifactKindLabel(kind);
}

function statusColor(status: MissionWorkflowStageStatus): string {
  switch (status) {
    case "done":
      return "success";
    case "blocked":
      return "error";
    case "active":
    case "attention":
      return "warning";
    case "pending":
      return "info";
  }
}

function statusLabel(status: MissionWorkflowStageStatus): string {
  switch (status) {
    case "done":
      return "Готово";
    case "blocked":
      return "Блокер";
    case "active":
      return "В работе";
    case "attention":
      return "Нужно решение";
    case "pending":
      return "Дальше по очереди";
  }
}

function artifactStatusLabel(status: MissionArtifactStatus): string {
  switch (status) {
    case "draft":
      return "Черновик";
    case "active":
      return "В работе";
    case "review":
      return "На review";
    case "blocked":
      return "Блокер";
    case "done":
      return "Готово";
  }
}

function activityTargetLabel(kind: MissionActivityItem["targetKind"]): string {
  switch (kind) {
    case "issue":
      return "Issue";
    case "pr":
      return "PR";
    case "run":
      return "Run";
    case "stage":
      return "Этап";
  }
}

function artifactsForStage(stageKey: MissionWorkspaceStageView["stageKey"]): MissionWorkspaceArtifactView[] {
  return props.artifactViews.filter((artifact) => artifact.stageKey === stageKey);
}

function artifactCount(stageKey: MissionWorkspaceStageView["stageKey"], kind: MissionWorkspaceArtifactView["kind"]): number {
  return props.artifactViews.filter((artifact) => artifact.stageKey === stageKey && artifact.kind === kind).length;
}

function stageRunCount(stageKey: MissionWorkspaceStageView["stageKey"]): number {
  return props.artifactViews
    .filter((artifact) => artifact.stageKey === stageKey)
    .reduce((sum, artifact) => sum + artifact.runSummary.total, 0);
}

function focusArtifact(artifactId?: string): void {
  if (!artifactId) {
    return;
  }

  emit("select-artifact", artifactId);
  emit("update:view", "artifacts");
}

function focusStage(stageKey: MissionWorkspaceStageView["stageKey"]): void {
  const stageArtifact = props.artifactViews.find((artifact) => artifact.stageKey === stageKey);
  if (stageArtifact) {
    emit("select-artifact", stageArtifact.artifactId);
  }
}

function primaryArtifactForKind(
  stageArtifacts: MissionWorkspaceArtifactView[],
  allArtifacts: MissionWorkspaceArtifactView[],
  kind: MissionWorkspaceArtifactView["kind"],
): MissionWorkspaceArtifactView | null {
  return (
    stageArtifacts.find((artifact) => artifact.kind === kind && artifact.status === "active") ??
    stageArtifacts.find((artifact) => artifact.kind === kind && artifact.status === "review") ??
    stageArtifacts.find((artifact) => artifact.kind === kind) ??
    allArtifacts.find((artifact) => artifact.kind === kind && artifact.status !== "done") ??
    allArtifacts.find((artifact) => artifact.kind === kind) ??
    null
  );
}
</script>

<style scoped>
.mission-workspace {
  display: grid;
  gap: 18px;
}

.mission-workspace__hero,
.mission-workspace__card,
.mission-workspace__artifact-section,
.mission-workspace__activity-item {
  border-radius: 24px;
  border: 1px solid rgba(223, 227, 233, 0.92);
  background: rgba(255, 255, 255, 0.94);
  box-shadow: 0 16px 34px rgba(26, 29, 35, 0.05);
}

.mission-workspace__hero {
  display: grid;
  gap: 12px;
  padding: 22px;
  background:
    radial-gradient(circle at top right, rgba(201, 229, 255, 0.26), transparent 30%),
    linear-gradient(145deg, rgba(252, 252, 250, 0.98), rgba(245, 248, 255, 0.94));
}

.mission-workspace__eyebrow,
.mission-workspace__section-label {
  font-size: 0.78rem;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  font-weight: 700;
  color: rgb(89, 98, 113);
}

.mission-workspace__title {
  margin: 0;
  font-size: 1.75rem;
  line-height: 1.2;
  color: rgb(31, 36, 43);
}

.mission-workspace__summary,
.mission-workspace__section-copy,
.mission-workspace__artifact-summary,
.mission-workspace__activity-summary,
.mission-workspace__activity-target,
.mission-workspace__artifact-meta,
.mission-workspace__flow-stage-copy,
.mission-workspace__flow-stage-meta {
  font-size: 0.9rem;
  line-height: 1.55;
  color: rgb(96, 104, 118);
}

.mission-workspace__chips,
.mission-workspace__quick-actions,
.mission-workspace__detail-meta,
.mission-workspace__artifact-meta,
.mission-workspace__activity-meta {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  align-items: center;
}

.mission-workspace__metrics,
.mission-workspace__overview-grid {
  display: grid;
  gap: 14px;
  grid-template-columns: 2fr 1fr;
}

.mission-workspace__metric,
.mission-workspace__object-card {
  display: grid;
  gap: 6px;
  padding: 14px 16px;
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.9);
  border: 1px solid rgba(223, 228, 235, 0.92);
  text-align: left;
}

.mission-workspace__metric span,
.mission-workspace__object-card span {
  font-size: 0.76rem;
  color: rgb(102, 111, 124);
}

.mission-workspace__metric strong,
.mission-workspace__object-card strong,
.mission-workspace__section-title {
  font-size: 1rem;
  line-height: 1.45;
  color: rgb(31, 36, 43);
}

.mission-workspace__modebar {
  display: flex;
  gap: 12px;
  align-items: center;
  flex-wrap: wrap;
}

.mission-workspace__overview,
.mission-workspace__flow,
.mission-workspace__artifacts,
.mission-workspace__activity {
  display: grid;
  gap: 16px;
}

.mission-workspace__card,
.mission-workspace__artifact-section {
  display: grid;
  gap: 14px;
  padding: 18px;
}

.mission-workspace__card--primary {
  background: linear-gradient(145deg, rgba(255, 251, 243, 0.96), rgba(255, 255, 255, 0.94));
}

.mission-workspace__action-block {
  display: grid;
  gap: 6px;
  padding: 14px 16px;
  border-radius: 18px;
  background: rgba(248, 250, 252, 0.92);
}

.mission-workspace__action-block span {
  font-size: 0.78rem;
  color: rgb(102, 111, 124);
}

.mission-workspace__action-block strong {
  font-size: 1rem;
  line-height: 1.45;
  color: rgb(31, 36, 43);
}

.mission-workspace__object-grid {
  display: grid;
  gap: 12px;
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.mission-workspace__object-card:disabled {
  opacity: 0.65;
}

.mission-workspace__policy-list,
.mission-workspace__activity-list,
.mission-workspace__artifact-list {
  display: grid;
  gap: 12px;
}

.mission-workspace__policy-item {
  padding: 12px 14px;
  border-radius: 16px;
  background: rgba(248, 250, 252, 0.92);
  border: 1px solid rgba(224, 229, 235, 0.92);
  font-size: 0.88rem;
  line-height: 1.5;
  color: rgb(88, 97, 112);
}

.mission-workspace__stage-strip {
  display: grid;
  gap: 10px;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
}

.mission-workspace__stage-pill {
  display: grid;
  gap: 6px;
  padding: 14px;
  border-radius: 18px;
  text-align: left;
  border: 1px solid rgba(224, 229, 235, 0.92);
  background: rgba(248, 250, 252, 0.94);
}

.mission-workspace__stage-pill--active,
.mission-workspace__flow-stage--active {
  border-color: rgba(245, 173, 74, 0.72);
  background: rgba(255, 248, 236, 0.98);
}

.mission-workspace__stage-pill--blocked,
.mission-workspace__flow-stage--blocked {
  border-color: rgba(239, 106, 94, 0.56);
  background: rgba(255, 242, 240, 0.98);
}

.mission-workspace__stage-pill--done,
.mission-workspace__flow-stage--done {
  border-color: rgba(125, 201, 152, 0.56);
  background: rgba(241, 252, 245, 0.98);
}

.mission-workspace__stage-pill-title,
.mission-workspace__flow-stage-title,
.mission-workspace__artifact-title,
.mission-workspace__activity-title {
  font-size: 0.98rem;
  font-weight: 700;
  color: rgb(31, 36, 43);
}

.mission-workspace__stage-pill-summary {
  font-size: 0.84rem;
  line-height: 1.45;
  color: rgb(95, 104, 118);
}

.mission-workspace__artifact-card {
  display: grid;
  gap: 10px;
  padding: 14px;
  border-radius: 18px;
  background: rgba(248, 250, 252, 0.92);
  border: 1px solid rgba(224, 229, 235, 0.92);
  text-align: left;
}

.mission-workspace__artifact-card--selected {
  border-color: rgba(78, 129, 253, 0.48);
  box-shadow: inset 0 0 0 1px rgba(78, 129, 253, 0.24);
}

.mission-workspace__artifact-topline {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  align-items: center;
}

.mission-workspace__flow-toolbar {
  display: flex;
  justify-content: space-between;
  gap: 14px;
  align-items: flex-start;
}

.mission-workspace__flow-strip {
  display: flex;
  gap: 12px;
  overflow-x: auto;
  padding-bottom: 8px;
}

.mission-workspace__flow-stage-wrap {
  display: flex;
  gap: 12px;
  align-items: center;
  flex: 0 0 auto;
}

.mission-workspace__flow-stage {
  width: 260px;
  display: grid;
  gap: 12px;
  padding: 16px;
  border-radius: 22px;
  border: 1px solid rgba(223, 228, 235, 0.92);
  background: rgba(255, 255, 255, 0.94);
  text-align: left;
}

.mission-workspace__flow-stage-head {
  display: flex;
  justify-content: space-between;
  gap: 10px;
  align-items: flex-start;
}

.mission-workspace__flow-stage-artifacts {
  display: grid;
  gap: 10px;
}

.mission-workspace__flow-artifact {
  display: grid;
  gap: 4px;
  padding: 10px 12px;
  border-radius: 16px;
  background: rgba(248, 250, 252, 0.94);
  border: 1px solid rgba(224, 229, 235, 0.92);
  text-align: left;
}

.mission-workspace__flow-artifact span {
  font-size: 0.72rem;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: rgb(108, 116, 130);
}

.mission-workspace__flow-artifact strong {
  font-size: 0.88rem;
  line-height: 1.4;
  color: rgb(31, 36, 43);
}

.mission-workspace__flow-arrow {
  color: rgb(117, 126, 139);
}

.mission-workspace__detail-meta {
  font-size: 0.84rem;
  color: rgb(102, 111, 124);
}

.mission-workspace__activity-item {
  display: grid;
  gap: 8px;
  padding: 16px;
}

@media (max-width: 1180px) {
  .mission-workspace__metrics,
  .mission-workspace__overview-grid,
  .mission-workspace__object-grid {
    grid-template-columns: minmax(0, 1fr);
  }
}

@media (max-width: 900px) {
  .mission-workspace__modebar,
  .mission-workspace__flow-toolbar {
    align-items: stretch;
    display: grid;
  }
}
</style>
