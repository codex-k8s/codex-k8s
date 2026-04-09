<template>
  <div class="mission-home">
    <section class="mission-home__composer">
      <div class="mission-home__composer-copy">
        <div class="mission-home__eyebrow">{{ projectTitle }}</div>
        <h2 class="mission-home__title">Что нужно запустить или решить?</h2>
        <p class="mission-home__summary">{{ projectSummary }}</p>
      </div>

      <div class="mission-home__composer-surface">
        <VTextarea
          v-model="draftPrompt"
          variant="outlined"
          rows="3"
          auto-grow
          hide-details
          placeholder="Опишите инициативу, задачу или желаемый результат. Можно в свободной форме."
        />

        <div class="mission-home__composer-actions">
          <VChip size="small" variant="tonal" color="primary">Полный цикл</VChip>
          <VChip size="small" variant="tonal" color="info">Короткая поставка</VChip>
          <VChip size="small" variant="tonal" color="warning">Горячее исправление</VChip>
          <VSpacer />
          <VBtn variant="tonal" prepend-icon="mdi-microphone" @click="$emit('open-voice')">Продиктовать</VBtn>
          <VBtn color="primary" prepend-icon="mdi-rocket-launch-outline" @click="$emit('launch-workflow')">
            Открыть workflow studio
          </VBtn>
        </div>
      </div>
    </section>

    <section class="mission-home__attention">
      <article
        v-for="card in attentionCards"
        :key="card.cardId"
        class="mission-home__attention-card"
        :class="`mission-home__attention-card--${card.tone}`"
      >
        <div class="mission-home__attention-label">{{ card.title }}</div>
        <div class="mission-home__attention-value">{{ card.valueLabel }}</div>
        <div class="mission-home__attention-summary">{{ card.summary }}</div>
      </article>
    </section>

    <section class="mission-home__board">
      <article v-for="column in columns" :key="column.columnId" class="mission-home__column">
        <div class="mission-home__column-head">
          <div>
            <div class="mission-home__column-title">{{ column.title }}</div>
            <div class="mission-home__column-summary">{{ column.summary }}</div>
          </div>
          <VChip size="small" variant="outlined">{{ column.items.length }}</VChip>
        </div>

        <div class="mission-home__column-items">
          <button
            v-for="item in column.items"
            :key="item.initiativeId"
            type="button"
            class="mission-home__initiative-card"
            @click="$emit('select-initiative', item.initiativeId)"
          >
            <div class="mission-home__initiative-topline">
              <VChip size="x-small" variant="tonal">{{ item.stageLabel }}</VChip>
              <VChip size="x-small" :color="toneColor(item.attentionTone)" variant="tonal">{{ item.attentionLabel }}</VChip>
            </div>

            <div class="mission-home__initiative-title">{{ item.title }}</div>
            <div class="mission-home__initiative-summary">{{ item.summary }}</div>

            <div class="mission-home__initiative-next">
              <span>Следующий шаг</span>
              <strong>{{ item.nextAction }}</strong>
            </div>

            <div class="mission-home__initiative-tags">
              <VChip v-for="tag in item.tags" :key="tag" size="x-small" variant="outlined">{{ tag }}</VChip>
            </div>

            <div class="mission-home__initiative-metrics">
              <span>Исполнений: {{ item.runSummary.total }}</span>
              <span>Активных: {{ item.runSummary.running }}</span>
              <span v-if="item.runSummary.failed > 0">Ошибок: {{ item.runSummary.failed }}</span>
            </div>
          </button>

          <div v-if="column.items.length === 0" class="mission-home__column-empty">
            По текущему проекту и фильтру здесь пока ничего нет.
          </div>
        </div>
      </article>
    </section>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";

import { missionAttentionToneColor } from "./presenters";
import type { MissionAttentionTone, MissionHomeAttentionCard, MissionHomeColumn } from "./types";

defineProps<{
  projectTitle: string;
  projectSummary: string;
  attentionCards: MissionHomeAttentionCard[];
  columns: MissionHomeColumn[];
}>();

defineEmits<{
  (event: "open-voice"): void;
  (event: "launch-workflow"): void;
  (event: "select-initiative", initiativeId: string): void;
}>();

const draftPrompt = ref("");

function toneColor(tone: MissionAttentionTone): string {
  return missionAttentionToneColor(tone);
}
</script>

<style scoped>
.mission-home {
  display: grid;
  gap: 20px;
}

.mission-home__composer {
  display: grid;
  gap: 16px;
  padding: 22px;
  border-radius: 28px;
  background:
    radial-gradient(circle at top left, rgba(255, 210, 128, 0.35), transparent 34%),
    linear-gradient(135deg, rgba(255, 249, 239, 0.98), rgba(255, 255, 255, 0.94));
  border: 1px solid rgba(214, 193, 145, 0.34);
  box-shadow: 0 22px 44px rgba(26, 29, 35, 0.06);
}

.mission-home__eyebrow {
  font-size: 0.78rem;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: rgb(141, 93, 31);
}

.mission-home__title {
  margin: 0;
  font-size: 1.65rem;
  line-height: 1.2;
  color: rgb(28, 33, 41);
}

.mission-home__summary {
  margin: 8px 0 0;
  max-width: 860px;
  font-size: 0.98rem;
  line-height: 1.55;
  color: rgb(82, 91, 107);
}

.mission-home__composer-surface {
  display: grid;
  gap: 14px;
  padding: 16px;
  border-radius: 22px;
  background: rgba(255, 255, 255, 0.9);
  border: 1px solid rgba(222, 226, 232, 0.82);
}

.mission-home__composer-actions {
  display: flex;
  gap: 10px;
  align-items: center;
  flex-wrap: wrap;
}

.mission-home__attention {
  display: grid;
  gap: 14px;
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.mission-home__attention-card {
  padding: 18px;
  border-radius: 22px;
  background: rgba(255, 255, 255, 0.88);
  border: 1px solid rgba(223, 228, 235, 0.86);
  box-shadow: 0 14px 28px rgba(26, 29, 35, 0.04);
}

.mission-home__attention-card--warning {
  background: linear-gradient(180deg, rgba(255, 246, 220, 0.96), rgba(255, 255, 255, 0.92));
}

.mission-home__attention-card--error {
  background: linear-gradient(180deg, rgba(255, 235, 232, 0.96), rgba(255, 255, 255, 0.92));
}

.mission-home__attention-card--success {
  background: linear-gradient(180deg, rgba(232, 250, 240, 0.96), rgba(255, 255, 255, 0.92));
}

.mission-home__attention-card--info {
  background: linear-gradient(180deg, rgba(236, 247, 255, 0.96), rgba(255, 255, 255, 0.92));
}

.mission-home__attention-label {
  font-size: 0.86rem;
  color: rgb(88, 97, 112);
}

.mission-home__attention-value {
  margin-top: 10px;
  font-size: 2rem;
  font-weight: 700;
  color: rgb(28, 33, 41);
}

.mission-home__attention-summary {
  margin-top: 8px;
  font-size: 0.9rem;
  line-height: 1.45;
  color: rgb(94, 102, 116);
}

.mission-home__board {
  display: grid;
  grid-template-columns: repeat(5, minmax(260px, 1fr));
  gap: 16px;
  align-items: start;
  overflow-x: auto;
  padding-bottom: 8px;
}

.mission-home__column {
  min-width: 260px;
  display: grid;
  gap: 14px;
  padding: 16px;
  border-radius: 24px;
  background: rgba(252, 252, 253, 0.86);
  border: 1px solid rgba(225, 229, 235, 0.82);
  box-shadow: 0 16px 34px rgba(26, 29, 35, 0.04);
}

.mission-home__column-head {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  align-items: flex-start;
}

.mission-home__column-title {
  font-size: 1rem;
  font-weight: 700;
  color: rgb(31, 36, 43);
}

.mission-home__column-summary {
  margin-top: 4px;
  font-size: 0.84rem;
  line-height: 1.45;
  color: rgb(103, 111, 124);
}

.mission-home__column-items {
  display: grid;
  gap: 12px;
}

.mission-home__initiative-card {
  display: grid;
  gap: 12px;
  width: 100%;
  padding: 16px;
  border: 1px solid rgba(224, 228, 235, 0.9);
  border-radius: 20px;
  background: white;
  text-align: left;
  transition:
    transform 0.18s ease,
    box-shadow 0.18s ease,
    border-color 0.18s ease;
}

.mission-home__initiative-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 14px 30px rgba(26, 29, 35, 0.08);
  border-color: rgba(156, 169, 191, 0.65);
}

.mission-home__initiative-topline {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.mission-home__initiative-title {
  font-size: 1rem;
  font-weight: 700;
  line-height: 1.35;
  color: rgb(28, 33, 41);
}

.mission-home__initiative-summary {
  font-size: 0.9rem;
  line-height: 1.5;
  color: rgb(86, 96, 112);
}

.mission-home__initiative-next {
  display: grid;
  gap: 4px;
  font-size: 0.85rem;
  color: rgb(103, 111, 124);
}

.mission-home__initiative-next strong {
  color: rgb(35, 41, 50);
}

.mission-home__initiative-tags,
.mission-home__initiative-metrics {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.mission-home__initiative-metrics {
  font-size: 0.8rem;
  color: rgb(107, 115, 129);
}

.mission-home__column-empty {
  padding: 18px 16px;
  border-radius: 18px;
  border: 1px dashed rgba(195, 201, 210, 0.9);
  font-size: 0.88rem;
  line-height: 1.5;
  color: rgb(105, 114, 128);
  background: rgba(255, 255, 255, 0.72);
}

@media (max-width: 1200px) {
  .mission-home__attention {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 720px) {
  .mission-home__attention {
    grid-template-columns: 1fr;
  }

  .mission-home__composer,
  .mission-home__column {
    padding: 16px;
  }
}
</style>
