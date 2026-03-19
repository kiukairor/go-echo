# go-echo — Claude Code Instructions

## What is this?

A minimal Go web server using the Echo framework. Built as a test/sandbox service deployed on the local Raspberry Pi k8s cluster.

## Repo Structure

```
go-echo/
├── main.go                          ← app entry point
├── otel.go                          ← OpenTelemetry tracing + metrics init
├── Dockerfile                       ← multi-stage ARM64 build
├── go.mod / go.sum
├── argocd/
│   ├── go-echo-staging.yaml         ← ArgoCD Application for staging
│   └── go-echo-prod.yaml            ← ArgoCD Application for prod
├── infra/
│   └── helm/
│       └── go-echo/
│           ├── Chart.yaml
│           ├── values.yaml                  ← base values, image tag updated by CI
│           ├── values-laberry-staging.yaml  ← staging overrides
│           ├── values-laberry-prod.yaml     ← prod overrides
│           └── templates/
│               ├── deployment.yaml
│               └── service.yaml
├── .github/
│   └── workflows/
│       └── build.yml                ← build linux/arm64 → push GHCR → update tag
└── scripts/
    └── smoke-test.sh                ← hits all endpoints for quick validation
```

## Endpoints

| Method | Path     | Description                        |
|--------|----------|------------------------------------|
| GET    | /health  | Health check → `{"status":"ok","service":"go-echo"}` |
| GET    | /hello   | Returns `Hello, World!`            |
| POST   | /items   | Validates + enriches `{"name":"","value":""}` → returns `{"name":"","value":"","enriched":true}` with HTTP 201; HTTP 422 if `name` or `value` missing |

## GitOps Flow

```
git push → GitHub Actions
         → docker build linux/arm64 → push to ghcr.io/kiukairor/go-echo:sha-XXXXXXX
         → sed updates infra/helm/go-echo/values.yaml image tag
         → git commit [skip ci]
         → ArgoCD detects drift → deploys staging + prod to otel-test namespace on Pi cluster
```

## Infrastructure

- **Cluster**: kubeadm Kubernetes on Raspberry Pi (pimaster + piworker)
- **Namespace**: `otel-test`
- **ArgoCD apps**: `argocd/go-echo-staging.yaml` and `argocd/go-echo-prod.yaml` (defined in this repo)
- **Image registry**: `ghcr.io/kiukairor/go-echo`
- **Image pull secret**: `ghcr-secret` (must exist in `otel-test` namespace)
- **Environments**: staging (`OTEL_SERVICE_NAME=go-echo-staging`) and prod (`OTEL_SERVICE_NAME=go-echo-prod`)

## Local Development

```bash
# Run locally
PORT=8080 go run main.go

# Run smoke tests (requires port-forward or local server)
./scripts/smoke-test.sh
./scripts/smoke-test.sh http://localhost:8080   # custom base URL
```

## Accessing on the Cluster

```bash
kubectl port-forward svc/go-echo 8080:8080 -n otel-test
# then in another terminal:
curl http://localhost:8080/hello
```

## Code Conventions

- Port via `PORT` env var, default `8080`
- All endpoints return JSON errors, never plain text errors
- Health endpoint always: `GET /health` → `{"status":"ok","service":"go-echo"}`
- ARM64 Docker builds only (`platforms: linux/arm64`)
- Never hardcode secrets
