<template>
  <div class="mission-home">
    <section class="mission-home__hero">
      <div class="mission-home__eyebrow">{{ projectTitle }}</div>
      <h2 class="mission-home__title">Что требует решения по проекту?</h2>
      <p class="mission-home__summary">{{ projectSummary }}</p>
      <div class="mission-home__hero-note">
        Здесь показываются только живые инициативы проекта. Создание новой работы вынесено в отдельное меню, а голосовой
        запуск остается в плавающей кнопке справа снизу.
      </div>
    </section>

    <section v-if="selectedInitiativeTitle || selectedFilterLabel" class="mission-home__focus">
      <div class="mission-home__focus-copy">
        <div v-if="selectedInitiativeTitle" class="mission-home__focus-item">
          <span>Фокус по инициативе</span>
          <strong>{{ selectedInitiativeTitle }}</strong>
        </div>
        <div v-if="selectedFilterLabel" class="mission-home__focus-item">
          <span>Быстрый фильтр</span>
          <strong>{{ selectedFilterLabel }}</strong>
        </div>
      </div>

      <div class="mission-home__focus-actions">
        <VBtn
          v-if="selectedInitiativeTitle"
          size="small"
          variant="text"
          prepend-icon="mdi-close"
          @click="$emit('clear-initiative')"
        >
          Сбросить инициативу
        </VBtn>
        <VBtn
          v-if="selectedFilterLabel"
          size="small"
          variant="text"
          prepend-icon="mdi-filter-off-outline"
          @click="$emit('clear-filter')"
        >
          Сбросить фильтр
        </VBtn>
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
        <div class="mission-home__attention-actions">
          <VBtn size="small" variant="tonal" :color="toneColor(card.tone)" @click="$emit('open-attention', card.cardId)">
            {{ card.actionLabel }}
          </VBtn>
        </div>
      </article>
    </section>

    <section v-if="columns.length > 0" class="mission-home__board">
      <article v-for="column in columns" :key="column.columnId" class="mission-home__column">
        <div class="mission-home__column-head">
          <div>
            <div class="mission-home__column-title">{{ column.title }}</div>
            <div class="mission-home__column-summary">{{ column.summary }}</div>
          </div>
          <VChip size="small" variant="outlined">{{ column.items.length }}</VChip>
        </div>

        <div class="mission-home__column-items">
          <article
            v-for="item in column.items"
            :key="item.initiativeId"
            class="mission-home__initiative-card"
          >
            <div class="mission-home__initiative-topline">
              <VChip size="x-small" variant="tonal">{{ item.stageLabel }}</VChip>
              <VChip size="x-small" :color="toneColor(item.attentionTone)" variant="tonal">{{ item.attentionLabel }}</VChip>
            </div>

            <div class="mission-home__initiative-title">{{ item.title }}</div>
            <div class="mission-home__initiative-summary">{{ item.summary }}</div>

            <div class="mission-home__initiative-objects">
              <div class="mission-home__object-line">
                <span>Текущий issue</span>
                <strong>{{ item.primaryIssueTitle }}</strong>
              </div>
              <div class="mission-home__object-line">
                <span>Текущий PR</span>
                <strong>{{ item.primaryPrTitle }}</strong>
              </div>
            </div>

            <div class="mission-home__initiative-next">
              <span>Что сделать сейчас</span>
              <strong>{{ item.nextAction }}</strong>
            </div>

            <div class="mission-home__initiative-metrics">
              <span>Всего исполнений: {{ item.runSummary.total }}</span>
              <span v-if="item.runSummary.running > 0">Идут: {{ item.runSummary.running }}</span>
              <span v-if="item.runSummary.failed > 0">Ошибки: {{ item.runSummary.failed }}</span>
            </div>

            <div class="mission-home__initiative-actions">
              <VBtn size="small" variant="text" @click="$emit('select-initiative', item.initiativeId)">Фокус</VBtn>
              <VBtn size="small" color="primary" variant="tonal" @click="$emit('open-workspace', item.initiativeId)">
                Открыть инициативу
              </VBtn>
            </div>
          </article>
        </div>
      </article>
    </section>

    <section v-else class="mission-home__empty">
      <div class="mission-home__empty-title">По текущему фильтру ничего не найдено</div>
      <div class="mission-home__empty-summary">
        Сбросьте поиск или фильтр и выберите другую инициативу из кнопки в верхней панели.
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { missionAttentionToneColor } from "./presenters";
import type { MissionAttentionTone, MissionHomeAttentionCard, MissionHomeColumn } from "./types";

defineProps<{
  projectTitle: string;
  projectSummary: string;
  attentionCards: MissionHomeAttentionCard[];
  columns: MissionHomeColumn[];
  selectedInitiativeTitle: string;
  selectedFilterLabel: string;
}>();

defineEmits<{
  (event: "open-attention", cardId: string): void;
  (event: "select-initiative", initiativeId: string): void;
  (event: "open-workspace", initiativeId: string): void;
  (event: "clear-initiative"): void;
  (event: "clear-filter"): void;
}>();

function toneColor(tone: MissionAttentionTone): string {
  return missionAttentionToneColor(tone);
}
</script>

<style scoped>
.mission-home {
  display: grid;
  gap: 20px;
}

.mission-home__hero,
.mission-home__focus,
.mission-home__empty,
.mission-home__column,
.mission-home__attention-card {
  border-radius: 24px;
  border: 1px solid rgba(223, 228, 235, 0.92);
  background: rgba(255, 255, 255, 0.92);
  box-shadow: 0 16px 34px rgba(26, 29, 35, 0.05);
}

.mission-home__hero {
  display: grid;
  gap: 10px;
  padding: 22px;
  background:
    radial-gradient(circle at top left, rgba(255, 216, 142, 0.2), transparent 30%),
    linear-gradient(140deg, rgba(255, 251, 242, 0.98), rgba(250, 251, 255, 0.94));
}

.mission-home__eyebrow {
  font-size: 0.78rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: rgb(138, 91, 28);
}

.mission-home__title {
  margin: 0;
  font-size: 1.55rem;
  line-height: 1.2;
  color: rgb(30, 35, 43);
}

.mission-home__summary,
.mission-home__hero-note,
.mission-home__attention-summary,
.mission-home__column-summary,
.mission-home__initiative-summary,
.mission-home__initiative-metrics,
.mission-home__empty-summary {
  font-size: 0.92rem;
  line-height: 1.55;
  color: rgb(96, 104, 118);
}

.mission-home__hero-note {
  max-width: 980px;
}

.mission-home__focus {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
  padding: 16px 18px;
}

.mission-home__focus-copy {
  display: flex;
  gap: 18px;
  flex-wrap: wrap;
}

.mission-home__focus-item {
  display: grid;
  gap: 4px;
}

.mission-home__focus-item span {
  font-size: 0.78rem;
  color: rgb(102, 111, 124);
}

.mission-home__focus-item strong {
  font-size: 0.95rem;
  color: rgb(30, 35, 43);
}

.mission-home__focus-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.mission-home__attention {
  display: grid;
  gap: 14px;
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.mission-home__attention-card {
  display: grid;
  gap: 10px;
  padding: 18px;
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
  font-size: 1.9rem;
  font-weight: 700;
  color: rgb(28, 33, 41);
}

.mission-home__attention-actions {
  display: flex;
  justify-content: flex-start;
}

.mission-home__board {
  display: grid;
  gap: 16px;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.mission-home__column {
  display: grid;
  gap: 14px;
  padding: 18px;
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

.mission-home__column-items {
  display: grid;
  gap: 12px;
}

.mission-home__initiative-card {
  display: grid;
  gap: 12px;
  padding: 16px;
  border-radius: 20px;
  background: rgba(248, 250, 252, 0.94);
  border: 1px solid rgba(224, 229, 235, 0.92);
}

.mission-home__initiative-topline {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.mission-home__initiative-title {
  font-size: 1rem;
  font-weight: 700;
  color: rgb(30, 35, 43);
}

.mission-home__initiative-objects {
  display: grid;
  gap: 10px;
}

.mission-home__object-line,
.mission-home__initiative-next {
  display: grid;
  gap: 4px;
}

.mission-home__object-line span,
.mission-home__initiative-next span {
  font-size: 0.76rem;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: rgb(111, 119, 133);
}

.mission-home__object-line strong,
.mission-home__initiative-next strong {
  font-size: 0.92rem;
  line-height: 1.45;
  color: rgb(31, 36, 43);
}

.mission-home__initiative-metrics {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.mission-home__initiative-actions {
  display: flex;
  justify-content: space-between;
  gap: 10px;
  align-items: center;
}

.mission-home__empty {
  padding: 24px;
}

.mission-home__empty-title {
  font-size: 1rem;
  font-weight: 700;
  color: rgb(31, 36, 43);
}

@media (max-width: 1260px) {
  .mission-home__attention,
  .mission-home__board {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 900px) {
  .mission-home__attention,
  .mission-home__board {
    grid-template-columns: minmax(0, 1fr);
  }

  .mission-home__focus,
  .mission-home__initiative-actions {
    grid-template-columns: minmax(0, 1fr);
    display: grid;
  }
}
</style>
