# §4 — The mental model

Everything in Kubernetes is one loop:

```
declare desired state (etcd)
        -> a controller watches for changes
        -> it reconciles actual state toward desired
        -> forever, idempotently
```

There is no imperative "deploy" action. You write down what should be true and
independent controllers spin until reality matches.

## Prove it to yourself

Before reading any source, do this against the cluster:

```bash
kubectl scale deployment web --replicas=3
kubectl delete pod <one-of-them>        # watch it come back
kubectl get pods -w
```

Nothing "re-ran" your delete. A controller noticed the gap and closed it.

## The Go payoff

`client-go` is a watch-driven cache feeding a producer-consumer work queue:

```
Reflector -> DeltaFIFO -> Informer -> Indexer -> workqueue -> worker goroutines
```

That is the same shape as the concurrency examples already in this repo, just
distributed across a network:

| client-go piece | Closest example in `src/` |
| --- | --- |
| Reflector (watch -> feed) | `src/producer_consumer` — producer side |
| DeltaFIFO / workqueue | `src/producer_consumer_fanin` — bounded hand-off, fan-in |
| worker goroutines | consumer side of the same examples |
| dedup of in-flight keys | `src/single_flight` — collapse duplicate work |
| rate-limited requeue | `src/rate` — `golang.org/x/time/rate` |
| watch cancellation / shutdown | `src/context`, `src/background_task` |

Read the informer source with that mapping in mind and it reads like familiar
code rather than framework magic.

- https://github.com/kubernetes/client-go — read `tools/cache` and `util/workqueue` directly

## Notes

<!-- where the analogy holds, and where it breaks down -->
