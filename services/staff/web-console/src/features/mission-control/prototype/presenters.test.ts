import test from "node:test";
import assert from "node:assert/strict";

import { missionControlPrototypeModel } from "./fixtures.ts";
import {
  buildAttentionCards,
  buildExecutionGroups,
  buildHomeColumns,
  buildWorkflowStudioNodes,
  buildWorkspaceFlowNodes,
  buildWorkspaceStageViews,
} from "./presenters.ts";

test("buildHomeColumns раскладывает инициативы по верхнеуровневым колонкам", () => {
  const columns = buildHomeColumns(missionControlPrototypeModel, "project-kodex", "", "", "all");

  const designColumn = columns.find((column) => column.columnId === "design");
  const validationColumn = columns.find((column) => column.columnId === "validation");

  assert.ok(designColumn);
  assert.ok(validationColumn);
  assert.ok(designColumn.items.some((item) => item.initiativeId === "initiative-mission-control"));
  assert.ok(validationColumn.items.some((item) => item.initiativeId === "initiative-release-guard"));
});

test("buildAttentionCards считает предупреждения и блокеры по проекту", () => {
  const cards = buildAttentionCards(missionControlPrototypeModel, "project-kodex");

  const needsDecision = cards.find((card) => card.cardId === "needs-decision");
  const blocked = cards.find((card) => card.cardId === "blocked");

  assert.equal(needsDecision?.valueLabel, "1");
  assert.equal(blocked?.valueLabel, "2");
});

test("buildWorkspaceStageViews строит последовательность стадий по workflow инициативы", () => {
  const initiative = missionControlPrototypeModel.initiatives.find((item) => item.initiativeId === "initiative-mission-control") ?? null;
  const workflow = missionControlPrototypeModel.workflows.find((item) => item.workflowId === "workflow-owner-showcase") ?? null;
  const stageViews = buildWorkspaceStageViews(initiative, workflow);

  assert.equal(stageViews[0]?.stageKey, "vision");
  assert.equal(stageViews[1]?.status, "active");
  assert.match(stageViews[1]?.summary ?? "", /UX/i);
});

test("buildWorkspaceFlowNodes использует stage-centric модель вместо run-centric узлов", () => {
  const initiative = missionControlPrototypeModel.initiatives.find((item) => item.initiativeId === "initiative-hotfix-login") ?? null;
  const workflow = missionControlPrototypeModel.workflows.find((item) => item.workflowId === "workflow-hotfix") ?? null;
  const stageViews = buildWorkspaceStageViews(initiative, workflow);
  const nodes = buildWorkspaceFlowNodes(stageViews);

  assert.equal(nodes.length, workflow?.stages.length);
  assert.equal(nodes[0]?.kind, "stage");
  assert.equal(nodes[0]?.stageKey, "triage");
});

test("buildWorkflowStudioNodes добавляет gate nodes для design и qa", () => {
  const workflow = missionControlPrototypeModel.workflows.find((item) => item.workflowId === "workflow-owner-showcase") ?? null;
  const nodes = buildWorkflowStudioNodes(workflow);

  assert.ok(nodes.some((node) => node.nodeId === "studio-gate-design"));
  assert.ok(nodes.some((node) => node.nodeId === "studio-gate-qa"));
});

test("buildExecutionGroups группирует исполнения по артефактам", () => {
  const groups = buildExecutionGroups(missionControlPrototypeModel, "project-kodex", "");
  const missionControlGroup = groups.find((group) => group.groupId === "artifact-mc-prototype-task");

  assert.ok(missionControlGroup);
  assert.equal(missionControlGroup?.items.length, 1);
  assert.match(missionControlGroup?.initiativeTitle ?? "", /инициативами и агентами/i);
});
