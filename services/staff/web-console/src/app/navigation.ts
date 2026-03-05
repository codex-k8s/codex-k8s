export type NavGroupId = "operations" | "projects" | "platform";

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
  { id: "projects", titleKey: "nav.projectManagement" },
  { id: "platform", titleKey: "nav.platform" },
];

export const navItems: NavItem[] = [
  // Operations
  { groupId: "operations", routeName: "runs", titleKey: "nav.runs", icon: "mdi-play-circle-outline" },
  { groupId: "operations", routeName: "runtime-deploy-tasks", titleKey: "nav.runtimeDeployTasks", icon: "mdi-rocket-launch-outline" },
  { groupId: "operations", routeName: "wait-queue", titleKey: "nav.waitQueue", icon: "mdi-timer-sand" },
  { groupId: "operations", routeName: "approvals", titleKey: "nav.approvals", icon: "mdi-check-decagram-outline" },

  // Project management
  { groupId: "projects", routeName: "projects", titleKey: "nav.projects", icon: "mdi-folder-outline" },
  { groupId: "projects", routeName: "project-repositories", titleKey: "nav.repositories", icon: "mdi-source-repository", requiresProject: true, adminOnly: true },
  { groupId: "projects", routeName: "project-members", titleKey: "nav.members", icon: "mdi-account-group-outline", requiresProject: true, adminOnly: true },

  // Platform
  { groupId: "platform", routeName: "users", titleKey: "nav.users", icon: "mdi-account-multiple-outline", adminOnly: true },

  // Platform settings (scaffold)
  { groupId: "platform", routeName: "system-settings", titleKey: "nav.systemSettings", icon: "mdi-cog-outline", adminOnly: true },
];

export function findNavItemByRouteName(name: string | undefined): NavItem | undefined {
  if (!name) return undefined;
  return navItems.find((i) => i.routeName === name);
}
