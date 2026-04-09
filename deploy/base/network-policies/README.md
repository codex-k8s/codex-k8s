# Network Policies

## Назначение

Каталог содержит baseline шаблоны `NetworkPolicy` для `kodex`:

- `platform-baseline.yaml.tpl` — безопасный baseline для platform namespace:
  - ограничение ingress в `postgres` только от `kodex`;
  - ограничение egress у `kodex` до `postgres`, DNS, Kubernetes API (обычно `nodeIP:6443`) и web (`80/443`).
- `project-agent-baseline.yaml.tpl` — шаблон изоляции namespace проектов/агентов:
  - default deny ingress/egress;
  - allow ingress от `platform/system` namespace;
  - allow egress в DNS, MCP платформы и web (`80/443`).

## Namespace labels

Для работы меж-namespace правил используются labels:

- `kodex.io/network-zone=platform`
- `kodex.io/network-zone=system`
- `kodex.io/network-zone=project`

Лейблы проставляет runtime deploy сервис control-plane во время reconcile namespace.

## Применение

Platform baseline (по умолчанию в bootstrap/runtime deploy):

```bash
go run ./cmd/codex-bootstrap render-manifest \
  --template deploy/base/network-policies/platform-baseline.yaml.tpl \
  --var KODEX_PRODUCTION_NAMESPACE=kodex-prod \
  --var KODEX_K8S_API_CIDR=<node-ip>/32 \
  --var KODEX_K8S_API_PORT=6443 \
  | kubectl apply -f -
```

Project/agent baseline для конкретного namespace:

```bash
go run ./cmd/codex-bootstrap render-manifest \
  --template deploy/base/network-policies/project-agent-baseline.yaml.tpl \
  --var KODEX_TARGET_NAMESPACE=project-demo \
  --var KODEX_PLATFORM_MCP_PORT=80 \
  | kubectl apply -f -
```
