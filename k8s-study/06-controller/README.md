# §5 — Write a controller

A CRD plus a reconcile loop, ~200 lines of Go with kubebuilder /
controller-runtime. The guide's claim: nothing teaches the reconciliation model
faster, because you have to make your own loop idempotent and handle being
called again for reasons you did not cause.

Do this **after** §2–§4. It is the payoff, not the entry point.

## This gets its own Go module

Do not let kubebuilder scaffold into the parent `tutorial1` module. The
controller-runtime dependency tree is large and unrelated to the tutorial
examples, and a nested `main` package here would break `go build` at the repo
root the same way the examples listed in `../../CLAUDE.md` already do.

`kubebuilder init` writes its own `go.mod` in whatever directory you run it in,
so running it here keeps the module boundary right — just confirm afterward
that a `go.mod` exists in this folder before committing anything.

```bash
brew install kubebuilder
# from inside this directory:
kubebuilder init --domain futasoft.dev --repo futasoft.dev/k8s-study-controller
kubebuilder create api --group study --version v1alpha1 --kind Widget
```

Then verify the boundary held:

```bash
ls go.mod                            # must exist here
cd ../.. && go build ./... 2>&1 | grep k8s-study   # must print nothing
```

## What to actually make it do

Keep the CRD trivial — the reconcile semantics are the lesson, not the schema.
Something like "a Widget with `spec.replicas` that creates and owns that many
ConfigMaps" is enough to force every real question:

- Reconcile gets called again with no change. Does it do anything harmful?
- Someone deletes a ConfigMap you own. Do you notice, and how?
- Someone edits one of your ConfigMaps by hand. Do you fight them, and forever?
- Your process restarts mid-reconcile. Is the half-finished state recoverable?

That last set is exactly the idempotency argument from §4, except now it is
your bug.

- https://book.kubebuilder.io/

## Notes

<!-- -->
