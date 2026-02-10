import type { RouteRecordRaw } from "vue-router";

import ProjectMembersPage from "../pages/ProjectMembersPage.vue";
import ProjectDetailsPage from "../pages/ProjectDetailsPage.vue";
import ProjectRepositoriesPage from "../pages/ProjectRepositoriesPage.vue";
import ProjectsPage from "../pages/ProjectsPage.vue";
import RunDetailsPage from "../pages/RunDetailsPage.vue";
import RunsPage from "../pages/RunsPage.vue";
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
  { path: "/users", name: "users", component: UsersPage, meta: { adminOnly: true, section: "users" } },
];
