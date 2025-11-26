
# cluster destroy — Detailed

Location: `cmd/cluster/destroy/root.go`

Purpose:

- Stop and remove a local `kind` cluster created by this tool.

Usage:

```bash
k8s-local-bench cluster destroy [flags]
```

Flags:

- inherited: `--cluster-name` (optional), `--directory` (root CLI directory)

Behavior and details:

- If no `--cluster-name` is provided, the command lists existing `kind` clusters and prompts the user to select one interactively.
- The command deletes the cluster via the `utils/kind` helper and then attempts a best-effort shutdown of any `cloud-provider-kind` processes (uses `pkill -f 'sudo cloud-provider-kind'`).
- The CLI polls to confirm the cluster is no longer present and performs cleanup of local files for the cluster.

How `findKindConfig` searches for kind configs (used for locating cluster-specific config):

- If `cluster-name` is set: `$(directory)/clusters/<cluster-name>/kind-config.yaml|yml` and `kind*.y*ml` glob.
- Then `$(directory)/kind-config.yaml|yml` and `kind*.y*ml` glob.
- Then current working directory, same filename patterns.

Example:

```bash
# Delete cluster and associated load balancer
./k8s-local-bench cluster destroy --cluster-name local-bench
```

Developer note:

- Consider adding a `--force` flag for non-interactive deletion in the future and expand unit tests around `utils/kind` helpers.
