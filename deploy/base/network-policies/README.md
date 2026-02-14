# Network Policies

## Назначение

Каталог содержит baseline шаблоны `NetworkPolicy` для `codex-k8s`:

- `platform-baseline.yaml.tpl` — безопасный baseline для platform namespace:
  - ограничение ingress в `postgres` только от `codex-k8s`;
  - ограничение egress у `codex-k8s` до `postgres`, DNS, Kubernetes API (обычно `nodeIP:6443`) и web (`80/443`).
- `project-agent-baseline.yaml.tpl` — шаблон изоляции namespace проектов/агентов:
  - default deny ingress/egress;
  - allow ingress от `platform/system` namespace;
  - allow egress в DNS, MCP платформы и web (`80/443`).

## Namespace labels

Для работы меж-namespace правил используются labels:

- `codexk8s.io/network-zone=platform`
- `codexk8s.io/network-zone=system`
- `codexk8s.io/network-zone=project`

Лейблы проставляет `deploy/scripts/apply_network_policy_baseline.sh`.

## Применение

Platform baseline (по умолчанию в bootstrap):

```bash
export CODEXK8S_STAGING_NAMESPACE=codex-k8s-ai-staging
export CODEXK8S_K8S_API_CIDR="<node-ip>/32"
export CODEXK8S_K8S_API_PORT=6443
bash deploy/scripts/apply_network_policy_baseline.sh
```

Project/agent baseline для конкретного namespace:

```bash
export CODEXK8S_APPLY_PROJECT_AGENT_POLICY=true
export CODEXK8S_TARGET_NAMESPACE=project-demo
export CODEXK8S_PLATFORM_MCP_PORT=80
bash deploy/scripts/apply_network_policy_baseline.sh
```
