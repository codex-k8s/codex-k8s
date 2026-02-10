import type { RouteRecordRaw } from "vue-router";

import ProjectMembersPage from "../pages/ProjectMembersPage.vue";
import ProjectRepositoriesPage from "../pages/ProjectRepositoriesPage.vue";
import ProjectsPage from "../pages/ProjectsPage.vue";
import RunDetailsPage from "../pages/RunDetailsPage.vue";
import RunsPage from "../pages/RunsPage.vue";
import UsersPage from "../pages/UsersPage.vue";

export const routes: RouteRecordRaw[] = [
  { path: "/", name: "projects", component: ProjectsPage },
  {
    path: "/projects/:projectId/repositories",
    name: "project-repositories",
    component: ProjectRepositoriesPage,
    props: true,
    meta: { adminOnly: true },
  },
  {
    path: "/projects/:projectId/members",
    name: "project-members",
    component: ProjectMembersPage,
    props: true,
    meta: { adminOnly: true },
  },
  { path: "/runs", name: "runs", component: RunsPage },
  { path: "/runs/:runId", name: "run-details", component: RunDetailsPage, props: true },
  { path: "/users", name: "users", component: UsersPage, meta: { adminOnly: true } },
];

