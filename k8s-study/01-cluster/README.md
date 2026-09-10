# §1 — Get a real cluster on your laptop

## Install

```bash
brew install kind          # kubectl is already on this machine (v1.29.3)
```

## Create

```bash
kind create cluster --config kind-config.yaml
kubectl cluster-info --context kind-study
kubectl get nodes          # expect 1 control-plane + 2 workers
```

`kind-config.yaml` differs from the guide's bare `kind create cluster --name
study` in two ways, both deliberate — see the comments in that file.

## Teardown / reset

Cheap and expected. Break the cluster freely; rebuilding takes ~30s.

```bash
kind delete cluster --name study
```

## Apple Silicon note

Your images build as arm64, which is fine for this local kind cluster. The
cross-arch trap only bites when an image is destined for an amd64 cloud
cluster — the failure is silent, with empty logs. See §1 of the guide.

Loading a locally built image into the cluster (kind nodes do not share your
host's Docker image store):

```bash
kind load docker-image myapp:dev --name study
```

## Notes

<!-- what surprised you here -->
