# K8S-Lab-Everything (cka-lab-runner)

## Credits

This project is based on [**cka-lab-runner**](https://github.com/CuriousLearner/cka-lab-runner), originally created and developed by **[CuriousLearner](https://github.com/CuriousLearner)**. All credit for the original concept, architecture, and the initial 15 labs belongs to the original author.

Original repository: https://github.com/CuriousLearner/cka-lab-runner
Original license: MIT (see [LICENSE](LICENSE))

This fork extends the original work with additional practice labs and cross-platform (Windows / macOS / Linux) build & setup tooling, maintained by [WhoisMonesh](https://github.com/WhoisMonesh).

---

Practice for the Certified Kubernetes Administrator (CKA) exam by debugging realistic broken scenarios in a local Kubernetes cluster.

## What is this?

A CLI tool that creates broken Kubernetes scenarios for you to fix—just like the CKA exam. It sets up a local cluster, breaks something specific, and lets you practice troubleshooting. When you're done, it verifies you fixed it correctly.

## Quick Start

### Prerequisites

- Docker (Docker Desktop on Windows/macOS, Docker Engine on Linux)
- kubectl
- At least one of: kind, k3d, or minikube
- Go 1.24+ (only if building from source)

### One-Command Setup Check

Run the prerequisite checker for your platform:

**Windows (PowerShell):**
```powershell
git clone https://github.com/WhoisMonesh/K8S-Lab-Everything.git
cd K8S-Lab-Everything
powershell -ExecutionPolicy Bypass -File scripts\setup.ps1
```

**macOS / Linux:**
```bash
git clone https://github.com/WhoisMonesh/K8S-Lab-Everything.git
cd K8S-Lab-Everything
./scripts/setup.sh
```

The script checks for Go, Docker, kubectl, kind/k3d/minikube and prints exact install commands for anything missing.

### Install

**Windows (PowerShell):**
```powershell
go install ./cmd/cka-lab-runner
# Binary lands in %GOPATH%\bin (usually %USERPROFILE%\go\bin) - make sure it's on your PATH
cka-lab-runner --help
```

**macOS / Linux:**
```bash
make build
sudo mv bin/cka-lab-runner /usr/local/bin/
# or simply:
go install ./cmd/cka-lab-runner
```

Prefer a ready-made binary? Download one from GitHub Actions artifacts (`make build-all` matrix) or build locally:

```bash
make build-all   # cross-compiles for windows/linux/macos (amd64 + arm64)
```

### Run Your First Lab

```bash
# Create config and cluster
cka-lab-runner init
cka-lab-runner up

# List available labs
cka-lab-runner lab list

# Run a lab
cka-lab-runner lab run pod_crashloop

# Fix the issue using kubectl...

# Verify your fix
cka-lab-runner lab verify pod_crashloop

# View solution if needed
cka-lab-runner lab solution pod_crashloop

# Clean up
cka-lab-runner down
```

> Windows note: run everything from PowerShell or Windows Terminal. kind/k3d/minikube and Docker Desktop handle the Linux node inside a VM transparently — every lab works identically on all platforms.

## Available Labs (38)

### Control Plane
- **etcd_wrong_ip** (Medium, 25min) - Fix API server → etcd communication
- **scheduler_not_running** (Medium, 20min) - Debug broken kube-scheduler
- **cluster_upgrade** (Hard, 30min) - Cluster upgrade simulation
- **etcd_backup_restore** (Hard, 30min) - etcd backup and restore
- **kubelet_stopped** (Medium, 20min) - Fix stopped kubelet service
- **node_not_ready** (Medium, 20min) - Fix kubelet on NotReady node
- **node_pressure** (Hard, 25min) - Clear disk/memory pressure on node

### Networking
- **network_policy_blocking** (Medium, 20min) - Fix NetworkPolicy blocking traffic
- **ingress_broken** (Medium, 20min) - Fix Ingress configuration
- **service_no_endpoints** (Medium, 20min) - Fix Service with no endpoints
- **service_wrong_selector** (Easy, 10min) - Fix Service selector not matching pods
- **multi_container_pod** (Medium, 15min) - Fix multi-container pod communication

### Scheduling
- **taint_no_toleration** (Medium, 20min) - Schedule pods onto tainted nodes
- **node_affinity_mismatch** (Hard, 25min) - Fix broken node affinity selectors
- **pod_scheduling_failed** (Easy, 10min) - Fix pod nodeSelector mismatch

### DNS
- **coredns_broken_config** (Easy, 15min) - Fix CoreDNS configuration

### Storage
- **pvc_pending** (Medium, 20min) - Debug PVC stuck in Pending
- **pv_not_binding** (Medium, 20min) - Fix PersistentVolume not binding to PVC

### RBAC
- **rbac_permission_denied** (Medium, 20min) - Fix missing Role permissions

### Security
- **cert_expiration** (Hard, 25min) - Check certificate expiration
- **secret_env_broken** (Easy, 15min) - Fix app failing due to bad Secret data
- **secret_missing** (Easy, 10min) - Create missing Secret for pod
- **pod_security_context** (Medium, 15min) - Fix pod securityContext misconfiguration

### Workloads
- **pod_crashloop** (Easy, 15min) - Debug CrashLoopBackOff
- **image_pull_backoff** (Easy, 10min) - Fix image name typo
- **statefulset_broken** (Medium, 25min) - Fix StatefulSet configuration
- **daemonset_not_scheduled** (Medium, 20min) - Fix DaemonSet scheduling
- **oomkilled_limits** (Easy, 15min) - Fix pods OOMKilled by low memory limits
- **liveness_probe_flap** (Medium, 20min) - Fix misconfigured liveness probes restarting pods
- **init_container_fail** (Medium, 20min) - Debug failed init container blocking startup
- **resource_quota_block** (Medium, 20min) - Fix pods blocked by exceeded ResourceQuota
- **container_image_tag_wrong** (Easy, 10min) - Fix non-existent image tag
- **container_command_wrong** (Easy, 10min) - Fix container command causing CrashLoop
- **configmap_wrong_key** (Easy, 10min) - Fix ConfigMap key reference mismatch
- **env_var_missing** (Easy, 10min) - Add missing environment variable to pod
- **hpa_not_working** (Medium, 20min) - Fix HPA target reference
- **deployment_rolling_update_stuck** (Medium, 20min) - Fix stuck rolling update
- **deployment_replicas_mismatch** (Medium, 20min) - Fix readiness probe for full replicas
- **liveness_probe_wrong** (Medium, 15min) - Fix wrong liveness probe port
- **readiness_probe_wrong** (Medium, 15min) - Fix wrong readiness probe path
- **cronjob_failed** (Medium, 20min) - Fix broken CronJob image
- **pod_oomkilled_memory** (Easy, 10min) - Increase memory limits for Redis
- **pod_stuck_in_init** (Medium, 15min) - Fix failing init container
- **resource_quota_exceeded** (Medium, 20min) - Clean up or increase ResourceQuota

## Commands

```bash
# Setup
cka-lab-runner init                    # Create config file
cka-lab-runner up                      # Create cluster
cka-lab-runner up --recreate           # Recreate existing cluster
cka-lab-runner down                    # Delete cluster

# Update
cka-lab-runner version                 # Show current version
cka-lab-runner update                  # Update to latest release from GitHub

# Labs
cka-lab-runner lab list                           # List all labs
cka-lab-runner lab list --category networking     # Filter by category
cka-lab-runner lab list --difficulty easy         # Filter by difficulty
cka-lab-runner lab random                         # Random lab
cka-lab-runner lab random --category storage      # Random lab in category
cka-lab-runner lab run <lab-id>                   # Run a lab
cka-lab-runner lab verify <lab-id>                # Verify your fix
cka-lab-runner lab solution <lab-id>              # Show solution
```

### Auto-Update

The binary automatically checks for new versions on GitHub when you run any command.
If an update is available, you'll see a notification. To install the update:

```bash
cka-lab-runner update
```

To skip the background update check:
```bash
CKA_LAB_SKIP_UPDATE_CHECK=1 cka-lab-runner lab list
```

## Adding Your Own Labs

Labs are Go types that implement the `Lab` interface. Create a file in `internal/labs/`:

```go
package labs

import "context"

func init() {
    Register(&MyLab{})
}

type MyLab struct {
    BaseLab
}

func (l *MyLab) ID() string { return "my_lab" }
func (l *MyLab) Title() string { return "My Lab Title" }
func (l *MyLab) Category() Category { return CategoryWorkloads }
func (l *MyLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *MyLab) EstimatedTime() int { return 20 }
func (l *MyLab) Tags() []string { return []string{"pods", "troubleshooting"} }

func (l *MyLab) Description() string {
    return "Problem description"
}

func (l *MyLab) Hints() []string {
    return []string{
        "Check the pod status",
        "Look at the pod logs",
        "Check the pod configuration",
        "Fix the image tag",
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

Rebuild with `make build` (or `go build ./cmd/cka-lab-runner`) and your lab is ready.

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed examples.

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

## Development

```bash
make build         # Build binary for current OS
make test          # Run tests
make install       # Install binary (Unix; on Windows use: go install ./cmd/cka-lab-runner)
make build-all     # Cross-compile for windows/linux/darwin (amd64 + arm64)
make clean         # Clean build artifacts
```

Cross-compile output lands in `bin/` as e.g. `bin/cka-lab-runner-windows-amd64.exe`, `bin/cka-lab-runner-linux-arm64`, `bin/cka-lab-runner-darwin-arm64`.

## Contributing

Contributions welcome! Add new labs, fix bugs, or improve docs. See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT License - see [LICENSE](LICENSE). Original work Copyright (c) CuriousLearner; modifications Copyright (c) WhoisMonesh.
