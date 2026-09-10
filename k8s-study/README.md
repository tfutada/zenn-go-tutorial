# Kubernetes Study Workspace

Working directory for `../KUBERNETES_STUDY.md`. That file is the syllabus; this
tree is where you actually build things. Each numbered folder maps to a section
of the guide.

The point of the layout is that **the folders start empty on purpose.** The
guide's whole argument is that you learn the object model by writing manifests
and breaking them, not by reading finished ones. Pre-filled examples would let
you skip the part that teaches.

| Folder | Guide section | What goes in it |
| --- | --- | --- |
| `01-cluster/` | §1 | kind cluster config and setup notes |
| `02-objects/` | §2 | one manifest per object type, written by you |
| `03-debugging/` | §3 | deliberately broken manifests, one per symptom |
| `04-reconciliation/` | §4 | notes on the controller loop and client-go |
| `05-networking/` | §5 | Service types, DNS, kube-proxy experiments |
| `06-controller/` | §5 | your CRD + reconcile loop (own Go module) |
| `scratch/` | — | `kubectl -o yaml` dumps; gitignored |

## Current local state

- `kubectl` v1.29.3 — installed via Homebrew, aliased to `kubecolor` in your shell
- `docker` — installed
- `kind` — **not installed yet**, start here: `brew install kind`

## Suggested order

Sections 1 → 3 in order, and do not skip ahead to 4. The guide is emphatic that
the debugging loop comes before the theory, because the theory only sticks once
you have seen the failure modes it explains.

## Progress

- [ ] §1 kind cluster running
- [ ] §2 all eight object types created and inspected
- [ ] §3 all five failure modes reproduced and diagnosed
- [ ] §4 informer/workqueue source read against `src/producer_consumer_fanin`
- [ ] §5 networking experiments
- [ ] §5 controller written
