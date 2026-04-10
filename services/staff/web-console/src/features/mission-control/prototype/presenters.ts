import type {
  MissionArtifact,
  MissionArtifactStatus,
  MissionAttentionTone,
  MissionCanvasNode,
  MissionCanvasRelation,
  MissionControlPrototypeModel,
  MissionExecutionGroup,
  MissionHomeAttentionCard,
  MissionHomeColumn,
  MissionHomeFilter,
  MissionHomeInitiativeCard,
  MissionInitiative,
  MissionProjectOption,
  MissionRunSummary,
  MissionWorkflowStageKey,
  MissionWorkflowTemplate,
  MissionWorkflowOption,
  MissionWorkflowStageStatus,
  MissionWorkspaceArtifactView,
  MissionWorkspaceStageView,
} from "./types";

const stageColumnCatalog: Array<{
  columnId: string;
  title: string;
  summary: string;
  stageKeys: MissionWorkflowStageKey[];
}> = [
  {
    columnId: "formation",
    title: "Формирование",
    summary: "Прием, видение и требования.",
    stageKeys: ["intake", "vision", "prd", "triage"] satisfies MissionWorkflowStageKey[],
  },
  {
    columnId: "design",
    title: "Проектирование",
    summary: "Архитектура, дизайн и план.",
    stageKeys: ["arch", "design", "plan"] satisfies MissionWorkflowStageKey[],
  },
  {
    columnId: "delivery",
    title: "Разработка",
    summary: "Код, сборка и подготовка PR.",
    stageKeys: ["dev", "fix"] satisfies MissionWorkflowStageKey[],
  },
  {
    columnId: "validation",
    title: "Проверка",
    summary: "QA, walkthrough и решение по рискам.",
    stageKeys: ["qa"] satisfies MissionWorkflowStageKey[],
  },
  {
    columnId: "release",
    title: "Релиз и сопровождение",
    summary: "Выкладка, postdeploy и ops.",
    stageKeys: ["release", "postdeploy", "ops"] satisfies MissionWorkflowStageKey[],
  },
];

function normalizeToken(value: string): string {
  return value.trim().toLowerCase();
}

function sumRunSummary(items: MissionRunSummary[]): MissionRunSummary {
  return items.reduce<MissionRunSummary>(
    (acc, item) => ({
      total: acc.total + item.total,
      running: acc.running + item.running,
      waiting: acc.waiting + item.waiting,
      failed: acc.failed + item.failed,
    }),
    { total: 0, running: 0, waiting: 0, failed: 0 },
  );
}

function initiativeMatchesSearch(initiative: MissionInitiative, artifacts: MissionArtifact[], search: string): boolean {
  const needle = normalizeToken(search);
  if (needle === "") {
    return true;
  }

  const artifactTokens = artifacts.flatMap((artifact) => [artifact.title, artifact.summary, ...artifact.badgeLabels]);
  return [initiative.title, initiative.summary, initiative.nextAction, ...initiative.tags, ...artifactTokens]
    .map(normalizeToken)
    .some((token) => token.includes(needle));
}

function stageToColumnId(stageKey: MissionWorkflowStageKey): string {
  return (
    stageColumnCatalog.find((column) => column.stageKeys.includes(stageKey))?.columnId ??
    stageColumnCatalog[stageColumnCatalog.length - 1].columnId
  );
}

function statusToTone(status: MissionWorkflowStageStatus): MissionAttentionTone {
  switch (status) {
    case "pending":
      return "info";
    case "active":
      return "warning";
    case "attention":
      return "warning";
    case "blocked":
      return "error";
    case "done":
      return "success";
  }
}

function stageArtifactPriority(status: MissionArtifactStatus): number {
  switch (status) {
    case "active":
      return 0;
    case "review":
      return 1;
    case "blocked":
      return 2;
    case "draft":
      return 3;
    case "done":
      return 4;
  }
}

function primaryArtifactForKind(
  initiative: MissionInitiative,
  artifacts: MissionArtifact[],
  kind: MissionArtifact["kind"],
): MissionArtifact | null {
  const currentStageCandidates = artifacts
    .filter((artifact) => artifact.kind === kind && artifact.stageKey === initiative.currentStageKey)
    .sort((left, right) => stageArtifactPriority(left.status) - stageArtifactPriority(right.status));

  if (currentStageCandidates.length > 0) {
    return currentStageCandidates[0];
  }

  const fallbackCandidates = artifacts
    .filter((artifact) => artifact.kind === kind)
    .sort((left, right) => stageArtifactPriority(left.status) - stageArtifactPriority(right.status));

  return fallbackCandidates[0] ?? null;
}

export function missionAttentionToneColor(tone: MissionAttentionTone): string {
  switch (tone) {
    case "info":
      return "info";
    case "success":
      return "success";
    case "warning":
      return "warning";
    case "error":
      return "error";
  }
}

export function missionArtifactStatusColor(status: MissionArtifactStatus): string {
  switch (status) {
    case "draft":
      return "secondary";
    case "active":
      return "info";
    case "review":
      return "warning";
    case "blocked":
      return "error";
    case "done":
      return "success";
  }
}

export function missionArtifactKindLabel(kind: MissionArtifact["kind"]): string {
  switch (kind) {
    case "issue":
      return "Issue";
    case "pr":
      return "PR";
  }
}

export function buildProjectOptions(model: MissionControlPrototypeModel | null): MissionProjectOption[] {
  if (!model) {
    return [];
  }

  return model.projects.map((project) => ({
    projectId: project.projectId,
    title: project.title,
  }));
}

export function buildWorkflowOptions(
  model: MissionControlPrototypeModel | null,
  projectId: string,
): MissionWorkflowOption[] {
  if (!model) {
    return [];
  }

  return model.workflows
    .filter((workflow) => workflow.kind === "system" || workflow.projectId === projectId)
    .map((workflow) => ({
      workflowId: workflow.workflowId,
      title: workflow.title,
      kind: workflow.kind,
    }));
}

export function buildAttentionCards(model: MissionControlPrototypeModel | null, projectId: string): MissionHomeAttentionCard[] {
  if (!model) {
    return [];
  }

  const projectInitiatives = model.initiatives.filter((initiative) => initiative.projectId === projectId);
  const projectExecutions = model.executions.filter((execution) =>
    projectInitiatives.some((initiative) => initiative.initiativeId === execution.initiativeId),
  );

  return [
    {
      cardId: "needs-decision",
      title: "Нуждаются в решении",
      valueLabel: String(projectInitiatives.filter((initiative) => initiative.attentionTone === "warning").length),
      summary: "Инициативы, где владелец должен принять решение или задать направление.",
      tone: "warning",
      actionLabel: "Открыть",
    },
    {
      cardId: "blocked",
      title: "Есть блокеры",
      valueLabel: String(
        projectInitiatives.filter((initiative) => initiative.attentionTone === "error" || initiative.blockedReason).length,
      ),
      summary: "Работа не может двигаться дальше без устранения блокера.",
      tone: "error",
      actionLabel: "Открыть",
    },
    {
      cardId: "active-runs",
      title: "Активные исполнения",
      valueLabel: String(projectExecutions.filter((execution) => execution.status === "running").length),
      summary: "Технические исполнения скрыты из основного потока, но живут за артефактами.",
      tone: "info",
      actionLabel: "К исполнению",
    },
    {
      cardId: "release-ready",
      title: "Почти готовы к выпуску",
      valueLabel: String(projectInitiatives.filter((initiative) => stageToColumnId(initiative.currentStageKey) === "release").length),
      summary: "Инициативы на релизе, postdeploy или ops.",
      tone: "success",
      actionLabel: "Открыть",
    },
  ];
}

function initiativeMatchesHomeFilter(initiative: MissionInitiative, homeFilter: MissionHomeFilter): boolean {
  switch (homeFilter) {
    case "all":
      return true;
    case "needs-decision":
      return initiative.attentionTone === "warning";
    case "blocked":
      return initiative.attentionTone === "error" || Boolean(initiative.blockedReason);
    case "release-ready":
      return stageToColumnId(initiative.currentStageKey) === "release";
  }
}

export function buildHomeColumns(
  model: MissionControlPrototypeModel | null,
  projectId: string,
  search: string,
  selectedInitiativeId: string,
  homeFilter: MissionHomeFilter,
): MissionHomeColumn[] {
  if (!model) {
    return [];
  }

  const projectTitle = model.projects.find((project) => project.projectId === projectId)?.title ?? "";

  return stageColumnCatalog.map((column) => {
    const items: MissionHomeInitiativeCard[] = model.initiatives
      .filter((initiative) => initiative.projectId === projectId)
      .filter((initiative) => (selectedInitiativeId === "" ? true : initiative.initiativeId === selectedInitiativeId))
      .filter((initiative) => initiativeMatchesHomeFilter(initiative, homeFilter))
      .filter((initiative) => stageToColumnId(initiative.currentStageKey) === column.columnId)
      .filter((initiative) => {
        const initiativeArtifacts = model.artifacts.filter((artifact) => artifact.initiativeId === initiative.initiativeId);
        return initiativeMatchesSearch(initiative, initiativeArtifacts, search);
      })
      .map((initiative) => {
        const workflow = model.workflows.find((candidate) => candidate.workflowId === initiative.workflowId);
        const stageLabel = workflow?.stages.find((stage) => stage.stageKey === initiative.currentStageKey)?.label ?? "Этап";
        const initiativeArtifacts = model.artifacts.filter((artifact) => artifact.initiativeId === initiative.initiativeId);
        const primaryIssue = primaryArtifactForKind(initiative, initiativeArtifacts, "issue");
        const primaryPr = primaryArtifactForKind(initiative, initiativeArtifacts, "pr");

        return {
          initiativeId: initiative.initiativeId,
          projectTitle,
          title: initiative.title,
          summary: initiative.summary,
          stageLabel,
          nextAction: initiative.nextAction,
          attentionLabel: initiative.attentionLabel,
          attentionTone: initiative.attentionTone,
          primaryIssueTitle: primaryIssue?.title ?? "Issue ещё не создан",
          primaryPrTitle: primaryPr?.title ?? "PR ещё не открыт",
          runSummary: initiative.runSummary,
          tags: initiative.tags,
        };
      });

    return {
      columnId: column.columnId,
      title: column.title,
      summary: column.summary,
      items,
    };
  }).filter((column) => column.items.length > 0);
}

export function buildWorkspaceStageViews(
  initiative: MissionInitiative | null,
  workflow: MissionWorkflowTemplate | null,
): MissionWorkspaceStageView[] {
  if (!initiative || !workflow) {
    return [];
  }

  const stageStateByKey = new Map(initiative.stageStates.map((stageState) => [stageState.stageKey, stageState]));

  return workflow.stages.map((stageDefinition) => {
    const state = stageStateByKey.get(stageDefinition.stageKey);
    return {
      stageKey: stageDefinition.stageKey,
      label: stageDefinition.label,
      summary: state?.summary ?? stageDefinition.summary,
      ownerLabel: stageDefinition.ownerLabel,
      outputLabel: stageDefinition.outputLabel,
      status: state?.status ?? "pending",
      exitLabel: state?.exitLabel ?? `Нужен артефакт: ${stageDefinition.outputLabel}`,
      artifactIds: state?.artifactIds ?? [],
    };
  });
}

export function buildWorkspaceArtifactViews(
  artifacts: MissionArtifact[],
  selectedArtifactId: string,
  search: string,
): MissionWorkspaceArtifactView[] {
  const needle = normalizeToken(search);

  return artifacts
    .filter((artifact) => {
      if (needle === "") {
        return true;
      }

      return [artifact.title, artifact.summary, ...artifact.badgeLabels]
        .map(normalizeToken)
        .some((token) => token.includes(needle));
    })
    .map((artifact) => ({
      artifactId: artifact.artifactId,
      stageKey: artifact.stageKey,
      kind: artifact.kind,
      title: artifact.title,
      summary: artifact.summary,
      status: artifact.status,
      ownerLabel: artifact.ownerLabel,
      badgeLabels: artifact.badgeLabels,
      updatedAtLabel: artifact.updatedAtLabel,
      runSummary: artifact.runSummary,
      selected: artifact.artifactId === selectedArtifactId,
    }));
}

export function buildWorkspaceFlowNodes(stageViews: MissionWorkspaceStageView[]): MissionCanvasNode[] {
  return stageViews.map((stageView, index) => ({
    nodeId: `stage-${stageView.stageKey}`,
    kind: "stage",
    title: stageView.label,
    summary: stageView.summary,
    statusLabel: stageView.exitLabel,
    tone: statusToTone(stageView.status),
    layoutX: 56 + index * 220,
    layoutY: index % 2 === 0 ? 150 : 110,
    artifactIds: stageView.artifactIds,
    stageKey: stageView.stageKey,
  }));
}

export function buildWorkspaceFlowRelations(stageViews: MissionWorkspaceStageView[]): MissionCanvasRelation[] {
  return stageViews.slice(1).map((stageView, index) => ({
    relationId: `relation-${stageViews[index].stageKey}-${stageView.stageKey}`,
    sourceNodeId: `stage-${stageViews[index].stageKey}`,
    targetNodeId: `stage-${stageView.stageKey}`,
    label: stageView.status === "blocked" ? "есть блокер на переходе" : "следующий этап",
  }));
}

export function buildWorkflowStudioNodes(workflow: MissionWorkflowTemplate | null): MissionCanvasNode[] {
  if (!workflow) {
    return [];
  }

  const stageNodes: MissionCanvasNode[] = workflow.stages.map((stageDefinition, index) => ({
    nodeId: `studio-stage-${stageDefinition.stageKey}`,
    kind: "stage",
    title: stageDefinition.label,
    summary: stageDefinition.summary,
    statusLabel: `Выход: ${stageDefinition.outputLabel}`,
    tone: index === 0 ? "warning" : "info",
    layoutX: 48 + index * 205,
    layoutY: index % 2 === 0 ? 160 : 100,
    artifactIds: [],
    stageKey: stageDefinition.stageKey,
  }));

  const gateNodes: MissionCanvasNode[] = workflow.stages
    .filter((stageDefinition) => stageDefinition.stageKey === "design" || stageDefinition.stageKey === "qa")
    .map((stageDefinition, index) => ({
      nodeId: `studio-gate-${stageDefinition.stageKey}`,
      kind: "gate",
      title: stageDefinition.stageKey === "design" ? "Owner review" : "Quality gate",
      summary: stageDefinition.stageKey === "design" ? "Владелец принимает структуру решения." : "Выпуск идет только после проверки.",
      statusLabel: "Gate node",
      tone: "warning",
      layoutX: index === 0 ? 430 : 850,
      layoutY: 310,
      artifactIds: [],
    }));

  return [...stageNodes, ...gateNodes];
}

export function buildWorkflowStudioRelations(workflow: MissionWorkflowTemplate | null): MissionCanvasRelation[] {
  if (!workflow) {
    return [];
  }

  const relations: MissionCanvasRelation[] = workflow.stages.slice(1).map((stageDefinition, index) => ({
    relationId: `studio-relation-${workflow.stages[index].stageKey}-${stageDefinition.stageKey}`,
    sourceNodeId: `studio-stage-${workflow.stages[index].stageKey}`,
    targetNodeId: `studio-stage-${stageDefinition.stageKey}`,
    label: "переход workflow",
  }));

  if (workflow.stages.some((stageDefinition) => stageDefinition.stageKey === "design")) {
    relations.push({
      relationId: "studio-design-gate",
      sourceNodeId: "studio-stage-design",
      targetNodeId: "studio-gate-design",
      label: "решение владельца",
    });
  }

  if (workflow.stages.some((stageDefinition) => stageDefinition.stageKey === "qa")) {
    relations.push({
      relationId: "studio-qa-gate",
      sourceNodeId: "studio-stage-qa",
      targetNodeId: "studio-gate-qa",
      label: "контроль качества",
    });
  }

  return relations;
}

export function buildExecutionGroups(
  model: MissionControlPrototypeModel | null,
  projectId: string,
  search: string,
): MissionExecutionGroup[] {
  if (!model) {
    return [];
  }

  const needle = normalizeToken(search);

  const groups = model.executions
    .filter((execution) =>
      model.initiatives.some(
        (initiative) => initiative.initiativeId === execution.initiativeId && initiative.projectId === projectId,
      ),
    )
    .reduce<Map<string, MissionExecutionGroup>>((acc, execution) => {
      const initiative = model.initiatives.find((candidate) => candidate.initiativeId === execution.initiativeId);
      const artifact = model.artifacts.find((candidate) => candidate.artifactId === execution.artifactId);
      if (!initiative || !artifact) {
        return acc;
      }

      const tokens = [initiative.title, artifact.title, execution.title, execution.summary].map(normalizeToken);
      if (needle !== "" && !tokens.some((token) => token.includes(needle))) {
        return acc;
      }

      const current = acc.get(execution.artifactId);
      if (current) {
        current.items.push(execution);
        return acc;
      }

      acc.set(execution.artifactId, {
        groupId: execution.artifactId,
        initiativeTitle: initiative.title,
        artifactTitle: artifact.title,
        artifactKind: artifact.kind,
        summary: artifact.summary,
        items: [execution],
      });

      return acc;
    }, new Map<string, MissionExecutionGroup>());

  return Array.from(groups.values()).sort((left, right) => left.initiativeTitle.localeCompare(right.initiativeTitle, "ru"));
}
