# §2 — Learn the object model by creating things

One folder per object type, in dependency order. Write the manifest yourself,
apply it, then read back what the API server filled in that you did not write:

```bash
kubectl apply -f 03-deployment/deployment.yaml
kubectl get deployment web -o yaml > ../scratch/deployment-readback.yaml
```

The defaulting, the `status` block, and the generated selectors are the lesson.
Diff your input against the readback — that gap is what Kubernetes did for you.

`kubectl explain` is the live API schema, and beats any tutorial:

```bash
kubectl explain deployment.spec.strategy
kubectl explain pod.spec.containers.resources --recursive
```

## Order

| # | Object | Why it comes here |
| --- | --- | --- |
| 01 | Pod | the only thing that actually runs; everything else manages these |
| 02 | ReplicaSet | adds "keep N of them alive" — your first reconcile loop |
| 03 | Deployment | adds rollout/rollback on top of ReplicaSet |
| 04 | Service | stable identity + load balancing across Pods |
| 05 | Ingress | HTTP routing from outside the cluster (needs the port mappings in §1) |
| 06 | ConfigMap / Secret | config injection, env vs volume |
| 07 | Namespace | scoping, and what is *not* namespaced |
| 08 | StatefulSet, Job/CronJob, DaemonSet | the remaining workload shapes |

Do 01 and 02 by hand before reaching for `kubectl create deployment`. Creating
a ReplicaSet manually and then deleting one of its Pods is the fastest way to
feel the reconciliation loop from §4.

## Worth doing at 05-ingress

kind does not ship an ingress controller. Install one before the Ingress object
does anything:

```bash
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
```
