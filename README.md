# go-echo

A minimal Go web server using the [Echo](https://echo.labstack.com/) framework, deployed on a Raspberry Pi Kubernetes cluster via ArgoCD.

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Health check |
| `GET` | `/hello` | Returns Hello, World! |
| `POST` | `/items` | Accepts `{"name":"...","value":"..."}`, echoes it back |

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

## Deployment

Deployed automatically on every push to `main`:

1. GitHub Actions builds a `linux/arm64` Docker image and pushes to `ghcr.io/kiukairor/go-echo`
2. CI commits the new image tag to `infra/helm/go-echo/values.yaml`
3. ArgoCD detects the change and rolls out to the `otel-test` namespace on the Pi cluster

To check status:

```bash
kubectl get pods -n otel-test
kubectl get application go-echo -n argocd
```
