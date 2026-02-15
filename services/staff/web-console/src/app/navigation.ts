export type NavGroupId = "operations" | "platform" | "governance" | "admin" | "configuration";

export type NavGroup = {
  id: NavGroupId;
  titleKey: string;
};

export type NavItem = {
  groupId: NavGroupId;
  routeName: string;
  titleKey: string;
  icon: string;
  comingSoon?: boolean;
  adminOnly?: boolean;
  requiresProject?: boolean;
};

export const navGroups: NavGroup[] = [
  { id: "operations", titleKey: "nav.operations" },
  { id: "platform", titleKey: "nav.platform" },
  { id: "governance", titleKey: "nav.governance" },
  { id: "admin", titleKey: "nav.adminCluster" },
  { id: "configuration", titleKey: "nav.configuration" },
];

export const navItems: NavItem[] = [
  // Operations
  { groupId: "operations", routeName: "runs", titleKey: "nav.runs", icon: "mdi-play-circle-outline" },
  { groupId: "operations", routeName: "running-jobs", titleKey: "nav.runningJobs", icon: "mdi-server-outline" },
  { groupId: "operations", routeName: "wait-queue", titleKey: "nav.waitQueue", icon: "mdi-timer-sand" },
  { groupId: "operations", routeName: "approvals", titleKey: "nav.approvals", icon: "mdi-check-decagram-outline" },

  // Platform
  { groupId: "platform", routeName: "projects", titleKey: "nav.projects", icon: "mdi-folder-outline" },
  { groupId: "platform", routeName: "project-details", titleKey: "nav.projectDetails", icon: "mdi-folder-information-outline", requiresProject: true },
  { groupId: "platform", routeName: "project-repositories", titleKey: "nav.repositories", icon: "mdi-source-repository", requiresProject: true, adminOnly: true },
  { groupId: "platform", routeName: "project-members", titleKey: "nav.members", icon: "mdi-account-group-outline", requiresProject: true, adminOnly: true },
  { groupId: "platform", routeName: "users", titleKey: "nav.users", icon: "mdi-account-multiple-outline", adminOnly: true },

  // Governance (scaffold)
  { groupId: "governance", routeName: "audit-log", titleKey: "nav.auditLog", icon: "mdi-history", comingSoon: true },
  { groupId: "governance", routeName: "labels-stages", titleKey: "nav.labelsStages", icon: "mdi-tag-multiple-outline", comingSoon: true },

  // Admin / Cluster (scaffold)
  { groupId: "admin", routeName: "cluster-namespaces", titleKey: "nav.cluster.namespaces", icon: "mdi-home-group", comingSoon: true },
  { groupId: "admin", routeName: "cluster-configmaps", titleKey: "nav.cluster.configMaps", icon: "mdi-file-cog-outline", comingSoon: true },
  { groupId: "admin", routeName: "cluster-secrets", titleKey: "nav.cluster.secrets", icon: "mdi-key-outline", comingSoon: true },
  { groupId: "admin", routeName: "cluster-deployments", titleKey: "nav.cluster.deployments", icon: "mdi-rocket-launch-outline", comingSoon: true },
  { groupId: "admin", routeName: "cluster-pods", titleKey: "nav.cluster.pods", icon: "mdi-cube-outline", comingSoon: true },
  { groupId: "admin", routeName: "cluster-jobs", titleKey: "nav.cluster.jobs", icon: "mdi-briefcase-outline", comingSoon: true },
  { groupId: "admin", routeName: "cluster-pvc", titleKey: "nav.cluster.pvc", icon: "mdi-database-outline", comingSoon: true },

  // Configuration (scaffold)
  { groupId: "configuration", routeName: "agents", titleKey: "nav.agents", icon: "mdi-robot-outline", comingSoon: true },
  { groupId: "configuration", routeName: "system-settings", titleKey: "nav.systemSettings", icon: "mdi-cog-outline", comingSoon: true },
  { groupId: "configuration", routeName: "docs-knowledge", titleKey: "nav.docs", icon: "mdi-book-open-page-variant-outline", comingSoon: true },
  { groupId: "configuration", routeName: "mcp-tools", titleKey: "nav.mcpTools", icon: "mdi-wrench-cog-outline", comingSoon: true },
];

export function findNavItemByRouteName(name: string | undefined): NavItem | undefined {
  if (!name) return undefined;
  return navItems.find((i) => i.routeName === name);
}
