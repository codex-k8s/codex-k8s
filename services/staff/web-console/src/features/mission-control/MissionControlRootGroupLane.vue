<template>
  <section class="mission-root-group" :class="{ 'mission-root-group--blocking': rootGroup.has_blocking_gap }">
    <header class="mission-root-group__header">
      <div class="mission-root-group__headline">
        <div class="mission-root-group__eyebrow">
          <VChip size="x-small" variant="tonal" color="primary">
            {{ t(missionControlNodeKindLabelKey(rootGroup.root_node_kind)) }}
          </VChip>
          <VChip v-if="rootGroup.has_blocking_gap" size="x-small" variant="tonal" color="error">
            {{ t("pages.missionControl.blockingGap") }}
          </VChip>
        </div>
        <div class="mission-root-group__title">{{ rootGroup.root_title }}</div>
        <div class="mission-root-group__meta">
          <span class="mono">{{ rootGroup.root_node_public_id }}</span>
          <span>{{ t("pages.missionControl.groupNodes", { count: nodeCount }) }}</span>
        </div>
      </div>
      <div class="mission-root-group__header-side">
        <div class="text-caption text-medium-emphasis">{{ t("pages.missionControl.lastActivity") }}</div>
        <div class="mono">{{ formatCompactDateTime(rootGroup.latest_activity_at, locale) }}</div>
      </div>
    </header>

    <div class="mission-root-group__canvas">
      <div
        v-for="column in columns"
        :key="column.columnIndex"
        class="mission-root-group__column"
        :style="{ '--column-index': String(column.columnIndex) }"
      >
        <div class="mission-root-group__column-header">
          <div class="mission-root-group__column-title">{{ columnTitle(column) }}</div>
          <div class="text-caption text-medium-emphasis">{{ t("pages.missionControl.columnIndex", { value: column.columnIndex + 1 }) }}</div>
        </div>
        <div class="mission-root-group__stack">
          <MissionControlEntityCard
            v-for="node in column.nodes"
            :key="node.node_kind + ':' + node.node_public_id"
            :node="node"
            :selected="selectedNodeKey === node.node_kind + ':' + node.node_public_id"
            :locale="locale"
            @select="$emit('selectNode', { node_kind: node.node_kind, node_public_id: node.node_public_id })"
          />
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";

import { formatCompactDateTime } from "../../shared/lib/datetime";
import { missionControlGraphColumns } from "./lib";
import { missionControlNodeKindLabelKey } from "./presenters";
import type { MissionControlNode, MissionControlRootGroup, MissionControlSelectedNodeRef } from "./types";
import MissionControlEntityCard from "./MissionControlEntityCard.vue";

const props = defineProps<{
  rootGroup: MissionControlRootGroup;
  nodes: MissionControlNode[];
  selectedRef: MissionControlSelectedNodeRef | null;
  locale: string;
}>();

defineEmits<{
  selectNode: [ref: MissionControlSelectedNodeRef];
}>();

const { t } = useI18n({ useScope: "global" });

const nodesByKey = computed(() => {
  const index = new Map<string, MissionControlNode>();
  for (const node of props.nodes) {
    index.set(`${node.node_kind}:${node.node_public_id}`, node);
  }
  return index;
});

const columns = computed(() => missionControlGraphColumns(props.rootGroup.root_node_public_id, nodesByKey.value));
const nodeCount = computed(() => columns.value.reduce((total, column) => total + column.nodes.length, 0));
const selectedNodeKey = computed(() =>
  props.selectedRef ? `${props.selectedRef.node_kind}:${props.selectedRef.node_public_id}` : "",
);

function columnTitle(column: ReturnType<typeof missionControlGraphColumns>[number]): string {
  const keys = column.nodeKinds.map((kind) => missionControlNodeKindLabelKey(kind));
  return keys.map((key) => t(key)).join(" / ");
}
</script>

<style scoped>
.mission-root-group {
  border: 1px solid rgba(15, 23, 42, 0.08);
  border-radius: 30px;
  padding: 18px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(245, 248, 252, 0.98)),
    radial-gradient(300px 180px at 100% 0%, rgba(199, 233, 252, 0.36), transparent 60%);
}

.mission-root-group--blocking {
  border-color: rgba(220, 38, 38, 0.18);
}

.mission-root-group__header {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
}

.mission-root-group__headline {
  min-width: 0;
}

.mission-root-group__eyebrow {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.mission-root-group__title {
  margin-top: 10px;
  font-size: 1.1rem;
  font-weight: 800;
  color: rgb(15, 23, 42);
}

.mission-root-group__meta {
  margin-top: 8px;
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  color: rgba(15, 23, 42, 0.66);
}

.mission-root-group__header-side {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 4px;
  color: rgb(15, 23, 42);
}

.mission-root-group__canvas {
  margin-top: 18px;
  display: grid;
  gap: 16px;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
}

.mission-root-group__column {
  position: relative;
  border-radius: 24px;
  padding: 14px;
  background: rgba(255, 255, 255, 0.84);
  box-shadow: inset 0 0 0 1px rgba(15, 23, 42, 0.06);
}

.mission-root-group__column::before {
  content: "";
  position: absolute;
  top: 22px;
  left: -8px;
  width: 16px;
  height: 2px;
  background: linear-gradient(90deg, rgba(15, 23, 42, 0.08), rgba(14, 116, 144, 0.3));
}

.mission-root-group__column:first-child::before {
  display: none;
}

.mission-root-group__column-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 8px;
  margin-bottom: 12px;
}

.mission-root-group__column-title {
  font-size: 0.92rem;
  font-weight: 700;
  color: rgb(15, 23, 42);
}

.mission-root-group__stack {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}

@media (max-width: 960px) {
  .mission-root-group__header-side {
    align-items: flex-start;
  }

  .mission-root-group__canvas {
    grid-template-columns: 1fr;
  }

  .mission-root-group__column::before {
    display: none;
  }
}
</style>
