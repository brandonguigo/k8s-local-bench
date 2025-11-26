# Charts: `localplane` and `local-stack`

This document describes the repository-provided Helm chart `localplane` (used to install cluster addons) and the `local-stack` chart that the CLI downloads into each local cluster project to enable direct development and deployment via the local ArgoCD instance.

Overview
- The repository contains a top-level chart `charts/localplane/` which is intended to install the project-provided addons (Headlamp, HAProxy, Victoria Metrics, reloader, httpbin, etc.) into a cluster.
- During `localplane cluster create`, the CLI ensures a per-project `local-argo` Git repository exists (under the configured `--directory` or current working dir). If `local-argo/charts/local-stack` is missing, the CLI downloads the `charts/local-stack` content from the repository (owner `brandonguigo`, ref `main`) into that path and commits it to the `local-argo` repo.

Why this matters
- Each local cluster project gets its own `local-stack` chart under `local-argo/charts/local-stack` so you can iterate on charts and have ArgoCD manage deployments from the local repo.
- The CLI patches the cluster kind config to mount `local-argo` into the nodes at `/mnt/local-argo` and installs ArgoCD (unless `--disable-argocd`), mounting the repo into ArgoCD so changes committed locally can be deployed by ArgoCD.

Paths and layout
- Repository chart: `charts/localplane/`
- Project-local chart (download target): `<project-base>/local-argo/charts/local-stack/`
- Bootstrap manifests applied into the cluster after create: `<project-base>/local-argo/charts/local-stack/bootstrap/` (these are applied by the CLI after ArgoCD installation)

How to use `local-stack` for development
1. Create a cluster (default behavior will download `local-stack` into `local-argo`):

```bash
go run main.go cluster create -y
```

2. Edit or add charts under `local-argo/charts/local-stack` in your project directory.

3. Commit your changes to the `local-argo` git repo (the CLI initializes and commits automatically when it downloads the chart; subsequent changes should be committed by you):

```bash
cd <project-base>/local-argo
git add charts/local-stack
git commit -m "Work on local-stack chart"
git push (if you have a remote; not required for a local Argo setup)
```

4. In ArgoCD (running inside the created cluster) configure an application that points to the `local-argo` repo path `charts/local-stack` (the CLI's bootstrap manifests may already create apps). When the repo changes, ArgoCD can sync and deploy your chart into the cluster.

Notes and caveats
- The CLI downloads `local-stack` from the repo owner `brandonguigo` and path `charts/local-stack` by default; adjust these values in code if you want a different source or ref.
- `--disable-argocd` skips creating `local-argo`, patching the kind config mount, and installing ArgoCD; in that case you can still manually copy `charts/local-stack` into a repo and configure your own delivery mechanism.
- The `local-argo` repo is created under the configured `--directory` (root CLI directory); default is `.`.

Troubleshooting
- If ArgoCD does not see changes, ensure the ArgoCD Application points at the correct repo path and that the repo is accessible from the cluster (the CLI mounts the local path into the cluster to make it available to ArgoCD).
- Check CLI logs for messages about downloading `local-stack` and for any git commit errors during the initial setup.

Want me to extend this section with examples for creating an ArgoCD Application YAML that points to `local-argo/charts/local-stack`? Say the word and I will add it.
