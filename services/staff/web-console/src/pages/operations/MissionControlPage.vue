<template>
  <div class="mission-control-page">
    <PageHeader :title="t('pages.missionControl.title')" :hint="t('pages.missionControl.hint')">
      <template #actions>
        <AdaptiveBtn
          variant="tonal"
          icon="mdi-refresh"
          :label="t('common.refresh')"
          :loading="missionControl.refreshing"
          @click="refreshWorkspace"
        />
      </template>
    </PageHeader>

    <VAlert v-if="missionControl.error" type="error" variant="tonal" class="mt-4">
      {{ t(missionControl.error.messageKey) }}
    </VAlert>

    <VAlert v-if="realtimeAlert" :type="realtimeAlert.type" variant="tonal" class="mt-4">
      <div class="d-flex flex-column ga-2">
        <div class="font-weight-medium">{{ t(realtimeAlert.titleKey) }}</div>
        <div class="text-body-2">{{ realtimeAlert.text }}</div>
        <div class="d-flex flex-wrap ga-2">
          <AdaptiveBtn
            variant="text"
            icon="mdi-refresh"
            :label="t('pages.missionControl.refreshNow')"
            :loading="missionControl.refreshing"
            @click="refreshWorkspace"
          />
          <AdaptiveBtn
            v-if="routeState.viewMode !== 'list'"
            variant="text"
            icon="mdi-format-list-bulleted"
            :label="t('pages.missionControl.listProjection')"
            @click="updateRoute({ viewMode: 'list' })"
          />
        </div>
      </div>
    </VAlert>

    <div class="mission-control-summary mt-4">
      <VCard v-for="item in summaryCards" :key="item.key" class="mission-control-summary__card" variant="outlined">
        <VCardText>
          <div class="mission-control-summary__label">{{ t(item.labelKey) }}</div>
          <div class="mission-control-summary__value">{{ item.value }}</div>
          <div class="mission-control-summary__hint">{{ item.hint }}</div>
        </VCardText>
      </VCard>
    </div>

    <VCard class="mt-4 mission-control-shell" variant="outlined">
      <div class="mission-control-shell__toolbar">
        <div class="mission-control-shell__filters">
          <VChip size="small" variant="outlined">
            {{ t("pages.missionControl.fixedFilters.openScope") }}
          </VChip>
          <VChip size="small" variant="outlined">
            {{ t("pages.missionControl.fixedFilters.assignmentScope") }}
          </VChip>
        </div>

        <VSelect
          :model-value="routeState.statePreset"
          class="mission-control-shell__filter"
          density="compact"
          variant="solo-filled"
          hide-details
          :label="t('pages.missionControl.filterLabel')"
          :items="filterOptions"
          @update:model-value="onFilterChange"
        />

        <VTextField
          v-model="searchInput"
          class="mission-control-shell__search"
          density="compact"
          variant="solo-filled"
          hide-details
          clearable
          :label="t('pages.missionControl.searchLabel')"
          prepend-inner-icon="mdi-magnify"
          @click:clear="onSearchClear"
          @keyup.enter="applySearch"
        />

        <VBtnToggle
          :model-value="routeState.viewMode"
          class="mission-control-shell__toggle"
          color="primary"
          density="comfortable"
          mandatory
        >
          <VBtn value="graph" @click="updateRoute({ viewMode: 'graph' })">
            <VIcon icon="mdi-graph-outline" start />
            {{ t("pages.missionControl.viewModes.graph") }}
          </VBtn>
          <VBtn value="list" @click="updateRoute({ viewMode: 'list' })">
            <VIcon icon="mdi-format-list-bulleted" start />
            {{ t("pages.missionControl.viewModes.list") }}
          </VBtn>
        </VBtnToggle>

        <div class="mission-control-shell__status">
          <VChip size="small" variant="tonal" :color="freshnessChip.color">
            {{ t("pages.missionControl.freshness") }}: {{ t(freshnessChip.labelKey) }}
          </VChip>
          <VChip size="small" variant="tonal" :color="realtimeChip.color">
            {{ t("pages.missionControl.realtime") }}: {{ t(realtimeChip.labelKey) }}
          </VChip>
        </div>
      </div>

      <div class="mission-control-shell__body">
        <div class="mission-control-shell__canvas">
          <div class="mission-control-shell__canvas-head">
            <div>
              <div class="text-overline">{{ t("pages.missionControl.primaryWorkspace") }}</div>
              <div class="text-h6">{{ t(canvasTitleKey) }}</div>
              <div class="text-body-2 text-medium-emphasis mt-1">
                {{
                  t("pages.missionControl.rootsVisible", {
                    count: missionControl.snapshot?.root_groups.length ?? 0,
                    total: missionControl.snapshot?.summary.root_count ?? 0,
                  })
                }}
              </div>
            </div>
          </div>

          <div v-if="workspaceWatermarks.length" class="mission-control-shell__watermarks">
            <VChip
              v-for="watermark in workspaceWatermarks"
              :key="watermark.watermark_kind + ':' + watermark.observed_at"
              size="small"
              variant="tonal"
              :color="missionControlWatermarkColor(watermark.status)"
            >
              {{ t(missionControlWatermarkLabelKey(watermark.watermark_kind)) }} · {{ watermark.summary }}
            </VChip>
          </div>

          <div v-if="missionControl.loading" class="mission-control-loading">
            <VSkeletonLoader type="article, article, article" />
          </div>

          <template v-else-if="hasWorkspaceContent">
            <div v-if="routeState.viewMode === 'graph'" class="mission-control-graph">
              <MissionControlRootGroupLane
                v-for="rootGroup in rootGroups"
                :key="rootGroup.root_node_kind + ':' + rootGroup.root_node_public_id"
                :root-group="rootGroup"
                :nodes="workspaceNodes"
                :selected-ref="missionControl.selectedRef"
                :locale="locale"
                @select-node="openNodeRef"
              />
            </div>

            <div v-else class="mission-control-list">
              <VTable density="comfortable" fixed-header>
                <thead>
                  <tr>
                    <th>{{ t("pages.missionControl.columns.kind") }}</th>
                    <th>{{ t("pages.missionControl.columns.title") }}</th>
                    <th>{{ t("pages.missionControl.columns.state") }}</th>
                    <th>{{ t("pages.missionControl.columns.continuity") }}</th>
                    <th>{{ t("pages.missionControl.columns.visibility") }}</th>
                    <th>{{ t("pages.missionControl.columns.root") }}</th>
                    <th>{{ t("pages.missionControl.columns.updated") }}</th>
                  </tr>
                </thead>
                <tbody>
                  <tr
                    v-for="node in listNodes"
                    :key="node.node_kind + ':' + node.node_public_id"
                    class="mission-control-list__row"
                    :class="{ 'mission-control-list__row--selected': selectedNodeKey === node.node_kind + ':' + node.node_public_id }"
                    @click="openNodeRef({ node_kind: node.node_kind, node_public_id: node.node_public_id })"
                  >
                    <td>
                      <VChip size="x-small" variant="tonal" :color="missionControlStateColor(node.active_state)">
                        {{ t(missionControlNodeKindLabelKey(node.node_kind)) }}
                      </VChip>
                    </td>
                    <td>
                      <div class="font-weight-medium">{{ node.title }}</div>
                      <div class="text-body-2 text-medium-emphasis mono">
                        {{ node.provider_reference?.external_id || node.node_public_id }}
                      </div>
                    </td>
                    <td>{{ t(`pages.missionControl.states.${node.active_state}`) }}</td>
                    <td>{{ t(`pages.missionControl.continuity.${node.continuity_status}`) }}</td>
                    <td>{{ t(`pages.missionControl.visibility.${node.visibility_tier}`) }}</td>
                    <td class="mono">{{ node.root_node_public_id }}</td>
                    <td class="mono">{{ formatCompactDateTime(node.last_activity_at, locale) }}</td>
                  </tr>
                </tbody>
              </VTable>
            </div>

            <div v-if="missionControl.hasMore" class="mt-4 d-flex justify-center">
              <AdaptiveBtn
                variant="text"
                icon="mdi-chevron-down"
                :label="t('pages.missionControl.loadMore')"
                :loading="missionControl.loadingMore"
                @click="missionControl.loadSnapshot({ append: true })"
              />
            </div>
          </template>

          <div v-else class="mission-control-empty">
            <VIcon icon="mdi-radar" size="48" class="mb-4 text-medium-emphasis" />
            <div class="text-h6">{{ t("pages.missionControl.emptyTitle") }}</div>
            <div class="text-body-2 text-medium-emphasis mt-2">
              {{ t("pages.missionControl.emptyText") }}
            </div>
          </div>
        </div>

        <div v-if="missionControl.selectedRef && !display.mobile.value" class="mission-control-shell__drawer">
          <MissionControlSidePanel
            :details="missionControl.selectedDetails"
            :loading="missionControl.selectedLoading"
            :error="missionControl.selectedDetailsError"
            :activity="missionControl.selectedActivity"
            :activity-error="missionControl.selectedActivityError"
            :activity-loading="missionControl.selectedActivityLoading"
            :has-more-activity="missionControl.hasSelectedActivityMore"
            :locale="locale"
            @close="closeNode"
            @select-node="openNodeRef"
            @load-more-activity="missionControl.loadSelectedActivity({ append: true })"
            @open-preview="openPreview"
          />
        </div>
      </div>
    </VCard>

    <VDialog v-model="mobilePanelOpen" fullscreen transition="dialog-bottom-transition">
      <VCard>
        <MissionControlSidePanel
          :details="missionControl.selectedDetails"
          :loading="missionControl.selectedLoading"
          :error="missionControl.selectedDetailsError"
          :activity="missionControl.selectedActivity"
          :activity-error="missionControl.selectedActivityError"
          :activity-loading="missionControl.selectedActivityLoading"
          :has-more-activity="missionControl.hasSelectedActivityMore"
          :locale="locale"
          @close="closeNode"
          @select-node="openNodeRef"
          @load-more-activity="missionControl.loadSelectedActivity({ append: true })"
          @open-preview="openPreview"
        />
      </VCard>
    </VDialog>

    <MissionControlLaunchPreviewDialog
      :open="previewDialogOpen"
      :loading="previewLoading"
      :error="previewError"
      :preview="previewResult"
      :command-template="previewCommandTemplate"
      :node-title="previewNodeTitle"
      :known-gaps="missionControl.selectedDetails?.continuity_gaps ?? []"
      @close="closePreview"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useI18n } from "vue-i18n";
import { useDisplay } from "vuetify";

import { normalizeApiError, type ApiError } from "../../shared/api/errors";
import { formatCompactDateTime } from "../../shared/lib/datetime";
import { bindRealtimePageLifecycle } from "../../shared/ws/lifecycle";
import { previewMissionControlLaunch } from "../../features/mission-control/api";
import MissionControlEntityCard from "../../features/mission-control/MissionControlEntityCard.vue";
import MissionControlLaunchPreviewDialog from "../../features/mission-control/MissionControlLaunchPreviewDialog.vue";
import MissionControlRootGroupLane from "../../features/mission-control/MissionControlRootGroupLane.vue";
import MissionControlSidePanel from "../../features/mission-control/MissionControlSidePanel.vue";
import { buildMissionControlRouteQuery, missionControlRouteStateEquals, normalizeMissionControlRouteQuery } from "../../features/mission-control/lib";
import {
  missionControlNodeKindLabelKey,
  missionControlStateColor,
  missionControlWatermarkColor,
  missionControlWatermarkLabelKey,
} from "../../features/mission-control/presenters";
import { subscribeMissionControlRealtime } from "../../features/mission-control/realtime";
import { useMissionControlStore } from "../../features/mission-control/store";
import type {
  MissionControlLaunchPreview,
  MissionControlLaunchSurface,
  MissionControlRealtimeNotice,
  MissionControlRouteState,
  MissionControlSelectedNodeRef,
  MissionControlStageNextStepTemplate,
  MissionControlWorkspaceFreshnessStatus,
} from "../../features/mission-control/types";
import AdaptiveBtn from "../../shared/ui/AdaptiveBtn.vue";
import PageHeader from "../../shared/ui/PageHeader.vue";

const route = useRoute();
const router = useRouter();
const display = useDisplay();
const missionControl = useMissionControlStore();
const { t, locale } = useI18n({ useScope: "global" });

const mobilePanelOpen = ref(false);
const searchInput = ref("");
const searchDebounceTimer = ref<number | null>(null);
const realtimeRefreshTimer = ref<number | null>(null);
const stopRealtimeRef = ref<(() => void) | null>(null);
const stopLifecycleBindingRef = ref<(() => void) | null>(null);
const activeRouteState = ref<MissionControlRouteState | null>(null);

const previewDialogOpen = ref(false);
const previewLoading = ref(false);
const previewError = ref<ApiError | null>(null);
const previewResult = ref<MissionControlLaunchPreview | null>(null);
const previewCommandTemplate = ref<MissionControlStageNextStepTemplate | null>(null);
const previewNodeTitle = ref("");

const routeState = computed(() => normalizeMissionControlRouteQuery(route.query));
const workspaceWatermarks = computed(() => missionControl.snapshot?.workspace_watermarks ?? []);
const workspaceNodes = computed(() => missionControl.snapshot?.nodes ?? []);
const rootGroups = computed(() => missionControl.snapshot?.root_groups ?? []);
const hasWorkspaceContent = computed(() => rootGroups.value.length > 0 || workspaceNodes.value.length > 0);
const selectedNodeKey = computed(() =>
  missionControl.selectedRef ? `${missionControl.selectedRef.node_kind}:${missionControl.selectedRef.node_public_id}` : "",
);

const listNodes = computed(() =>
  [...workspaceNodes.value].sort((left, right) => {
    if ((left.last_activity_at || "") === (right.last_activity_at || "")) {
      return left.title.localeCompare(right.title);
    }
    return String(left.last_activity_at || "") < String(right.last_activity_at || "") ? 1 : -1;
  }),
);

const summaryCards = computed(() => {
  const summary = missionControl.snapshot?.summary;
  return [
    { key: "roots", labelKey: "pages.missionControl.summary.roots", value: summary?.root_count ?? 0, hint: t("pages.missionControl.summaryHints.roots") },
    { key: "nodes", labelKey: "pages.missionControl.summary.nodes", value: summary?.node_count ?? 0, hint: t("pages.missionControl.summaryHints.nodes") },
    {
      key: "blocking",
      labelKey: "pages.missionControl.summary.blocking_gaps",
      value: summary?.blocking_gap_count ?? 0,
      hint: t("pages.missionControl.summaryHints.blocking_gaps"),
    },
    {
      key: "warning",
      labelKey: "pages.missionControl.summary.warning_gaps",
      value: summary?.warning_gap_count ?? 0,
      hint: t("pages.missionControl.summaryHints.warning_gaps"),
    },
    {
      key: "recent_closed",
      labelKey: "pages.missionControl.summary.recent_closed_context",
      value: summary?.recent_closed_context_count ?? 0,
      hint: t("pages.missionControl.summaryHints.recent_closed_context"),
    },
    { key: "review", labelKey: "pages.missionControl.summary.review", value: summary?.review_count ?? 0, hint: t("pages.missionControl.summaryHints.review") },
  ];
});

const filterOptions = computed(() => [
  { title: t("pages.missionControl.filters.all_active"), value: "all_active" },
  { title: t("pages.missionControl.filters.working"), value: "working" },
  { title: t("pages.missionControl.filters.waiting"), value: "waiting" },
  { title: t("pages.missionControl.filters.blocked"), value: "blocked" },
  { title: t("pages.missionControl.filters.review"), value: "review" },
  { title: t("pages.missionControl.filters.recent_critical_updates"), value: "recent_critical_updates" },
]);

const freshnessChip = computed(() => {
  switch (missionControl.effectiveFreshnessStatus) {
    case "stale":
      return { color: "warning", labelKey: "pages.missionControl.freshnessStates.stale" };
    case "degraded":
      return { color: "error", labelKey: "pages.missionControl.freshnessStates.degraded" };
    case "fresh":
      return { color: "success", labelKey: "pages.missionControl.freshnessStates.fresh" };
    default:
      return { color: "secondary", labelKey: "pages.missionControl.freshnessStates.unknown" };
  }
});

const realtimeChip = computed(() => {
  switch (missionControl.realtimeState) {
    case "connected":
      return { color: "success", labelKey: "pages.missionControl.realtimeStates.connected" };
    case "reconnecting":
      return { color: "warning", labelKey: "pages.missionControl.realtimeStates.reconnecting" };
    case "connecting":
      return { color: "secondary", labelKey: "pages.missionControl.realtimeStates.connecting" };
    default:
      return { color: "secondary", labelKey: "pages.missionControl.realtimeStates.disconnected" };
  }
});

const canvasTitleKey = computed(() =>
  routeState.value.viewMode === "graph" ? "pages.missionControl.graphTitle" : "pages.missionControl.listTitle",
);

const realtimeAlert = computed(() => buildRealtimeAlert(missionControl.realtimeNotice, missionControl.effectiveFreshnessStatus));

watch(
  routeState,
  async (next) => {
    searchInput.value = next.search;
    const previous = activeRouteState.value;
    activeRouteState.value = next;

    const queryChanged =
      !previous ||
      next.viewMode !== previous.viewMode ||
      next.statePreset !== previous.statePreset ||
      next.search !== previous.search;
    if (queryChanged) {
      missionControl.configureQuery({
        viewMode: next.viewMode,
        statePreset: next.statePreset,
        search: next.search,
      });
      await missionControl.loadSnapshot();
      restartRealtime();
    }

    const nodeChanged =
      !previous ||
      next.nodeKind !== previous.nodeKind ||
      next.nodePublicId !== previous.nodePublicId;
    if (nodeChanged) {
      if (next.nodeKind && next.nodePublicId) {
        await missionControl.loadSelectedNode({
          node_kind: next.nodeKind,
          node_public_id: next.nodePublicId,
        });
        if (display.mobile.value) {
          mobilePanelOpen.value = true;
        }
      } else {
        missionControl.clearSelectedNode();
        mobilePanelOpen.value = false;
      }
    }
  },
  { deep: true, immediate: true },
);

watch(
  () => display.mobile.value,
  (isMobile) => {
    if (!isMobile) {
      mobilePanelOpen.value = false;
    }
  },
  { immediate: true },
);

watch(searchInput, () => {
  scheduleSearch();
});

function clearSearchDebounce(): void {
  if (searchDebounceTimer.value !== null) {
    window.clearTimeout(searchDebounceTimer.value);
    searchDebounceTimer.value = null;
  }
}

function clearRealtimeRefreshTimer(): void {
  if (realtimeRefreshTimer.value !== null) {
    window.clearTimeout(realtimeRefreshTimer.value);
    realtimeRefreshTimer.value = null;
  }
}

function scheduleSearch(): void {
  clearSearchDebounce();
  searchDebounceTimer.value = window.setTimeout(() => {
    applySearch();
    searchDebounceTimer.value = null;
  }, 350);
}

function scheduleRealtimeRefresh(delayMs = 500): void {
  if (realtimeRefreshTimer.value !== null) {
    return;
  }
  realtimeRefreshTimer.value = window.setTimeout(() => {
    realtimeRefreshTimer.value = null;
    void refreshWorkspace();
  }, delayMs);
}

function updateRoute(patch: Partial<MissionControlRouteState>): void {
  const nextState: MissionControlRouteState = {
    ...routeState.value,
    ...patch,
  };

  if ("viewMode" in patch || "statePreset" in patch || "search" in patch) {
    nextState.nodeKind = "";
    nextState.nodePublicId = "";
  }

  if (missionControlRouteStateEquals(nextState, routeState.value)) {
    return;
  }

  void router.replace({
    name: "mission-control",
    query: buildMissionControlRouteQuery(nextState),
  });
}

function applySearch(): void {
  const trimmed = searchInput.value.trim();
  if (trimmed === routeState.value.search) {
    return;
  }
  updateRoute({ search: trimmed });
}

function onSearchClear(): void {
  searchInput.value = "";
  updateRoute({ search: "" });
}

function onFilterChange(value: string): void {
  if (value === routeState.value.statePreset) {
    return;
  }
  updateRoute({
    statePreset: value as MissionControlRouteState["statePreset"],
  });
}

function openNodeRef(ref: MissionControlSelectedNodeRef): void {
  updateRoute({
    nodeKind: ref.node_kind,
    nodePublicId: ref.node_public_id,
  });
}

function closeNode(): void {
  mobilePanelOpen.value = false;
  updateRoute({
    nodeKind: "",
    nodePublicId: "",
  });
}

function stopRealtime(): void {
  stopRealtimeRef.value?.();
  stopRealtimeRef.value = null;
  missionControl.setRealtimeState("closed");
}

function restartRealtime(): void {
  stopRealtime();
  const resumeToken = String(missionControl.snapshot?.resume_token || "").trim();
  if (resumeToken === "") {
    return;
  }

  missionControl.setRealtimeState("connecting");
  stopRealtimeRef.value = subscribeMissionControlRealtime({
    resumeToken,
    onMessage: (message) => {
      missionControl.setRealtimeState("connected");
      if (
        message.event_kind === "connected" &&
        (message.payload.snapshot_freshness_status === "fresh" || missionControl.realtimeNotice?.kind === "error")
      ) {
        missionControl.clearRealtimeNotice();
      }

      switch (message.event_kind) {
        case "connected":
          missionControl.setConnectedFreshnessStatus(message.payload.snapshot_freshness_status);
          return;
        case "delta":
          scheduleRealtimeRefresh();
          return;
        case "invalidate":
          missionControl.applyRealtimeNotice({
            kind: "invalidate",
            reason: message.payload.reason,
            refreshScope: message.payload.refresh_scope,
            affectedCount: message.payload.affected_count,
            occurredAt: message.occurred_at,
          });
          scheduleRealtimeRefresh(200);
          return;
        case "stale":
          missionControl.applyRealtimeNotice({
            kind: "stale",
            reason: message.payload.reason,
            staleSince: message.payload.stale_since,
            suggestedRefresh: message.payload.suggested_refresh,
            occurredAt: message.occurred_at,
          });
          return;
        case "degraded":
          missionControl.applyRealtimeNotice({
            kind: "degraded",
            reason: message.payload.reason,
            fallbackMode: message.payload.fallback_mode,
            affectedCapabilities: message.payload.affected_capabilities,
            occurredAt: message.occurred_at,
          });
          return;
        case "resync_required":
          missionControl.applyRealtimeNotice({
            kind: "resync_required",
            reason: message.payload.reason,
            requiredSnapshotId: message.payload.required_snapshot_id,
            droppedEventCount: message.payload.dropped_event_count,
            occurredAt: message.occurred_at,
          });
          scheduleRealtimeRefresh(0);
          return;
        case "error":
          missionControl.applyRealtimeNotice({
            kind: "error",
            code: message.payload.code,
            message: message.payload.message,
            retryable: message.payload.retryable,
            occurredAt: message.occurred_at,
          });
          return;
        case "heartbeat":
          return;
      }
    },
    onStateChange: (state) => {
      missionControl.setRealtimeState(state);
    },
    onInitialMessageTimeout: () => {
      missionControl.applyRealtimeNotice({
        kind: "error",
        code: "realtime_timeout",
        message: "realtime_initial_timeout",
        retryable: true,
        occurredAt: new Date().toISOString(),
      });
    },
  });
}

async function refreshWorkspace(): Promise<void> {
  clearRealtimeRefreshTimer();
  await missionControl.refreshSnapshot();
  restartRealtime();
}

async function openPreview(surface: MissionControlLaunchSurface): Promise<void> {
  const details = missionControl.selectedDetails;
  const template = surface.command_template;
  if (!details || !template) {
    return;
  }

  previewDialogOpen.value = true;
  previewLoading.value = true;
  previewError.value = null;
  previewResult.value = null;
  previewCommandTemplate.value = template;
  previewNodeTitle.value = details.node.title;

  try {
    previewResult.value = await previewMissionControlLaunch({
      nodeKind: details.node.node_kind,
      nodePublicId: details.node.node_public_id,
      threadKind: template.thread_kind,
      threadNumber: template.thread_number,
      targetLabel: template.target_label,
      removedLabels: template.removed_labels,
      expectedProjectionVersion: details.node.projection_version,
    });
  } catch (error) {
    previewError.value = normalizeApiError(error);
  } finally {
    previewLoading.value = false;
  }
}

function closePreview(): void {
  previewDialogOpen.value = false;
  previewLoading.value = false;
  previewError.value = null;
  previewResult.value = null;
  previewCommandTemplate.value = null;
  previewNodeTitle.value = "";
}

function buildRealtimeAlert(
  notice: MissionControlRealtimeNotice | null,
  freshnessStatus: MissionControlWorkspaceFreshnessStatus,
): { type: "info" | "warning" | "error"; titleKey: string; text: string } | null {
  if (!notice) {
    if (freshnessStatus === "stale") {
      return {
        type: "warning",
        titleKey: "pages.missionControl.alerts.staleTitle",
        text: t("pages.missionControl.alerts.snapshotStaleText"),
      };
    }
    if (freshnessStatus === "degraded") {
      return {
        type: "error",
        titleKey: "pages.missionControl.alerts.degradedTitle",
        text: t("pages.missionControl.alerts.snapshotDegradedText"),
      };
    }
    return null;
  }

  switch (notice.kind) {
    case "invalidate":
      return {
        type: "info",
        titleKey: "pages.missionControl.alerts.invalidateTitle",
        text: t("pages.missionControl.alerts.invalidateText", {
          reason: notice.reason,
          scope: notice.refreshScope,
          count: notice.affectedCount,
        }),
      };
    case "stale":
      return {
        type: "warning",
        titleKey: "pages.missionControl.alerts.staleTitle",
        text: t("pages.missionControl.alerts.staleText", {
          reason: notice.reason,
          since: formatCompactDateTime(notice.staleSince, locale.value),
        }),
      };
    case "degraded":
      return {
        type: "error",
        titleKey: "pages.missionControl.alerts.degradedTitle",
        text: t("pages.missionControl.alerts.degradedText", {
          reason: notice.reason,
          fallback: notice.fallbackMode,
          capabilities: notice.affectedCapabilities.join(", "),
        }),
      };
    case "resync_required":
      return {
        type: "warning",
        titleKey: "pages.missionControl.alerts.resyncTitle",
        text: t("pages.missionControl.alerts.resyncText", {
          reason: notice.reason,
          dropped: notice.droppedEventCount,
        }),
      };
    case "error":
      return {
        type: "error",
        titleKey: "pages.missionControl.alerts.errorTitle",
        text: t("pages.missionControl.alerts.errorText", {
          code: notice.code,
          message: notice.message,
        }),
      };
  }
}

stopLifecycleBindingRef.value = bindRealtimePageLifecycle({
  onResume: () => {
    void refreshWorkspace();
  },
  onSuspend: () => {
    stopRealtime();
  },
});

onBeforeUnmount(() => {
  clearSearchDebounce();
  clearRealtimeRefreshTimer();
  stopRealtime();
  stopLifecycleBindingRef.value?.();
  stopLifecycleBindingRef.value = null;
});
</script>

<style scoped>
.mission-control-summary {
  display: grid;
  gap: 14px;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
}

.mission-control-summary__card {
  border-radius: 24px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(246, 248, 252, 0.98)),
    radial-gradient(260px 150px at 100% 0%, rgba(214, 238, 255, 0.6), transparent 60%);
}

.mission-control-summary__label {
  font-size: 0.82rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: rgba(15, 23, 42, 0.56);
}

.mission-control-summary__value {
  margin-top: 10px;
  font-size: 2rem;
  font-weight: 800;
  line-height: 1;
  color: rgb(15, 23, 42);
}

.mission-control-summary__hint {
  margin-top: 8px;
  color: rgba(15, 23, 42, 0.66);
}

.mission-control-shell {
  border-radius: 32px;
  overflow: hidden;
}

.mission-control-shell__toolbar {
  display: flex;
  gap: 12px;
  align-items: center;
  flex-wrap: wrap;
  padding: 18px 20px;
  border-bottom: 1px solid rgba(15, 23, 42, 0.08);
  background:
    linear-gradient(180deg, rgba(250, 252, 255, 0.98), rgba(242, 247, 250, 0.96)),
    radial-gradient(240px 160px at 0% 0%, rgba(251, 191, 36, 0.14), transparent 60%);
}

.mission-control-shell__filters {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.mission-control-shell__filter {
  max-width: 280px;
}

.mission-control-shell__search {
  min-width: 280px;
  flex: 1 1 320px;
}

.mission-control-shell__toggle :deep(.v-btn) {
  text-transform: none;
}

.mission-control-shell__status {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  margin-left: auto;
}

.mission-control-shell__body {
  display: grid;
  grid-template-columns: minmax(0, 1.7fr) minmax(340px, 0.9fr);
  min-height: 720px;
}

.mission-control-shell__canvas {
  padding: 22px;
  background:
    linear-gradient(180deg, rgba(245, 247, 250, 0.98), rgba(255, 255, 255, 0.98)),
    radial-gradient(360px 220px at 100% 0%, rgba(186, 230, 253, 0.22), transparent 65%);
}

.mission-control-shell__canvas-head {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
}

.mission-control-shell__watermarks {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 16px;
}

.mission-control-shell__drawer {
  padding: 18px;
  border-left: 1px solid rgba(15, 23, 42, 0.08);
  background: rgba(249, 250, 251, 0.88);
}

.mission-control-loading {
  padding: 20px 0;
}

.mission-control-graph {
  display: flex;
  flex-direction: column;
  gap: 16px;
  margin-top: 18px;
}

.mission-control-list {
  margin-top: 18px;
  overflow-x: auto;
}

.mission-control-list__row {
  cursor: pointer;
}

.mission-control-list__row--selected {
  background: rgba(8, 145, 178, 0.08);
}

.mission-control-empty {
  min-height: 420px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
}

.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}

@media (max-width: 1280px) {
  .mission-control-shell__body {
    grid-template-columns: 1fr;
  }

  .mission-control-shell__drawer {
    display: none;
  }
}

@media (max-width: 960px) {
  .mission-control-shell__toolbar,
  .mission-control-shell__canvas {
    padding: 16px;
  }

  .mission-control-shell__status {
    margin-left: 0;
  }

  .mission-control-shell__search {
    min-width: 0;
    flex-basis: 100%;
  }
}
</style>
