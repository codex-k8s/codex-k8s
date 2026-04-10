export type MissionControlScreen = "home" | "initiative" | "studio" | "executions";

export type MissionInitiativeWorkspaceView = "overview" | "flow" | "artifacts" | "activity";

export type MissionHomeFilter = "all" | "needs-decision" | "blocked" | "release-ready";

export type MissionProjectAccent = "amber" | "teal" | "rose";

export type MissionWorkflowStageKey =
  | "intake"
  | "vision"
  | "prd"
  | "arch"
  | "design"
  | "plan"
  | "dev"
  | "qa"
  | "release"
  | "postdeploy"
  | "ops"
  | "triage"
  | "fix";

export type MissionWorkflowStageStatus = "pending" | "active" | "attention" | "blocked" | "done";

export type MissionAttentionTone = "info" | "success" | "warning" | "error";

export type MissionArtifactKind = "issue" | "pr";

export type MissionArtifactStatus = "draft" | "active" | "review" | "blocked" | "done";

export type MissionExecutionStatus = "running" | "waiting" | "failed" | "done";

export type MissionWorkflowTemplateKind = "system" | "project";

export type MissionControlPrototypeError = {
  messageKey: string;
  debugMessage?: string;
};

export type MissionControlPrototypeRouteState = {
  screen: MissionControlScreen;
  projectId: string;
  initiativeId: string;
  workflowId: string;
  artifactId: string;
  search: string;
  homeFilter: MissionHomeFilter;
  workspaceView: MissionInitiativeWorkspaceView;
};

export type MissionProject = {
  projectId: string;
  title: string;
  summary: string;
  accent: MissionProjectAccent;
};

export type MissionWorkflowStageDefinition = {
  stageKey: MissionWorkflowStageKey;
  label: string;
  summary: string;
  ownerLabel: string;
  outputLabel: string;
};

export type MissionWorkflowTemplate = {
  workflowId: string;
  title: string;
  summary: string;
  kind: MissionWorkflowTemplateKind;
  projectId?: string;
  stages: MissionWorkflowStageDefinition[];
  launchSummary: string;
  voiceHint: string;
  policyBullets: string[];
};

export type MissionRunSummary = {
  total: number;
  running: number;
  waiting: number;
  failed: number;
};

export type MissionArtifact = {
  artifactId: string;
  initiativeId: string;
  stageKey: MissionWorkflowStageKey;
  kind: MissionArtifactKind;
  title: string;
  summary: string;
  status: MissionArtifactStatus;
  ownerLabel: string;
  badgeLabels: string[];
  updatedAtLabel: string;
  runSummary: MissionRunSummary;
};

export type MissionInitiativeStageState = {
  stageKey: MissionWorkflowStageKey;
  status: MissionWorkflowStageStatus;
  summary: string;
  exitLabel: string;
  artifactIds: string[];
};

export type MissionInitiative = {
  initiativeId: string;
  projectId: string;
  workflowId: string;
  title: string;
  summary: string;
  ownerLabel: string;
  currentStageKey: MissionWorkflowStageKey;
  nextAction: string;
  statusLabel: string;
  attentionLabel: string;
  attentionTone: MissionAttentionTone;
  tags: string[];
  artifactIds: string[];
  stageStates: MissionInitiativeStageState[];
  blockedReason?: string;
  runSummary: MissionRunSummary;
};

export type MissionActivityItem = {
  itemId: string;
  initiativeId: string;
  title: string;
  summary: string;
  happenedAtLabel: string;
  actorLabel: string;
  targetKind: "issue" | "pr" | "run" | "stage";
  targetLabel: string;
  tone: MissionAttentionTone;
};

export type MissionExecution = {
  executionId: string;
  initiativeId: string;
  artifactId: string;
  title: string;
  summary: string;
  status: MissionExecutionStatus;
  agentRoleLabel: string;
  startedAtLabel: string;
  durationLabel: string;
};

export type MissionControlPrototypeModel = {
  projects: MissionProject[];
  workflows: MissionWorkflowTemplate[];
  initiatives: MissionInitiative[];
  artifacts: MissionArtifact[];
  activity: MissionActivityItem[];
  executions: MissionExecution[];
};

export type MissionProjectOption = {
  projectId: string;
  title: string;
};

export type MissionScreenOption = {
  screen: MissionControlScreen;
  label: string;
};

export type MissionWorkflowOption = {
  workflowId: string;
  title: string;
  kind: MissionWorkflowTemplateKind;
};

export type MissionHomeAttentionCard = {
  cardId: string;
  title: string;
  valueLabel: string;
  summary: string;
  tone: MissionAttentionTone;
  actionLabel: string;
};

export type MissionHomeInitiativeCard = {
  initiativeId: string;
  projectTitle: string;
  title: string;
  summary: string;
  stageLabel: string;
  nextAction: string;
  attentionLabel: string;
  attentionTone: MissionAttentionTone;
  primaryIssueTitle: string;
  primaryPrTitle: string;
  runSummary: MissionRunSummary;
  tags: string[];
};

export type MissionHomeColumn = {
  columnId: string;
  title: string;
  summary: string;
  items: MissionHomeInitiativeCard[];
};

export type MissionWorkspaceStageView = {
  stageKey: MissionWorkflowStageKey;
  label: string;
  summary: string;
  ownerLabel: string;
  outputLabel: string;
  status: MissionWorkflowStageStatus;
  exitLabel: string;
  artifactIds: string[];
};

export type MissionWorkspaceArtifactView = {
  artifactId: string;
  stageKey: MissionWorkflowStageKey;
  kind: MissionArtifactKind;
  title: string;
  summary: string;
  status: MissionArtifactStatus;
  ownerLabel: string;
  badgeLabels: string[];
  updatedAtLabel: string;
  runSummary: MissionRunSummary;
  selected: boolean;
};

export type MissionCanvasNodeKind = "stage" | "gate";

export type MissionCanvasNode = {
  nodeId: string;
  kind: MissionCanvasNodeKind;
  title: string;
  summary: string;
  statusLabel: string;
  tone: MissionAttentionTone;
  layoutX: number;
  layoutY: number;
  artifactIds: string[];
  stageKey?: MissionWorkflowStageKey;
};

export type MissionCanvasRelation = {
  relationId: string;
  sourceNodeId: string;
  targetNodeId: string;
  label: string;
};

export type MissionExecutionGroup = {
  groupId: string;
  initiativeTitle: string;
  artifactTitle: string;
  artifactKind: MissionArtifactKind;
  summary: string;
  items: MissionExecution[];
};
