import type { RouteRecordRaw } from "vue-router";
import ProjectsView from "./views/ProjectsView.vue";
import ProjectRepositoriesView from "./views/ProjectRepositoriesView.vue";
import ProjectMembersView from "./views/ProjectMembersView.vue";
import RunsView from "./views/RunsView.vue";
import RunDetailsView from "./views/RunDetailsView.vue";
import UsersView from "./views/UsersView.vue";

export const routes: RouteRecordRaw[] = [
  { path: "/", name: "projects", component: ProjectsView },
  { path: "/projects/:project_id/repositories", name: "project-repositories", component: ProjectRepositoriesView },
  { path: "/projects/:project_id/members", name: "project-members", component: ProjectMembersView },
  { path: "/runs", name: "runs", component: RunsView },
  { path: "/runs/:run_id", name: "run-details", component: RunDetailsView },
  { path: "/users", name: "users", component: UsersView },
];
