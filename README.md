<div align="center">

# K8S-Lab-Everything

### Practice Kubernetes Troubleshooting Like the CKA Exam

[![CI](https://github.com/WhoisMonesh/K8S-Lab-Everything/actions/workflows/ci.yaml/badge.svg)](https://github.com/WhoisMonesh/K8S-Lab-Everything/actions/workflows/ci.yaml)
[![Release](https://img.shields.io/github/v/release/WhoisMonesh/K8S-Lab-Everything)](https://github.com/WhoisMonesh/K8S-Lab-Everything/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go)](https://go.dev)
[![Platform](https://img.shields.io/badge/Platform-macOS%20%7C%20Linux%20%7C%20Windows-blue)]()

</div>

---

## What is this?

**cka-lab-runner** is a CLI tool that creates broken Kubernetes scenarios for you to fix — just like the CKA exam. It spins up a local cluster, breaks something specific, and lets you practice real troubleshooting with `kubectl`. When you're done, it verifies your fix is correct.

```
╔═══════════════════════════════════════════════════════════════╗
║  $ cka-lab-runner lab run pod_crashloop                       ║
║                                                               ║
║  Lab: Pod in CrashLoopBackOff                                 ║
║  Category: workloads | Difficulty: easy | Time: 15min         ║
║                                                               ║
║  A deployment named 'webapp' in the default namespace is      ║
║  failing to start. The pods are stuck in CrashLoopBackOff.    ║
║                                                               ║
║  Hints:                                                       ║
║    1. Check the pod status and events                         ║
║    2. Look at the pod logs to see why it's crashing           ║
║    3. Check the container image and command configuration     ║
║    4. Environment variables might be incorrectly configured   ║
╚═══════════════════════════════════════════════════════════════╝
```

## Why?

| Problem | Solution |
|---------|----------|
| CKA exam is 100% hands-on, but most study material is theory | Break and fix real clusters |
| Setting up practice environments is tedious | One command does everything |
| Labs get stale and outdated | OTA auto-update fetches new labs |
| Hard to find realistic troubleshooting scenarios | 50 labs across 8 CKA domains |
| Cross-platform support is rare | Works on macOS, Linux, and Windows |

## Quick Start

### One-Command Setup

```bash
# Clone and run the automated setup
git clone https://github.com/WhoisMonesh/K8S-Lab-Everything.git
cd K8S-Lab-Everything
./scripts/setup.sh
```

The setup script automatically installs:
- **Go** (if not present)
- **Docker** (Docker Desktop on macOS/Windows, Docker Engine on Linux)
- **kubectl**
- **kind** (or k3d/minikube — pass `--install-cluster-provider k3d` to choose)
- **cka-lab-runner** binary (installed to `/usr/local/bin`)

**Windows:**
```powershell
git clone https://github.com/WhoisMonesh/K8S-Lab-Everything.git
cd K8S-Lab-Everything
powershell -ExecutionPolicy Bypass -File scripts\setup.ps1
```

### Your First Lab

```bash
# 1. Initialize config
$ cka-lab-runner init
✓ Created config file: cka-lab-runner.yaml

# 2. Create a local cluster
$ cka-lab-runner up
ℹ Creating cluster: cka-lab
✓ Cluster created: cka-lab

# 3. List available labs
$ cka-lab-runner lab list

# 4. Run a lab (breaks something on purpose)
$ cka-lab-runner lab run pod_crashloop

# 5. Debug and fix using kubectl
$ kubectl get pods
$ kubectl describe pod -l app=webapp
$ kubectl logs -l app=webapp

# 6. Verify your fix
$ cka-lab-runner lab verify pod_crashloop
✓ Congratulations! You successfully fixed: Pod in CrashLoopBackOff

# 7. View solution if stuck
$ cka-lab-runner lab solution pod_crashloop

# 8. Clean up
$ cka-lab-runner down
```

## Auto-Update

The binary automatically checks for new versions when you run any command. If an update is available:

```
=== UPDATE AVAILABLE ===
  Current version: 1.0.1
  Latest version:  1.1.0

  Run cka-lab-runner update to install the latest version.
```

```bash
# Update to latest release
$ cka-lab-runner update
New version available: 1.1.0 (current: 1.0.1)
Downloading...
Need admin permissions to install...
Updated successfully! (1.0.1 -> 1.1.0)
```

## Commands

| Command | Description |
|---------|-------------|
| `cka-lab-runner init` | Create config file |
| `cka-lab-runner up` | Create local cluster |
| `cka-lab-runner up --recreate` | Recreate existing cluster |
| `cka-lab-runner down` | Delete cluster |
| `cka-lab-runner version` | Show current version |
| `cka-lab-runner update` | Update to latest release |
| `cka-lab-runner lab list` | List all labs |
| `cka-lab-runner lab list --category networking` | Filter by category |
| `cka-lab-runner lab list --difficulty easy` | Filter by difficulty |
| `cka-lab-runner lab random` | Random lab |
| `cka-lab-runner lab random --category storage` | Random in category |
| `cka-lab-runner lab run <lab-id>` | Run a lab |
| `cka-lab-runner lab verify <lab-id>` | Verify your fix |
| `cka-lab-runner lab solution <lab-id>` | Show solution |

## Available Labs (50)

### Control Plane (7 labs)

| ID | Lab | Difficulty | Time |
|----|-----|-----------|------|
| `etcd_wrong_ip` | Fix API server → etcd communication | Medium | 25min |
| `scheduler_not_running` | Debug broken kube-scheduler | Medium | 20min |
| `cluster_upgrade` | Cluster upgrade simulation | Hard | 30min |
| `etcd_backup_restore` | etcd backup and restore | Hard | 30min |
| `kubelet_stopped` | Fix stopped kubelet service | Medium | 20min |
| `node_not_ready` | Fix kubelet on NotReady node | Medium | 20min |
| `node_pressure` | Clear disk/memory pressure on node | Hard | 25min |

### Networking (5 labs)

| ID | Lab | Difficulty | Time |
|----|-----|-----------|------|
| `network_policy_blocking` | Fix NetworkPolicy blocking traffic | Medium | 20min |
| `ingress_broken` | Fix Ingress configuration | Medium | 20min |
| `service_no_endpoints` | Fix Service with no endpoints | Medium | 20min |
| `service_wrong_selector` | Fix Service selector not matching pods | Easy | 10min |
| `multi_container_pod` | Fix multi-container pod communication | Medium | 15min |

### Scheduling (3 labs)

| ID | Lab | Difficulty | Time |
|----|-----|-----------|------|
| `taint_no_toleration` | Schedule pods onto tainted nodes | Medium | 20min |
| `node_affinity_mismatch` | Fix broken node affinity selectors | Hard | 25min |
| `pod_scheduling_failed` | Fix pod nodeSelector mismatch | Easy | 10min |

### DNS (1 lab)

| ID | Lab | Difficulty | Time |
|----|-----|-----------|------|
| `coredns_broken_config` | Fix CoreDNS configuration | Easy | 15min |

### Storage (3 labs)

| ID | Lab | Difficulty | Time |
|----|-----|-----------|------|
| `pvc_pending` | Debug PVC stuck in Pending | Medium | 20min |
| `pv_not_binding` | Fix PersistentVolume not binding to PVC | Medium | 20min |
| `pod_host_path_wrong` | Fix wrong hostPath mount | Medium | 15min |

### Security (4 labs)

| ID | Lab | Difficulty | Time |
|----|-----|-----------|------|
| `cert_expiration` | Check certificate expiration | Hard | 25min |
| `secret_env_broken` | Fix app failing due to bad Secret data | Easy | 15min |
| `secret_missing` | Create missing Secret for pod | Easy | 10min |
| `pod_security_context` | Fix pod securityContext misconfiguration | Medium | 15min |

### RBAC (1 lab)

| ID | Lab | Difficulty | Time |
|----|-----|-----------|------|
| `rbac_permission_denied` | Fix missing Role permissions | Medium | 20min |

### Workloads (26 labs)

| ID | Lab | Difficulty | Time |
|----|-----|-----------|------|
| `pod_crashloop` | Debug CrashLoopBackOff | Easy | 15min |
| `image_pull_backoff` | Fix image name typo | Easy | 10min |
| `statefulset_broken` | Fix StatefulSet configuration | Medium | 25min |
| `daemonset_not_scheduled` | Fix DaemonSet scheduling | Medium | 20min |
| `oomkilled_limits` | Fix pods OOMKilled by low memory limits | Easy | 15min |
| `liveness_probe_flap` | Fix misconfigured liveness probes | Medium | 20min |
| `init_container_fail` | Debug failed init container | Medium | 20min |
| `resource_quota_block` | Fix pods blocked by ResourceQuota | Medium | 20min |
| `container_image_tag_wrong` | Fix non-existent image tag | Easy | 10min |
| `container_command_wrong` | Fix container command causing CrashLoop | Easy | 10min |
| `configmap_wrong_key` | Fix ConfigMap key reference mismatch | Easy | 10min |
| `env_var_missing` | Add missing environment variable | Easy | 10min |
| `hpa_not_working` | Fix HPA target reference | Medium | 20min |
| `deployment_rolling_update_stuck` | Fix stuck rolling update | Medium | 20min |
| `deployment_replicas_mismatch` | Fix readiness probe for full replicas | Medium | 20min |
| `liveness_probe_wrong` | Fix wrong liveness probe port | Medium | 15min |
| `readiness_probe_wrong` | Fix wrong readiness probe path | Medium | 15min |
| `cronjob_failed` | Fix broken CronJob image | Medium | 20min |
| `pod_oomkilled_memory` | Increase memory limits for Redis | Easy | 10min |
| `pod_stuck_in_init` | Fix failing init container | Medium | 15min |
| `resource_quota_exceeded` | Clean up or increase ResourceQuota | Medium | 20min |
| `pod_missing_configmap` | Create missing ConfigMap mount | Easy | 10min |
| `image_pull_backoff_name` | Fix wrong registry image reference | Easy | 10min |
| `pod_wrong_env` | Fix wrong environment variable value | Easy | 10min |
| `daemonset_wrong_node_selector` | Fix DaemonSet nodeSelector | Medium | 15min |
| `deployment_wrong_strategy` | Change Recreate to RollingUpdate | Medium | 15min |

## Configuration

Default config (`cka-lab-runner.yaml`):

```yaml
cluster:
  provider: kind      # or k3d, minikube (auto-detected if not specified)
  name: cka-lab
  k8sVersion: v1.30.0

labs:
  defaultNamespace: lab
```

## Adding Your Own Labs

Labs are Go types that implement the `Lab` interface. Create a file in `internal/labs/`:

```go
package labs

import (
    "context"
    "fmt"
)

func init() {
    Register(&MyLab{})
}

type MyLab struct {
    BaseLab
}

func (l *MyLab) ID() string          { return "my_lab" }
func (l *MyLab) Title() string       { return "My Lab Title" }
func (l *MyLab) Category() Category  { return CategoryWorkloads }
func (l *MyLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *MyLab) EstimatedTime() int  { return 20 }
func (l *MyLab) Tags() []string      { return []string{"pods", "troubleshooting"} }

func (l *MyLab) Description() string {
    return "Problem description shown to the user"
}

func (l *MyLab) Hints() []string {
    return []string{
        "Check the pod status",
        "Look at the pod logs",
    }
}

func (l *MyLab) Break(ctx context.Context, kubeconfigPath string) error {
    manifest := `apiVersion: v1
kind: Pod
metadata:
  name: broken-pod
spec:
  containers:
  - name: nginx
    image: nginx:broken`
    return kubectlApply(ctx, kubeconfigPath, manifest)
}

func (l *MyLab) Verify(ctx context.Context, kubeconfigPath string) error {
    output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "broken-pod",
        "-o", "jsonpath={.status.phase}")
    if err != nil || output != "Running" {
        return fmt.Errorf("pod not running")
    }
    return nil
}

func (l *MyLab) SolutionSteps() []SolutionStep {
    return []SolutionStep{
        {
            Description: "Check pod status",
            Command:     "kubectl get pods",
            Notes:       "Pod should be in ImagePullBackOff",
        },
        {
            Description: "Fix the image",
            Command:     "kubectl edit pod broken-pod",
            Notes:       "Change image to nginx:alpine",
        },
    }
}
```

Rebuild with `make build` and your lab is ready.

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed examples.

## Development

```bash
make build         # Build binary for current OS
make test          # Run tests
make lint          # Run formatters and linters
make build-all     # Cross-compile for windows/linux/darwin (amd64 + arm64)
make clean         # Clean build artifacts
make ci            # Run CI checks (lint, test, build)
```

### Project Structure

```
K8S-Lab-Everything/
├── cmd/cka-lab-runner/     # CLI entry point
├── internal/
│   ├── cli/                # Terminal output formatting
│   ├── cluster/            # Cluster providers (kind/k3d/minikube)
│   ├── config/             # Configuration management
│   ├── labs/               # All 50 lab implementations
│   └── update/             # OTA auto-update system
├── scripts/
│   ├── setup.sh            # macOS/Linux automated setup
│   └── setup.ps1           # Windows automated setup
├── .github/workflows/      # CI/CD pipeline
├── Makefile                # Build automation
└── go.mod                  # Go module definition
```

## Platform Support

| Platform | Setup Script | Binary |
|----------|-------------|--------|
| macOS (Apple Silicon) | `./scripts/setup.sh` | `cka-lab-runner-darwin-arm64` |
| macOS (Intel) | `./scripts/setup.sh` | `cka-lab-runner-darwin-amd64` |
| Linux (x86_64) | `./scripts/setup.sh` | `cka-lab-runner-linux-amd64` |
| Linux (ARM64) | `./scripts/setup.sh` | `cka-lab-runner-linux-arm64` |
| Windows | `scripts\setup.ps1` | `cka-lab-runner-windows-amd64.exe` |

## Contributing

Contributions welcome! Add new labs, fix bugs, or improve docs.

1. Fork the repo
2. Create a feature branch (`git checkout -b feature/my-lab`)
3. Add your lab in `internal/labs/`
4. Run tests (`make test`)
5. Submit a PR

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed examples.

## Credits

This project is based on [**cka-lab-runner**](https://github.com/CuriousLearner/cka-lab-runner), originally created by **[CuriousLearner](https://github.com/CuriousLearner)**. All credit for the original concept, architecture, and the initial 15 labs belongs to the original author.

This fork extends the original work with additional labs, cross-platform support, and OTA auto-update.

## License

MIT License - see [LICENSE](LICENSE). Original work Copyright (c) CuriousLearner; modifications Copyright (c) WhoisMonesh.
