import type { RouteRecordRaw } from "vue-router";
import ProjectsView from "./views/ProjectsView.vue";
import RunsView from "./views/RunsView.vue";
import UsersView from "./views/UsersView.vue";

export const routes: RouteRecordRaw[] = [
  { path: "/", name: "projects", component: ProjectsView },
  { path: "/runs", name: "runs", component: RunsView },
  { path: "/users", name: "users", component: UsersView },
];

