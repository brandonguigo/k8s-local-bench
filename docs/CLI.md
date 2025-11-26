# k8s-local-bench CLI

This document describes the `k8s-local-bench` command-line interface, how configuration is loaded, and the cluster-related commands implemented in this repository.

## Overview

- Binary: `k8s-local-bench` (development: run with `go run main.go`)
- Purpose: create and manage a local Kubernetes cluster (using `kind` and a small load-balancer helper).

## Installation / Run (quick)

Development (no build step):

```bash
go run main.go <command> [flags]
```

Build a binary:

```bash
go build -o k8s-local-bench ./...
./k8s-local-bench <command> [flags]
```

## Configuration

The CLI supports configuration via (in precedence order): command-line flags, environment variables, and an optional config file.

- Config file flag: `--config, -c` — if provided, the specified file is used.
- Default config file: a hidden file named `.k8s-local-bench` (YAML) is searched for in `$HOME` if `--config` is not provided.
- Environment variables: prefixed with `K8S_LOCAL_BENCH` (e.g. `K8S_LOCAL_BENCH_DIRECTORY`).
- Flags are bound to Viper and can be set on the CLI; root persistent flags include `--directory` (`-d`).
 - Config file flag: `--config, -c` — if provided, the specified file is used.
 - Default config file: a hidden file named `.k8s-local-bench` (YAML) is searched for in `$HOME` if `--config` is not provided.
 - Environment variables: prefixed with `K8S_LOCAL_BENCH` (e.g. `K8S_LOCAL_BENCH_DIRECTORY`).
 - Flags are bound to Viper and can be set on the CLI; root persistent flags include `--directory` (`-d`) and `--config` (`-c`).

Config fields (unmarshalled into `config.CliConfig`):

- `debug` (bool): enable debug logging (can also be set via `LOG_LEVEL=debug`).
- `directory` (string): directory where configurations and data are stored. This is used by commands to look for cluster-specific config files (e.g. `clusters/<name>/kind-config.yaml`).

Example env usage:

```bash
export K8S_LOCAL_BENCH_DIRECTORY=/path/to/configs
export K8S_LOCAL_BENCH_DEBUG=true
go run main.go cluster create
```

## Project-provided Kind config example

There is a sample kind config at the repository root named `cluster-config.yaml`. A minimal example looks like:

```yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
```

Commands may also search for files named `kind-config.yaml`, `kind.yaml`, or similar under the configured `directory` and under `clusters/<cluster-name>`.

## Commands

Top-level command: `k8s-local-bench`.

Subcommand group: `cluster` — control local k8s clusters.

Persistent flags available to the `cluster` command and its subcommands:

- `--cluster-name` (string) — optional. If not provided, `cluster create` will prompt for a name and default to `local-bench` when left empty. Many commands also accept an interactive selection when the name is omitted.

### cluster create

Usage:

```bash
k8s-local-bench cluster create [flags]
```

What it does:

- Creates a local `kind` cluster (the implementation uses the `utils/kind` helper).
- Locates or creates a kind config: the CLI searches in this order:
  - `$(directory)/clusters/<cluster-name>/kind-config.yaml|yml` and `kind*.y*ml` glob
  - `$(directory)/kind-config.yaml|yml` and `kind*.y*ml` glob
  - current working directory `kind-config.yaml|yml` and `kind*.y*ml` glob
  If no config is found, `create` will write a default `kind-config.yaml` under `$(directory)/clusters/<cluster-name>/kind-config.yaml`.

Flags specific to `create`:

- `-y, --yes` (bool): don't ask for confirmation; assume yes.
- `--start-lb` (bool, default: true): start the local load balancer (cloud-provider-kind helper).
- `--lb-foreground` (bool, default: false): run load balancer in the foreground (blocking); otherwise it runs in background.
- `--disable-argocd` (bool, default: false): skip ArgoCD/local-argo setup and ArgoCD Helm install.

Examples:

```bash
# Run interactively and allow confirmation
go run main.go cluster create

# Provide directory and auto-confirm
go run main.go cluster create -d ../tmp -y

# Don't start the built-in load balancer
./k8s-local-bench cluster create --start-lb=false

# Run load balancer in foreground (blocking)
./k8s-local-bench cluster create --lb-foreground
```

Notes:

- The command references `config.CliConfig.Debug` and respects the global debug setting (or `LOG_LEVEL=debug`).
- `create` performs additional convenience steps by default:
  - It initializes a `local-argo` git repository under the configured `directory` (or CWD when not set), if not disabled.
  - It patches the kind config to mount the `local-argo` directory into the kind nodes at `/mnt/local-argo`.
  - If the `local-stack` chart is missing in `local-argo/charts/local-stack`, the CLI downloads the chart path from the repository `brandonguigo/k8s-local-bench` (ref: `main`) into that location and commits the change to the `local-argo` repo.
  - Unless `--disable-argocd` is set, the CLI will install or upgrade ArgoCD via the Helm SDK and mount the `local-argo` repo into ArgoCD.
  - After cluster creation the CLI applies bootstrap manifests from `local-argo/charts/local-stack/bootstrap` into the cluster.

### cluster destroy

Usage:

```bash
k8s-local-bench cluster destroy [flags]
```

What it does:

- Deletes/stops a local `kind` cluster using the `utils/kind` helper.
- If no `--cluster-name` is provided, it will list existing `kind` clusters and prompt for an interactive selection.

Flags:

- Inherits `--cluster-name` from `cluster` persistent flags.

Behavior details:

- The command attempts to delete the cluster via the `kind` helper. It then performs a best-effort stop of any running `cloud-provider-kind` processes (the implementation invokes `pkill -f 'sudo cloud-provider-kind'`).
- The CLI polls briefly to ensure the cluster has been removed and performs local cleanup of files associated with the cluster directory.

## Examples & common workflows


Create a cluster using a config stored under the CLI `directory`:

```bash
export K8S_LOCAL_BENCH_DIRECTORY=$PWD
go run main.go cluster create -y
```

Create a cluster with an explicit config file and run the load balancer in foreground:

```bash
go run main.go cluster create -y --lb-foreground
```

Destroy:

```bash
go run main.go cluster destroy --cluster-name local-bench
```

## Next steps / developer notes

- If you want `cluster destroy` to be available from the CLI, ensure it is added to the `cluster` command (the code for `destroy` exists under `cmd/cluster/destroy` but may not be wired into `cmd/cluster/root.go`).
- Consider implementing cluster status checks and more robust waiting logic after `kind` creation to confirm the cluster is ready.

- Charts: See `docs/charts.md` for details on the `k8s-local-bench` repository chart and the `local-stack` chart that the CLI downloads into each project for local ArgoCD-driven development.

---

If you want, I can also:
- add examples directly to the repository `README.md`, or
- wire `destroy` into the `cluster` root command so it's available at runtime.
