<template>
  <div v-if="workflow" class="mission-studio">
    <section class="mission-studio__toolbar">
      <div>
        <div class="mission-studio__eyebrow">Редактор workflow</div>
        <h2 class="mission-studio__title">{{ workflow.title }}</h2>
        <p class="mission-studio__summary">{{ workflow.summary }}</p>
      </div>

      <VSelect
        :model-value="workflow.workflowId"
        :items="workflowOptions"
        item-title="title"
        item-value="workflowId"
        label="Шаблон workflow"
        density="compact"
        variant="outlined"
        hide-details
        class="mission-studio__workflow-select"
        @update:model-value="onSelectWorkflow"
      />
    </section>

    <section class="mission-studio__grid">
      <article class="mission-studio__panel">
        <div class="mission-studio__panel-title">Что задает шаблон</div>
        <div class="mission-studio__block-list">
          <div class="mission-studio__block">
            <strong>Последовательность этапов</strong>
            <span>Определяет, какие Issue и PR должны появляться по очереди.</span>
          </div>
          <div class="mission-studio__block">
            <strong>Owner и quality gate</strong>
            <span>Говорит агенту, где нужно дождаться решения владельца или проверки качества.</span>
          </div>
          <div class="mission-studio__block">
            <strong>Follow-up правила</strong>
            <span>Описывает, какой следующий Issue должен быть создан после завершения этапа.</span>
          </div>
          <div class="mission-studio__block">
            <strong>Watermark и поля body</strong>
            <span>Напоминает, какие служебные поля должны появиться в GitHub-артефактах.</span>
          </div>
        </div>
      </article>

      <article class="mission-studio__panel mission-studio__panel--sequence">
        <div class="mission-studio__panel-title">Последовательность этапов</div>
        <div class="mission-studio__sequence">
          <button
            v-for="(stage, index) in workflow.stages"
            :key="stage.stageKey"
            type="button"
            class="mission-studio__stage-card"
            :class="{ 'mission-studio__stage-card--selected': selectedStageKey === stage.stageKey }"
            @click="selectedStageKey = stage.stageKey"
          >
            <div class="mission-studio__stage-index">Этап {{ index + 1 }}</div>
            <div class="mission-studio__stage-title">{{ stage.label }}</div>
            <div class="mission-studio__stage-summary">{{ stage.summary }}</div>
            <div class="mission-studio__stage-output">Выход: {{ stage.outputLabel }}</div>
            <div class="mission-studio__stage-badges">
              <VChip size="x-small" variant="tonal">{{ stage.ownerLabel }}</VChip>
              <VChip
                v-if="stage.stageKey === 'dev' || stage.stageKey === 'fix'"
                size="x-small"
                variant="tonal"
                color="primary"
              >
                PR этап
              </VChip>
              <VChip
                v-if="stage.stageKey === 'qa' || stage.stageKey === 'release' || stage.stageKey === 'postdeploy'"
                size="x-small"
                variant="tonal"
                color="warning"
              >
                Нужен follow-up
              </VChip>
            </div>
          </button>
        </div>

        <div class="mission-studio__preview-strip">
          <div v-for="stage in workflow.stages" :key="`preview-${stage.stageKey}`" class="mission-studio__preview-chip">
            {{ stage.label }}
          </div>
        </div>
      </article>

      <article class="mission-studio__panel">
        <div class="mission-studio__panel-title">Инспектор</div>

        <div v-if="selectedStage" class="mission-studio__inspector-section">
          <span>Выбранный этап</span>
          <strong>{{ selectedStage.label }}</strong>
          <p>{{ selectedStage.summary }}</p>
        </div>

        <div v-if="selectedStage" class="mission-studio__inspector-section">
          <span>Что агент должен создать</span>
          <strong>{{ expectedArtifactLabel(selectedStage.stageKey) }}</strong>
          <p>Артефакт создается через <code>gh</code>, а не напрямую из платформы.</p>
        </div>

        <div v-if="selectedStage" class="mission-studio__inspector-section">
          <span>Выход этапа</span>
          <strong>{{ selectedStage.outputLabel }}</strong>
          <p>После завершения этапа машина проверит watermark-блок и обязательные поля в body.</p>
        </div>

        <div class="mission-studio__inspector-section">
          <span>Запуск</span>
          <strong>{{ workflow.launchSummary }}</strong>
          <p>{{ workflow.voiceHint }}</p>
        </div>

        <div class="mission-studio__policy-list">
          <div v-for="bullet in workflow.policyBullets" :key="bullet" class="mission-studio__policy-item">
            {{ bullet }}
          </div>
        </div>
      </article>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";

import type { MissionCanvasNode, MissionCanvasRelation, MissionWorkflowOption, MissionWorkflowStageKey, MissionWorkflowTemplate } from "./types";

const props = defineProps<{
  workflow: MissionWorkflowTemplate | null;
  workflowOptions: MissionWorkflowOption[];
  nodes?: MissionCanvasNode[];
  relations?: MissionCanvasRelation[];
}>();

const emit = defineEmits<{
  (event: "select-workflow", workflowId: string): void;
}>();

const selectedStageKey = ref<MissionWorkflowStageKey | "">("");

const selectedStage = computed(
  () => props.workflow?.stages.find((stage) => stage.stageKey === selectedStageKey.value) ?? props.workflow?.stages[0] ?? null,
);

watch(
  () => props.workflow?.workflowId,
  () => {
    selectedStageKey.value = props.workflow?.stages[0]?.stageKey ?? "";
  },
  { immediate: true },
);

function onSelectWorkflow(nextWorkflowId: string | null): void {
  if (typeof nextWorkflowId === "string" && nextWorkflowId !== "") {
    emit("select-workflow", nextWorkflowId);
  }
}

function expectedArtifactLabel(stageKey: MissionWorkflowStageKey): string {
  if (stageKey === "dev" || stageKey === "fix") {
    return "Рабочий PR и, при необходимости, follow-up Issue";
  }

  if (stageKey === "qa" || stageKey === "release" || stageKey === "postdeploy") {
    return "Follow-up Issue следующего этапа";
  }

  return "Issue с зафиксированным результатом этапа";
}
</script>

<style scoped>
.mission-studio {
  display: grid;
  gap: 18px;
}

.mission-studio__toolbar,
.mission-studio__panel {
  border-radius: 26px;
  border: 1px solid rgba(223, 227, 233, 0.92);
  background: rgba(255, 255, 255, 0.94);
  box-shadow: 0 16px 34px rgba(26, 29, 35, 0.05);
}

.mission-studio__toolbar {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  align-items: flex-start;
  padding: 22px;
  background:
    linear-gradient(rgba(208, 213, 221, 0.22) 1px, transparent 1px),
    linear-gradient(90deg, rgba(208, 213, 221, 0.22) 1px, transparent 1px),
    linear-gradient(145deg, rgb(249, 250, 253), rgb(244, 246, 250));
  background-size: 32px 32px, 32px 32px, auto;
}

.mission-studio__workflow-select {
  width: 320px;
  max-width: 100%;
}

.mission-studio__eyebrow,
.mission-studio__panel-title,
.mission-studio__stage-index,
.mission-studio__inspector-section span {
  font-size: 0.78rem;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  font-weight: 700;
  color: rgb(105, 91, 39);
}

.mission-studio__title {
  margin: 8px 0 0;
  font-size: 1.55rem;
  line-height: 1.2;
  color: rgb(31, 36, 43);
}

.mission-studio__summary,
.mission-studio__block span,
.mission-studio__stage-summary,
.mission-studio__stage-output,
.mission-studio__inspector-section p,
.mission-studio__policy-item {
  margin: 8px 0 0;
  font-size: 0.9rem;
  line-height: 1.55;
  color: rgb(96, 104, 118);
}

.mission-studio__grid {
  display: grid;
  gap: 16px;
  grid-template-columns: 280px minmax(0, 1fr) 320px;
}

.mission-studio__panel {
  display: grid;
  gap: 14px;
  padding: 18px;
}

.mission-studio__panel--sequence {
  background:
    linear-gradient(rgba(208, 213, 221, 0.18) 1px, transparent 1px),
    linear-gradient(90deg, rgba(208, 213, 221, 0.18) 1px, transparent 1px),
    rgba(250, 251, 253, 0.94);
  background-size: 28px 28px;
}

.mission-studio__block-list,
.mission-studio__policy-list,
.mission-studio__sequence {
  display: grid;
  gap: 12px;
}

.mission-studio__block,
.mission-studio__policy-item,
.mission-studio__inspector-section {
  padding: 12px 14px;
  border-radius: 18px;
  background: rgba(248, 250, 252, 0.94);
  border: 1px solid rgba(224, 229, 235, 0.92);
}

.mission-studio__block strong,
.mission-studio__inspector-section strong,
.mission-studio__stage-title {
  display: block;
  font-size: 0.98rem;
  line-height: 1.4;
  color: rgb(31, 36, 43);
}

.mission-studio__stage-card {
  display: grid;
  gap: 8px;
  padding: 16px;
  border-radius: 20px;
  text-align: left;
  background: rgba(255, 255, 255, 0.95);
  border: 1px solid rgba(223, 227, 233, 0.94);
}

.mission-studio__stage-card--selected {
  border-color: rgba(82, 124, 255, 0.48);
  box-shadow: 0 0 0 1px rgba(82, 124, 255, 0.24);
}

.mission-studio__stage-badges,
.mission-studio__preview-strip {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.mission-studio__preview-strip {
  padding-top: 6px;
}

.mission-studio__preview-chip {
  padding: 6px 10px;
  border-radius: 999px;
  background: rgba(233, 239, 255, 0.98);
  font-size: 0.82rem;
  color: rgb(55, 72, 123);
}

@media (max-width: 1260px) {
  .mission-studio__grid {
    grid-template-columns: minmax(0, 1fr);
  }
}

@media (max-width: 900px) {
  .mission-studio__toolbar {
    display: grid;
  }

  .mission-studio__workflow-select {
    width: 100%;
  }
}
</style>
