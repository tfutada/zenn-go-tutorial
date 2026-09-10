# Kubernetes Study Guide (for Go developers)

Kubernetes clicks fastest if you treat it as a distributed reconciliation
system written in Go, not as a YAML tool. This guide assumes a Go background
with experience in concurrency and network programming — that background is a
real shortcut, and the path below leans on it.

## 1. Get a real cluster on your laptop, day one

```bash
brew install kind kubectl
kind create cluster --name study
kubectl cluster-info --context kind-study
```

`kind` (or `k3d`) runs natively on arm64 and starts in about 30 seconds. Skip
managed cloud clusters while learning — they cost money and hide the control
plane.

**Apple Silicon gotcha:** images you `docker build` on this machine are arm64.
That is fine for a local kind cluster, but the same image silently fails on an
amd64 cloud cluster — the ELF loader rejects the binary before your entrypoint
runs, so the logs are empty. For anything destined for a cloud cluster:

```bash
docker buildx build --platform linux/amd64 --load -t myapp:dev .
docker image inspect myapp:dev --format '{{.Os}}/{{.Architecture}}'
```

To use a locally built image inside kind, load it into the cluster nodes:

```bash
kind load docker-image myapp:dev --name study
```

## 2. Learn the object model by creating things, not reading about them

Work through these in order, creating each one for real:

1. Pod
2. ReplicaSet
3. Deployment
4. Service
5. Ingress
6. ConfigMap / Secret
7. Namespace
8. StatefulSet, Job / CronJob, DaemonSet

For each object, create it and then read back what the API server filled in
that you did not write:

```bash
kubectl create deployment web --image=nginx
kubectl get deployment web -o yaml
```

The defaulting, the status block, and the generated selectors teach more than
any tutorial. `kubectl explain` is the underrated tool here — it is the live API
schema in your terminal:

```bash
kubectl explain deployment.spec.strategy
kubectl explain pod.spec.containers.resources --recursive
```

## 3. Get fluent at the debugging loop before anything else

This is 80% of real Kubernetes work.

```bash
kubectl describe pod <name>
kubectl logs -f <pod> --previous          # --previous reads the crashed container
kubectl exec -it <pod> -- sh
kubectl port-forward svc/<name> 8080:80
kubectl get events --sort-by=.lastTimestamp
kubectl get pod <name> -o jsonpath='{.status.containerStatuses[0].state}'
```

Then break things deliberately and diagnose each one:

| Symptom | What you should learn to check |
| --- | --- |
| `CrashLoopBackOff` | `logs --previous`, exit code, entrypoint, missing env var |
| `ImagePullBackOff` | image name/tag, registry auth, wrong architecture |
| `OOMKilled` | memory limits vs actual usage, exit code 137 |
| Pod stuck `Pending` | `describe` events — no node fits requests, taints, no PV |
| Service returns nothing | readiness probe failing, so Endpoints is empty |

That last one is the single most common real-world confusion: a Service routes
only to *ready* Pods, so a broken readiness probe looks like a networking bug.

```bash
kubectl get endpoints <service>   # empty means no ready backends
```

Failure diagnosis is the skill. Writing YAML is not.

## 4. Then learn the mental model that makes it all cohere

Everything in Kubernetes is one loop:

```
declare desired state (etcd)
        -> a controller watches for changes
        -> it reconciles actual state toward desired
        -> forever, idempotently
```

There is no imperative "deploy" action. You write down what should be true, and
independent controllers spin until reality matches. Once this lands, `kubectl`
stops feeling like magic.

The Go payoff: the `client-go` architecture is a watch-driven cache feeding a
producer-consumer work queue — Reflector -> DeltaFIFO -> Informer -> Indexer ->
workqueue -> worker goroutines. That is the same shape as the producer-consumer
and fan-in examples in `src/` of this repo, just distributed. Read the informer
source with that mapping in mind and it reads like familiar code.

## 5. Lean into the parts your background makes you good at

**Networking.** ClusterIP vs NodePort vs LoadBalancer, how kube-proxy programs
iptables/IPVS rules, what the CNI plugin actually does to the network namespace,
and CoreDNS resolution (`<svc>.<ns>.svc.cluster.local`). With a net-programming
background this is where you will move fastest.

**Write a controller.** A CRD plus a reconcile loop, roughly 200 lines of Go
with kubebuilder / controller-runtime. Nothing teaches the reconciliation model
faster, because you have to make your own loop idempotent and handle the fact
that it will be called again for reasons you did not cause.

## 6. Structure, if you want a forcing function

The CKA is hands-on — a live cluster and a time limit, not multiple choice — so
studying for it forces the kubectl fluency from step 3. The exam-prep scenarios
are useful practice even if you never sit the exam.

## What to skip early

Helm, service meshes, GitOps/ArgoCD, and operator frameworks-as-a-product. They
are abstractions over concepts you do not have yet, and starting there leaves
you unable to debug anything underneath them.

## Resources

https://kubernetes.io/docs/concepts/  — unusually good; read it after you have played, not before

https://github.com/kelseyhightower/kubernetes-the-hard-way  — bootstrap a cluster from scratch, no kubeadm; do this once the basics are solid

https://killercoda.com/killer-shell-cka  — free browser-based scenarios on real clusters

https://book.kubebuilder.io/  — the operator / Go path

https://github.com/kubernetes/client-go  — read the informer and workqueue packages directly

"Kubernetes Up & Running" (O'Reilly) — solid conceptual intro

"Programming Kubernetes" (O'Reilly) — the client-go / controller book for Go devs
