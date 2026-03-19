# go-echo

A minimal Go web server using the [Echo](https://echo.labstack.com/) framework, deployed on a Raspberry Pi Kubernetes cluster via ArgoCD. Includes OpenTelemetry tracing exported to New Relic.

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Health check → `{"status":"ok","service":"go-echo"}` |
| `GET` | `/hello` | Returns `Hello, World!` |
| `POST` | `/items` | Validates and enriches `{"name":"...","value":"..."}` → returns `{"name":"...","value":"...","enriched":true}` (422 if name or value missing) |

## Run locally

```bash
go run main.go
# or with a custom port:
PORT=9000 go run main.go
```

## Smoke test

```bash
# against local server (default: http://localhost:8080)
./scripts/smoke-test.sh

# against the cluster via port-forward
kubectl port-forward svc/go-echo 8080:8080 -n otel-test
./scripts/smoke-test.sh http://localhost:8080
```

## Tracing

Traces are exported via OTLP HTTP to New Relic. Configure with env vars:

| Var | Default | Description |
|-----|---------|-------------|
| `NEW_RELIC_LICENSE_KEY` | — | New Relic ingest key (required) |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | `https://otlp.nr-data.net` | OTLP endpoint |
| `OTEL_SERVICE_NAME` | `go-echo` | Service name in traces |
| `SERVICE_ENV` | `laberry` | Deployment environment attribute |

## Deployment

Deployed automatically on every push to `main`:

1. GitHub Actions builds a `linux/arm64` Docker image and pushes to `ghcr.io/kiukairor/go-echo`
2. CI commits the new image tag to `infra/helm/go-echo/values.yaml`
3. ArgoCD detects the change and rolls out staging and prod to the `otel-test` namespace on the Pi cluster

ArgoCD apps are defined in `argocd/go-echo-staging.yaml` and `argocd/go-echo-prod.yaml`.

To check status:

```bash
kubectl get pods -n otel-test
kubectl get application go-echo-staging go-echo-prod -n argocd
```
