<div align="center">

# K8S-Lab-Everything

### Master Kubernetes Troubleshooting Through Hands-On Practice

[![CI](https://github.com/WhoisMonesh/K8S-Lab-Everything/actions/workflows/ci.yaml/badge.svg)](https://github.com/WhoisMonesh/K8S-Lab-Everything/actions/workflows/ci.yaml)
[![Release](https://img.shields.io/github/v/release/WhoisMonesh/K8S-Lab-Everything?color=blue)](https://github.com/WhoisMonesh/K8S-Lab-Everything/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go)](https://go.dev)
[![Platform](https://img.shields.io/badge/Platform-macOS%20%7C%20Linux%20%7C%20Windows-lightgrey)]()
[![Labs](https://img.shields.io/badge/Labs-143-orange)](#-available-labs)

<br>

**A CLI tool that breaks Kubernetes clusters on purpose — so you can learn to fix them.**

Just like the CKA exam: real `kubectl`, real problems, real solutions.

[Quick Start](#-quick-start) • [Available Labs](#-available-labs) • [Contributing](#-contributing)

</div>

---

## How It Works

```
┌─────────────┐     ┌──────────────┐     ┌─────────────────┐     ┌──────────────┐
│  cka-lab-   │────▶│  Break a     │────▶│  You debug &    │────▶│  Verify your │
│  runner up  │     │  scenario    │     │  fix with kubectl│     │  fix         │
└─────────────┘     └──────────────┘     └─────────────────┘     └──────────────┘
    Creates              Applies a            Trouble-               Checks if
    a local              realistic            shooting               you solved
    cluster              failure              practice               it correctly
```

### What happens when you run a lab:

```
$ cka-lab-runner lab run pod_crashloop

▸  Preparing lab environment...

▸  Applying broken scenario...

┌────────────────────────────────────────────────────────────────┐
│  Pod in CrashLoopBackOff                                       │
└────────────────────────────────────────────────────────────────┘

▸ Details

  ID          │ pod_crashloop
  Category    │ WORKLOADS
  Difficulty  │  EASY
  Est. Time   │ 15 min
  Tags        │ pods  crashloop  configmap  troubleshooting

▸ Description

  │ A deployment named 'webapp' in the default namespace is failing
  │ to start. The pods are stuck in CrashLoopBackOff state.

▸ Hints

  1.  ▸ Check the pod status and events
  2.  ▸ Look at the pod logs to see why it's crashing
  3.  ▸ Check the container image and command configuration

▸ Commands

  cka-lab-runner lab run pod_crashloop       # Apply the broken scenario
  cka-lab-runner lab verify pod_crashloop   # Check your fix
  cka-lab-runner lab hint pod_crashloop     # Get help
  cka-lab-runner lab solution pod_crashloop # Show full solution

✔  Lab scenario applied successfully!
```

## Features

<table>
<tr>
<td width="50%">

### One-Command Setup
Automatically installs Go, Docker, kubectl, and kind. No manual steps.

### Auto-Update
Checks for new labs on every run. Run `cka-lab-runner update` to install.

### Instant Verification
Know immediately if your fix is correct — no guessing.

</td>
<td width="50%">

### Cross-Platform
Works on macOS (Intel & Apple Silicon), Linux (x86_64 & ARM64), and Windows.

### 143 Realistic Labs
Covers all CKA exam domains with increasing difficulty levels.

### Extensible
Add your own labs in Go — just implement the `Lab` interface.

</td>
</tr>
</table>

## Quick Start

### Step 1: Clone & Setup

**macOS / Linux:**
```bash
git clone https://github.com/WhoisMonesh/K8S-Lab-Everything.git
cd K8S-Lab-Everything
./scripts/setup.sh
```

**Windows (PowerShell):**
```powershell
git clone https://github.com/WhoisMonesh/K8S-Lab-Everything.git
cd K8S-Lab-Everything
powershell -ExecutionPolicy Bypass -File scripts\setup.ps1
```

<details>
<summary>What gets installed?</summary>

| Tool | Purpose |
|------|---------|
| **Go** | Build the CLI tool |
| **Docker** | Run the local cluster |
| **kubectl** | Interact with Kubernetes |
| **kind** | Create local clusters |
| **cka-lab-runner** | The practice tool itself |

</details>

### Step 2: Run Your First Lab

```bash
# Create config and cluster
$ cka-lab-runner init
$ cka-lab-runner up

# List available labs
$ cka-lab-runner lab list

# Run a lab (breaks something on purpose)
$ cka-lab-runner lab run pod_crashloop

# Debug and fix using kubectl
$ kubectl get pods
$ kubectl describe pod -l app=webapp
$ kubectl logs -l app=webapp

# Verify your fix
$ cka-lab-runner lab verify pod_crashloop
✓ Congratulations! You successfully fixed: Pod in CrashLoopBackOff

# View solution if stuck
$ cka-lab-runner lab solution pod_crashloop

# Clean up
$ cka-lab-runner down
```

### Step 3: Auto-Update

```bash
# The binary checks for updates automatically
$ cka-lab-runner lab list

▸  New version available: 2.2.0 (current: 1.0.0)
   Run 'cka-lab-runner update' to install

# Install update (requires sudo for /usr/local/bin)
$ sudo cka-lab-runner update
Checking for latest release...
New version available: 2.2.0 (current: 1.0.0)
Downloading...
Updated successfully! (1.0.0 -> 2.2.0)
```

## Commands

<details>
<summary><b>Cluster Management</b></summary>

| Command | Description |
|---------|-------------|
| `cka-lab-runner init` | Create config file |
| `cka-lab-runner up` | Create local cluster |
| `cka-lab-runner up --recreate` | Recreate existing cluster |
| `cka-lab-runner down` | Delete cluster |

</details>

<details>
<summary><b>Lab Operations</b></summary>

| Command | Description |
|---------|-------------|
| `cka-lab-runner lab list` | List all labs |
| `cka-lab-runner lab list --category networking` | Filter by category |
| `cka-lab-runner lab list --difficulty easy` | Filter by difficulty |
| `cka-lab-runner lab random` | Random lab |
| `cka-lab-runner lab random --category storage` | Random in category |
| `cka-lab-runner lab run <lab-id>` | Run a lab |
| `cka-lab-runner lab verify <lab-id>` | Verify your fix |
| `cka-lab-runner lab solution <lab-id>` | Show solution |

</details>

<details>
<summary><b>System</b></summary>

| Command | Description |
|---------|-------------|
| `cka-lab-runner version` | Show current version |
| `cka-lab-runner update` | Update to latest release |

</details>

## Available Labs

### By Difficulty

| Difficulty | Labs | Best For |
|-----------|------|----------|
| **Easy** (35) | Quick wins, 10-15 min | Beginners, building confidence |
| **Medium** (82) | Real scenarios, 15-25 min | CKA exam prep |
| **Hard** (26) | Complex problems, 25-30 min | Advanced troubleshooting |

### Control Plane (18 labs)

| ID | Lab | Difficulty | Time |
|----|-----|-----------|------|
| `api_server_audit_log_disabled` | API Server Audit Logging Disabled | Hard | 25min |
| `cluster_upgrade` | Cluster upgrade simulation | Hard | 30min |
| `controller_manager_wrong_config` | Controller Manager Misconfiguration | Hard | 25min |
| `etcd_backup_restore` | etcd backup and restore | Hard | 30min |
| `etcd_wrong_ip` | Fix API server → etcd communication | Medium | 25min |
| `kubeadm_cert_renewal` | Kubeadm Certificate Expired | Hard | 30min |
| `kubelet_stopped` | Fix stopped kubelet service | Medium | 20min |
| `missing_crd_dependency` | Custom Resource fails — missing CRD | Hard | 20min |
| `namespace_finalizer_stuck` | Namespace stuck in Terminating | Hard | 20min |
| `node_cordoned` | Node cordoned — pods cannot schedule | Easy | 10min |
| `node_not_ready` | Fix kubelet on NotReady node | Medium | 20min |
| `node_pressure` | Clear disk/memory pressure on node | Hard | 25min |
| `node_registration_error` | Node Registration Error | Medium | 20min |
| `scheduler_not_running` | Debug broken kube-scheduler | Medium | 20min |
| `stray_static_pod` | Stray static pod consuming resources | Medium | 15min |

### Networking (17 labs)

| ID | Lab | Difficulty | Time |
|----|-----|-----------|------|
| `external_ip_not_assigned` | External IP Not Assigned | Medium | 15min |
| `ingress_broken` | Fix Ingress configuration | Medium | 20min |
| `ingress_tls_missing` | Ingress TLS Secret Missing | Medium | 20min |
| `loadbalancer_wrong_protocol` | LoadBalancer Wrong Protocol | Medium | 15min |
| `multi_container_pod` | Fix multi-container pod communication | Medium | 15min |
| `network_policy_audit_mode` | NetworkPolicy in Audit Mode | Medium | 20min |
| `network_policy_blocking` | Fix NetworkPolicy blocking traffic | Medium | 20min |
| `network_policy_egress_blocked` | NetworkPolicy Blocks Egress | Medium | 20min |
| `networkpolicy_egress_dns_blocked` | NetworkPolicy blocks DNS resolution | Hard | 20min |
| `pod_network_connectivity` | Pod-to-Pod Network Connectivity | Hard | 25min |
| `service_clusterip_not_working` | ClusterIP Service Not Responding | Medium | 15min |
| `service_loadbalancer_pending` | LoadBalancer Service stuck Pending | Easy | 10min |
| `service_no_endpoints` | Fix Service with no endpoints | Medium | 20min |
| `service_wrong_selector` | Fix Service selector not matching pods | Easy | 10min |
| `service_wrong_targetport` | Service points to wrong targetPort | Easy | 10min |

### Scheduling (17 labs)

| ID | Lab | Difficulty | Time |
|----|-----|-----------|------|
| `cm_immutable_migration` | Immutable ConfigMap migration | Hard | 20min |
| `limitrange_exceeded` | Pod rejected by LimitRange | Medium | 15min |
| `node_affinity_mismatch` | Fix broken node affinity selectors | Hard | 25min |
| `nodeselector_label_missing` | Pod Pending - Missing Node Label | Easy | 10min |
| `nodeselector_no_match` | Pod Pending — no node matches NodeSelector | Easy | 10min |
| `pod_antiaffinity_conflict` | Deployment can't schedule due to anti-affinity | Hard | 20min |
| `pod_preemption_occurred` | Low Priority Pod Preempted | Medium | 15min |
| `pod_scheduling_failed` | Fix pod nodeSelector mismatch | Easy | 10min |
| `pod_topology_spread_violation` | Pod Topology Spread Constraint Violation | Hard | 20min |
| `priorityclass_missing` | Pod uses nonexistent PriorityClass | Medium | 15min |
| `resource_request_too_high` | Pod Pending - Resource Request Too High | Easy | 10min |
| `taint_no_toleration` | Schedule pods onto tainted nodes | Medium | 20min |

### DNS (7 labs)

| ID | Lab | Difficulty | Time |
|----|-----|-----------|------|
| `coredns_broken_config` | Fix CoreDNS configuration | Easy | 15min |
| `dns_policy_wrong` | Pod cannot resolve cluster DNS | Medium | 15min |
| `dns_timeout_issues` | DNS Resolution Timeout | Medium | 20min |
| `external_dns_not_working` | External DNS Resolution Failing | Medium | 20min |
| `headless_service_dns` | Headless Service DNS Not Working | Medium | 15min |
| `hostalias_wrong_ip` | Pod /etc/hosts points to wrong IP | Medium | 15min |

### Storage (12 labs)

| ID | Lab | Difficulty | Time |
|----|-----|-----------|------|
| `csi_driver_not_installed` | CSI Driver Not Installed | Hard | 25min |
| `persistent_volume_reclaim_policy` | PV Reclaim Policy Prevents PVC Delete | Medium | 15min |
| `pod_host_path_wrong` | Fix wrong hostPath mount | Medium | 15min |
| `pv_not_binding` | Fix PersistentVolume not binding to PVC | Medium | 20min |
| `pvc_pending` | Debug PVC stuck in Pending | Medium | 20min |
| `storageclass_wrong_provisioner` | StorageClass Wrong Provisioner | Medium | 20min |
| `volume_mount_conflict` | Volume Mount Path Conflict | Medium | 15min |
| `volume_readonly_write_fail` | Pod CrashLoop — writing to read-only volume | Easy | 10min |
| `volume_snapshot_missing` | Volume Snapshot Not Found | Medium | 20min |
| `volume_subpath_missing` | Pod CrashLoop — wrong volume subPath | Medium | 15min |

### Security (12 labs)

| ID | Lab | Difficulty | Time |
|----|-----|-----------|------|
| `admission_controller_blocked` | Admission Controller Blocking Pods | Hard | 20min |
| `cert_expiration` | Check certificate expiration | Hard | 25min |
| `pod_security_context` | Fix pod securityContext misconfiguration | Medium | 15min |
| `pod_security_policy_violation` | Pod Security Policy Violation | Medium | 15min |
| `runasnonroot_rejected` | Pod rejected — runAsNonRoot violation | Medium | 15min |
| `seccomp_invalid_profile` | Pod rejected — invalid seccomp profile | Hard | 20min |
| `secret_encryption_disabled` | Secret Encryption at Rest Disabled | Hard | 25min |
| `secret_env_broken` | Fix app failing due to bad Secret data | Easy | 15min |
| `secret_missing` | Create missing Secret for pod | Easy | 10min |
| `service_account_token_expired` | Service Account Token Expired | Medium | 15min |

### RBAC (6 labs)

| ID | Lab | Difficulty | Time |
|----|-----|-----------|------|
| `cluster_role_binding_wrong` | ClusterRoleBinding References Wrong Role | Medium | 15min |
| `impersonation_header` | Impersonation Header Denied | Hard | 20min |
| `rbac_permission_denied` | Fix missing Role permissions | Medium | 20min |
| `rolebinding_wrong_role` | RoleBinding references missing Role | Medium | 15min |
| `service_account_missing_permissions` | Service Account Missing Permissions | Medium | 15min |

### Workloads (54 labs)

| ID | Lab | Difficulty | Time |
|----|-----|-----------|------|
| `bad_image_undo` | Roll back a bad image update | Medium | 15min |
| `configmap_env_from` | ConfigMap envFrom Reference Broken | Easy | 10min |
| `configmap_wrong_key` | Fix ConfigMap key reference mismatch | Easy | 10min |
| `container_command_wrong` | Fix container command causing CrashLoop | Easy | 10min |
| `container_image_tag_wrong` | Fix non-existent image tag | Easy | 10min |
| `cronjob_failed` | Fix broken CronJob image | Medium | 20min |
| `daemonset_not_scheduled` | Fix DaemonSet scheduling | Medium | 20min |
| `daemonset_wrong_node_selector` | Fix DaemonSet nodeSelector | Medium | 15min |
| `deployment_progress_deadline` | Deployment Progress Deadline Exceeded | Medium | 20min |
| `deployment_replicas_mismatch` | Fix readiness probe for full replicas | Medium | 20min |
| `deployment_rolling_update_stuck` | Fix stuck rolling update | Medium | 20min |
| `deployment_wrong_strategy` | Change Recreate to RollingUpdate | Medium | 15min |
| `downward_api_missing` | DownwardAPI Volume Missing | Medium | 15min |
| `env_var_missing` | Add missing environment variable | Easy | 10min |
| `ephemeral_container` | Debug Pod with Ephemeral Container | Medium | 20min |
| `hpa_not_working` | Fix HPA target reference | Medium | 20min |
| `image_pull_backoff` | Fix image name typo | Easy | 10min |
| `image_pull_backoff_name` | Fix wrong registry image reference | Easy | 10min |
| `init_container_fail` | Debug failed init container | Medium | 20min |
| `job_deadline_exceeded` | Job killed by activeDeadlineSeconds | Medium | 15min |
| `liveness_probe_flap` | Fix misconfigured liveness probes | Medium | 20min |
| `liveness_probe_wrong` | Fix wrong liveness probe port | Medium | 15min |
| `oomkilled_limits` | Fix pods OOMKilled by low memory limits | Easy | 15min |
| `paused_rollout_resume` | Deployment paused mid-rollout | Easy | 10min |
| `pod_crashloop` | Debug CrashLoopBackOff | Easy | 15min |
| `pod_missing_configmap` | Create missing ConfigMap mount | Easy | 10min |
| `pod_oomkilled_memory` | Increase memory limits for Redis | Easy | 10min |
| `pod_selector_no_match` | Deployment Selector Doesn't Match Labels | Medium | 15min |
| `pod_stuck_in_init` | Fix failing init container | Medium | 15min |
| `pod_wrong_env` | Fix wrong environment variable value | Easy | 10min |
| `prestop_hook_wrong` | PreStop Hook Causing Pod Termination Issues | Medium | 15min |
| `readiness_probe_wrong` | Fix wrong readiness probe path | Medium | 15min |
| `resource_quota_block` | Fix pods blocked by ResourceQuota | Medium | 20min |
| `resource_quota_exceeded` | Clean up or increase ResourceQuota | Medium | 20min |
| `rollback_revision_wrong` | Deployment Rollback to Wrong Revision | Medium | 15min |
| `sidecar_injector` | Sidecar Injection Not Working | Medium | 20min |
| `slow_pod_termination` | Pod stuck terminating | Medium | 15min |
| `startup_probe_missing` | Liveness probe kills slow-starting app | Medium | 20min |
| `statefulset_broken` | Fix StatefulSet configuration | Medium | 25min |
| `statefulset_headless_missing` | StatefulSet without headless Service | Medium | 20min |

## Configuration

```yaml
# cka-lab-runner.yaml
cluster:
  provider: kind      # kind | k3d | minikube (auto-detected)
  name: cka-lab
  k8sVersion: v1.30.0

labs:
  defaultNamespace: lab
```

## Uninstall

**macOS / Linux:**
```bash
./scripts/uninstall.sh
```

**Windows (PowerShell):**
```powershell
.\scripts\uninstall.ps1
```

Use `--dry-run` to see what would be removed without deleting.
Use `--force` to skip confirmation.

## Adding Your Own Labs

Create a new file in `internal/labs/`:

```go
package labs

import (
    "context"
    "fmt"
)

func init() { Register(&MyLab{}) }

type MyLab struct{ BaseLab }

func (l *MyLab) ID() string            { return "my_lab" }
func (l *MyLab) Title() string         { return "My Lab Title" }
func (l *MyLab) Category() Category    { return CategoryWorkloads }
func (l *MyLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *MyLab) EstimatedTime() int    { return 20 }
func (l *MyLab) Tags() []string        { return []string{"pods"} }
func (l *MyLab) Description() string   { return "Problem description" }
func (l *MyLab) Hints() []string       { return []string{"Check pod status"} }

func (l *MyLab) Break(ctx context.Context, kubeconfigPath string) error {
    return kubectlApply(ctx, kubeconfigPath, `apiVersion: v1
kind: Pod
metadata:
  name: broken-pod
spec:
  containers:
  - name: nginx
    image: nginx:broken`)
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
        {Description: "Check pod status", Command: "kubectl get pods"},
        {Description: "Fix the image", Command: "kubectl edit pod broken-pod"},
    }
}
```

Run `make build` and your lab is ready. See [CONTRIBUTING.md](CONTRIBUTING.md) for more details.

## Development

```bash
make build         # Build for current OS
make test          # Run tests
make lint          # Format and vet code
make build-all     # Cross-compile (darwin/linux/windows, amd64/arm64)
make ci            # Run all CI checks
make clean         # Remove build artifacts
```

### Project Structure

```
├── cmd/cka-lab-runner/      CLI entry point
├── internal/
│   ├── cli/                 Terminal output formatting
│   ├── cluster/             Cluster providers (kind/k3d/minikube)
│   ├── config/              Configuration management
│   ├── labs/                All 143 lab implementations
│   └── update/              OTA auto-update system
├── scripts/
│   ├── setup.sh             macOS/Linux setup
│   └── setup.ps1            Windows setup
├── .github/workflows/       CI/CD pipeline
├── Makefile                 Build automation
└── go.mod                   Go module definition
```

## Platform Support

| Platform | Setup | Binary |
|----------|-------|--------|
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

## Credits

Based on [**cka-lab-runner**](https://github.com/CuriousLearner/cka-lab-runner) by [**CuriousLearner**](https://github.com/CuriousLearner). This fork adds 143+ new labs, cross-platform support, OTA auto-update, and a modern CLI interface.

## License

MIT License — see [LICENSE](LICENSE)
