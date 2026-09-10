# §5 — Networking

The guide flags this as where a net-programming background moves fastest.

## Ground to cover

- **Service types** — ClusterIP vs NodePort vs LoadBalancer, and what each one
  actually allocates. (LoadBalancer stays `Pending` forever on kind; that is
  correct behavior, not a bug — there is no cloud provider to fulfill it.)
- **kube-proxy** — how it programs iptables/IPVS rules to make a virtual
  ClusterIP work with no process listening on it
- **CNI** — what the plugin does to the Pod's network namespace (veth pair,
  address assignment, routes)
- **CoreDNS** — `<svc>.<ns>.svc.cluster.local` resolution, and the search-domain
  shortcuts that make bare `<svc>` work inside a namespace

## Experiments worth running

```bash
# Watch the ClusterIP with nothing listening on it
kubectl get svc web -o wide
kubectl exec -it <pod> -- sh -c 'wget -qO- http://web'   # works anyway — why?

# DNS from inside the cluster
kubectl exec -it <pod> -- cat /etc/resolv.conf
kubectl exec -it <pod> -- nslookup web.default.svc.cluster.local

# The iptables rules kube-proxy wrote (on a kind node)
docker exec -it study-worker iptables -t nat -L KUBE-SERVICES -n | head -40
```

The `wget` succeeding against an IP that nothing binds to is the thing worth
sitting with — that is kube-proxy's NAT rules, not a listening socket.

## Notes

<!-- -->
