# Examples & Common Workflows

This page shows runnable examples for common scenarios.

# Create a cluster
1) Create a cluster using repository root config

```bash
# Use the repo root as the CLI directory so the included cluster-config.yaml is used
export LOCALPLANE_DIRECTORY=$(pwd)
go run main.go cluster create -y
```

2) Create a cluster using a named cluster config stored under the CLI directory

```bash
# Example: ~/.localplane/clusters/mytest/kind-config.yaml
export LOCALPLANE_DIRECTORY=$HOME/.localplane
go run main.go cluster create --cluster-name mytest -y
```

3) Create but don't start the load balancer

```bash
./localplane cluster create --start-lb=false -y
```

4) Run the load balancer in foreground for debugging

```bash
./localplane cluster create --lb-foreground
```

Kind config sample (repository `cluster-config.yaml`):

```yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  # You can add extraPortMappings to expose container ports to host
```

Verification commands after creation:

```bash
kubectl cluster-info --context kind-local-bench
kubectl get nodes
```

# Work inside the cluster locally

To work inside the cluster, a local-argo local git repository have been created.

Inside of this repository, you can edit charts/local-stack as you like : add ArgoCD apps, add pods, deployments, anything you'd like. 

You can also create new charts inside the charts directory, then use them as an ArgoCD app inside local-stack.

You can add an origin to the local-argo repository to push your local workbench to a remote Git server like GitHub.

Basically, you can deploy anything you'd like from a simple http api to a complete microservice stack reusing the production helm chart for your platform with databases, operators, redis, rabbitmq, clickhouse...

# Destroy the cluster

Destroying a cluster

```bash
./localplane cluster destroy --cluster-name local-bench
```