
# cluster create — Detailed

Location: `cmd/cluster/create/root.go`

Purpose:

- Create a local `kind` cluster and perform several convenience setup steps (load-balancer, local-argo, ArgoCD, bootstrap manifests).

Usage:

```bash
localplane cluster create [flags]
```

Flags:

- `-y, --yes` (bool): skip interactive confirmation and proceed.
- `--start-lb` (bool, default: true): whether to start the local load balancer helper.
- `--lb-foreground` (bool, default: false): if true, run the load balancer in the foreground (blocking); if false, it runs in the background.
- `--disable-argocd` (bool, default: false): skip ArgoCD and `local-argo` setup.
- inherited: `--cluster-name` (optional — if omitted `create` will prompt and default to `local-bench` when left empty), `--directory` (root CLI directory)

High-level flow (implementation notes):

1. Logs an informational message: "Creating local k8s cluster...".
2. Honors `config.CliConfig.Debug` to enable debug logging inside the command.
3. Locates a kind configuration file using the same search order as `FindKindConfig` (cluster-specific, configured directory, then CWD). If none found, the command writes a default `kind-config.yaml` under `$(directory)/clusters/<cluster-name>/kind-config.yaml`.
4. Sets up `local-argo` (unless `--disable-argocd`): creates `local-argo` directory, initializes a git repo, downloads the `local-stack` chart into `local-argo/charts/local-stack` when missing, and commits the changes.
5. Patches the kind config to add an extra mount for `local-argo` at `/mnt/local-argo` and saves the updated kind config.
6. Asks for confirmation unless `--yes` is provided.
7. Calls `kindsvc.Create(clusterName, kindCfgPath)` to create the `kind` cluster.
8. Starts the cloud-provider-kind load balancer according to `--start-lb` / `--lb-foreground` flags.
9. Waits for cluster readiness by polling `kubectl`.
10. Unless `--disable-argocd` is set, installs/upgrades ArgoCD via the Helm SDK and mounts the `local-argo` repo into ArgoCD.
11. Applies bootstrap manifests found under `local-argo/charts/local-stack/bootstrap` into the cluster.

Notes about `utils/kind` responsibilities (refer to `utils/kind/kind.go`):

- `Create(name, kindConfigPath)` encapsulates invoking `kind` to create a cluster. It may accept an empty config path to use default behavior.
- `StartLoadBalancer(name, background)` starts the cloud-provider-kind process (background vs foreground behavior).

Examples:

```bash
# Interactive create (asks for confirmation)
go run main.go cluster create

# Non-interactive, use alternate directory and cluster name
go run main.go cluster create -d ../tmp --cluster-name test-cluster -y

# Create and run load balancer in foreground
./localplane cluster create --lb-foreground
```

Testing and verification tips:

- After `kindsvc.Create` returns, verify the cluster with `kubectl cluster-info --context kind-<cluster-name>`.
- Check load-balancer logs if running in foreground; for background mode, the helper should log process start and PID to the configured output.
