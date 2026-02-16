import type { RouteRecordRaw } from "vue-router";

import ApprovalsCenterPage from "../pages/operations/ApprovalsCenterPage.vue";
import RegistryImagesPage from "../pages/operations/RegistryImagesPage.vue";
import RunningJobsPage from "../pages/operations/RunningJobsPage.vue";
import RuntimeDeployTaskDetailsPage from "../pages/operations/RuntimeDeployTaskDetailsPage.vue";
import RuntimeDeployTasksPage from "../pages/operations/RuntimeDeployTasksPage.vue";
import WaitQueuePage from "../pages/operations/WaitQueuePage.vue";

import ConfigMapsPage from "../pages/admin/ConfigMapsPage.vue";
import DeploymentsPage from "../pages/admin/DeploymentsPage.vue";
import JobsPage from "../pages/admin/JobsPage.vue";
import NamespacesPage from "../pages/admin/NamespacesPage.vue";
import PodsPage from "../pages/admin/PodsPage.vue";
import PvcPage from "../pages/admin/PvcPage.vue";
import ResourceDetailsPage from "../pages/admin/ResourceDetailsPage.vue";
import SecretsPage from "../pages/admin/SecretsPage.vue";

import AuditLogPage from "../pages/governance/AuditLogPage.vue";
import LabelsStagesPage from "../pages/governance/LabelsStagesPage.vue";

import AgentsPage from "../pages/configuration/AgentsPage.vue";
import AgentDetailsPage from "../pages/configuration/AgentDetailsPage.vue";
import DocsKnowledgePage from "../pages/configuration/DocsKnowledgePage.vue";
import McpToolsPage from "../pages/configuration/McpToolsPage.vue";
import ProjectMembersPage from "../pages/ProjectMembersPage.vue";
import ProjectDetailsPage from "../pages/ProjectDetailsPage.vue";
import ProjectRepositoriesPage from "../pages/ProjectRepositoriesPage.vue";
import ProjectsPage from "../pages/ProjectsPage.vue";
import RunDetailsPage from "../pages/RunDetailsPage.vue";
import RunsPage from "../pages/RunsPage.vue";
import SystemSettingsPage from "../pages/configuration/SystemSettingsPage.vue";
import UsersPage from "../pages/UsersPage.vue";

export const routes: RouteRecordRaw[] = [
  { path: "/", name: "projects", component: ProjectsPage, meta: { section: "projects" } },
  { path: "/projects/:projectId", name: "project-details", component: ProjectDetailsPage, props: true, meta: { section: "projects", crumbKey: "crumb.projectDetails" } },
  {
    path: "/projects/:projectId/repositories",
    name: "project-repositories",
    component: ProjectRepositoriesPage,
    props: true,
    meta: { adminOnly: true, section: "projects", crumbKey: "crumb.projectRepositories" },
  },
  {
    path: "/projects/:projectId/members",
    name: "project-members",
    component: ProjectMembersPage,
    props: true,
    meta: { adminOnly: true, section: "projects", crumbKey: "crumb.projectMembers" },
  },
  { path: "/runs", name: "runs", component: RunsPage, meta: { section: "runs" } },
  { path: "/runs/:runId", name: "run-details", component: RunDetailsPage, props: true, meta: { section: "runs", crumbKey: "crumb.runDetails" } },

  // Operations
  { path: "/runtime-deploy/tasks", name: "runtime-deploy-tasks", component: RuntimeDeployTasksPage, meta: { section: "operations", crumbKey: "crumb.runtimeDeployTasks" } },
  { path: "/runtime-deploy/tasks/:runId", name: "runtime-deploy-task-details", component: RuntimeDeployTaskDetailsPage, props: true, meta: { section: "operations", crumbKey: "crumb.runtimeDeployTaskDetails" } },
  { path: "/runtime-deploy/images", name: "runtime-deploy-images", component: RegistryImagesPage, meta: { section: "operations", crumbKey: "crumb.registryImages" } },
  { path: "/running-jobs", name: "running-jobs", component: RunningJobsPage, meta: { section: "operations", crumbKey: "crumb.runningJobs" } },
  { path: "/wait-queue", name: "wait-queue", component: WaitQueuePage, meta: { section: "operations", crumbKey: "crumb.waitQueue" } },
  { path: "/approvals", name: "approvals", component: ApprovalsCenterPage, meta: { section: "operations", crumbKey: "crumb.approvals" } },

  // Governance (scaffold)
  { path: "/governance/audit-log", name: "audit-log", component: AuditLogPage, meta: { section: "governance", crumbKey: "crumb.auditLog" } },
  { path: "/governance/labels-stages", name: "labels-stages", component: LabelsStagesPage, meta: { section: "governance", crumbKey: "crumb.labelsStages" } },

  // Admin / Cluster (scaffold)
  { path: "/admin/namespaces", name: "cluster-namespaces", component: NamespacesPage, meta: { section: "admin", crumbKey: "crumb.clusterNamespaces" } },
  {
    path: "/admin/namespaces/:name",
    name: "cluster-namespaces-details",
    component: ResourceDetailsPage,
    props: (r) => ({ kind: "namespaces", name: typeof r.params.name === "string" ? r.params.name : "" }),
    meta: { section: "admin", crumbKey: "crumb.clusterNamespaces" },
  },
  { path: "/admin/configmaps", name: "cluster-configmaps", component: ConfigMapsPage, meta: { section: "admin", crumbKey: "crumb.clusterConfigMaps" } },
  {
    path: "/admin/configmaps/:name",
    name: "cluster-configmaps-details",
    component: ResourceDetailsPage,
    props: (r) => ({ kind: "configmaps", name: typeof r.params.name === "string" ? r.params.name : "" }),
    meta: { section: "admin", crumbKey: "crumb.clusterConfigMaps" },
  },
  { path: "/admin/secrets", name: "cluster-secrets", component: SecretsPage, meta: { section: "admin", crumbKey: "crumb.clusterSecrets" } },
  {
    path: "/admin/secrets/:name",
    name: "cluster-secrets-details",
    component: ResourceDetailsPage,
    props: (r) => ({ kind: "secrets", name: typeof r.params.name === "string" ? r.params.name : "" }),
    meta: { section: "admin", crumbKey: "crumb.clusterSecrets" },
  },
  { path: "/admin/deployments", name: "cluster-deployments", component: DeploymentsPage, meta: { section: "admin", crumbKey: "crumb.clusterDeployments" } },
  {
    path: "/admin/deployments/:name",
    name: "cluster-deployments-details",
    component: ResourceDetailsPage,
    props: (r) => ({ kind: "deployments", name: typeof r.params.name === "string" ? r.params.name : "" }),
    meta: { section: "admin", crumbKey: "crumb.clusterDeployments" },
  },
  { path: "/admin/pods", name: "cluster-pods", component: PodsPage, meta: { section: "admin", crumbKey: "crumb.clusterPods" } },
  {
    path: "/admin/pods/:name",
    name: "cluster-pods-details",
    component: ResourceDetailsPage,
    props: (r) => ({ kind: "pods", name: typeof r.params.name === "string" ? r.params.name : "" }),
    meta: { section: "admin", crumbKey: "crumb.clusterPods" },
  },
  { path: "/admin/jobs", name: "cluster-jobs", component: JobsPage, meta: { section: "admin", crumbKey: "crumb.clusterJobs" } },
  {
    path: "/admin/jobs/:name",
    name: "cluster-jobs-details",
    component: ResourceDetailsPage,
    props: (r) => ({ kind: "jobs", name: typeof r.params.name === "string" ? r.params.name : "" }),
    meta: { section: "admin", crumbKey: "crumb.clusterJobs" },
  },
  { path: "/admin/pvc", name: "cluster-pvc", component: PvcPage, meta: { section: "admin", crumbKey: "crumb.clusterPvc" } },
  {
    path: "/admin/pvc/:name",
    name: "cluster-pvc-details",
    component: ResourceDetailsPage,
    props: (r) => ({ kind: "pvc", name: typeof r.params.name === "string" ? r.params.name : "" }),
    meta: { section: "admin", crumbKey: "crumb.clusterPvc" },
  },

  // Configuration (scaffold)
  { path: "/configuration/agents", name: "agents", component: AgentsPage, meta: { section: "configuration", crumbKey: "crumb.agents" } },
  {
    path: "/configuration/agents/:agentName",
    name: "agent-details",
    component: AgentDetailsPage,
    props: (r) => ({ agentName: typeof r.params.agentName === "string" ? r.params.agentName : "" }),
    meta: { section: "configuration", crumbKey: "crumb.agents" },
  },
  { path: "/configuration/system-settings", name: "system-settings", component: SystemSettingsPage, meta: { section: "configuration", crumbKey: "crumb.systemSettings" } },
  { path: "/configuration/docs", name: "docs-knowledge", component: DocsKnowledgePage, meta: { section: "configuration", crumbKey: "crumb.docs" } },
  { path: "/configuration/mcp-tools", name: "mcp-tools", component: McpToolsPage, meta: { section: "configuration", crumbKey: "crumb.mcpTools" } },

  { path: "/users", name: "users", component: UsersPage, meta: { adminOnly: true, section: "users" } },
];
