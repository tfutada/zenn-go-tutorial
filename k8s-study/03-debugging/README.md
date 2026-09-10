# §3 — The debugging loop

80% of real Kubernetes work. Get fluent here before §4.

## The loop

```bash
kubectl describe pod <name>
kubectl logs -f <pod> --previous          # --previous reads the crashed container
kubectl exec -it <pod> -- sh
kubectl port-forward svc/<name> 8080:80
kubectl get events --sort-by=.lastTimestamp
kubectl get pod <name> -o jsonpath='{.status.containerStatuses[0].state}'
```

## Break it on purpose

One folder per symptom. Write a manifest that *causes* the symptom, then
diagnose it with only the commands above — no peeking at your own manifest
until you have found it from the cluster's side.

| Folder | Symptom | What to check |
| --- | --- | --- |
| `crashloopbackoff/` | `CrashLoopBackOff` | `logs --previous`, exit code, entrypoint, missing env var |
| `imagepullbackoff/` | `ImagePullBackOff` | image name/tag, registry auth, wrong architecture |
| `oomkilled/` | `OOMKilled` | memory limits vs actual usage, exit code 137 |
| `pending/` | Pod stuck `Pending` | `describe` events — no node fits requests, taints, no PV |
| `empty-endpoints/` | Service returns nothing | readiness probe failing, so Endpoints is empty |

## The one that catches everyone

`empty-endpoints/`. A Service routes only to **ready** Pods, so a broken
readiness probe presents as a networking bug. The Pod is running, the Service
exists, and requests still go nowhere.

```bash
kubectl get endpoints <service>   # empty means no ready backends
```

Reproduce it by giving a healthy Pod a readiness probe pointed at a path or
port that does not exist. Then fix the probe and watch Endpoints populate.

## Notes

<!-- symptom -> what actually told you the answer -->
