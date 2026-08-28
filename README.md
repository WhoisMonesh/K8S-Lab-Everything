<div align="center">

# K8S-Lab-Everything

### Master Kubernetes Troubleshooting Through Hands-On Practice

[![CI](https://github.com/WhoisMonesh/K8S-Lab-Everything/actions/workflows/ci.yaml/badge.svg)](https://github.com/WhoisMonesh/K8S-Lab-Everything/actions/workflows/ci.yaml)
[![Release](https://img.shields.io/github/v/release/WhoisMonesh/K8S-Lab-Everything?color=blue)](https://github.com/WhoisMonesh/K8S-Lab-Everything/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go)](https://go.dev)
[![Platform](https://img.shields.io/badge/Platform-macOS%20%7C%20Linux%20%7C%20Windows-lightgrey)]()
[![Labs](https://img.shields.io/badge/Labs-348-orange)](#-available-labs)

<br>

**A CLI tool that breaks Kubernetes clusters on purpose — so you can learn to fix them.**

Just like the CKA/CKAD/CKS exam: real `kubectl`, real problems, real solutions.

[Quick Start](#-quick-start) • [Available Labs](#-available-labs) • [Contributing](#-contributing)

</div>

---

## About

**K8S-Lab-Everything** is an open-source CLI tool built to help Kubernetes practitioners prepare for CKA, CKAD, and CKS certifications through hands-on practice. It creates local Kubernetes clusters and deliberately breaks them so you can learn to troubleshoot and fix real-world issues.

| | |
|---|---|
| **Author** | [Monesh Ram](https://github.com/WhoisMonesh) |
| **GitHub** | [github.com/WhoisMonesh/K8S-Lab-Everything](https://github.com/WhoisMonesh/K8S-Lab-Everything) |
| **Certifications** | CKA, CKAD, CKS |
| **Labs** | 348 hands-on scenarios |
| **Platforms** | macOS, Linux, Windows |
| **License** | MIT |

### Why K8S-Lab-Everything?

- **378 Real Labs** — Covers CKA, CKAD, and CKS exam domains with realistic failure scenarios
- **Zero Cloud Costs** — Practice entirely on your local machine using KinD
- **Exam Simulation** — Timed mode with random labs matching real exam structure
- **Instant Verification** — Know immediately if your fix is correct
- **Cross-Platform** — Works on macOS (Intel & Apple Silicon), Linux, and Windows
- **OTA Auto-Update** — Always get the latest labs and features

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

### Exam Simulation
Timed exam mode with random labs matching real CKA/CKAD/CKS structure.

</td>
<td width="50%">

### Cross-Platform
Works on macOS (Intel & Apple Silicon), Linux (x86_64 & ARM64), and Windows.

### 348 Realistic Labs
Covers all CKA, CKAD, and CKS exam domains with increasing difficulty.

### Progress Tracking
Track completions, streaks, and ratings. Export as markdown or certificate.

### Shell Completion
Bash, Zsh, Fish, and PowerShell autocompletion for all commands.

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

▸  New version available: 3.2.0 (current: 1.0.0)
   Run 'cka-lab-runner update' to install

# Install update (requires sudo for /usr/local/bin)
$ sudo cka-lab-runner update
```

## Commands

**Cluster Management**

| Command | Description |
|---------|-------------|
| `cka-lab-runner init` | Create config file |
| `cka-lab-runner up` | Create local cluster |
| `cka-lab-runner up --recreate` | Recreate existing cluster |
| `cka-lab-runner up --version v1.35.0` | Select KinD node version |
| `cka-lab-runner up --workers 2` | Create multi-node cluster |
| `cka-lab-runner down` | Delete cluster |

**Lab Operations**

| Command | Description |
|---------|-------------|
| `cka-lab-runner lab list` | List all labs |
| `cka-lab-runner lab list --cert CKA` | Filter by certification |
| `cka-lab-runner lab list --resource pod` | Filter by K8s resource |
| `cka-lab-runner lab list --search pod` | Search labs |
| `cka-lab-runner lab random` | Random lab |
| `cka-lab-runner lab run <lab-id>` | Run a lab |
| `cka-lab-runner lab verify <lab-id>` | Verify your fix |
| `cka-lab-runner lab solution <lab-id>` | Show solution |
| `cka-lab-runner lab hint <lab-id>` | Get hint |
| `cka-lab-runner lab hint <lab-id> --level 3` | Specific hint |

**Exam & Progress**

| Command | Description |
|---------|-------------|
| `cka-lab-runner lab exam --cert CKA` | Start exam simulation |
| `cka-lab-runner lab streak` | View practice streak |
| `cka-lab-runner lab rate <id> <1-5>` | Rate a lab |
| `cka-lab-runner lab status` | View progress |
| `cka-lab-runner lab stats` | View statistics |
| `cka-lab-runner lab export` | Export as JSON |
| `cka-lab-runner lab export --format markdown` | Export as markdown |
| `cka-lab-runner lab export --format certificate` | Export certificate |

**System**

| Command | Description |
|---------|-------------|
| `cka-lab-runner version` | Show current version |
| `cka-lab-runner update` | Update to latest release |
| `cka-lab-runner completion bash` | Generate shell completion |

## Available Labs

### By Certification

| Certification | Labs | Domains |
|--------------|------|---------|
| **CKA** | ~235 | Cluster Architecture, Workloads, Networking, Storage, Troubleshooting |
| **CKAD** | ~77 | App Design, Deployment, Observability, Config/Security, Networking |
| **CKS** | ~66 | Cluster Setup, Hardening, System Hardening, Microservices, Supply Chain |

### By Difficulty

| Difficulty | Labs | Best For |
|-----------|------|----------|
| **Easy** (35) | Quick wins, 10-15 min | Beginners, building confidence |
| **Medium** (82) | Real scenarios, 15-25 min | CKA exam prep |
| **Hard** (26) | Complex problems, 25-30 min | Advanced troubleshooting |

<details>
<summary>View all labs by category (click to expand)</summary>

### Control Plane (18 labs)
| ID | Lab | Difficulty |
|----|-----|-----------|
| `api_server_audit_log_disabled` | API Server Audit Logging Disabled | Hard |
| `cluster_upgrade` | Cluster upgrade simulation | Hard |
| `controller_manager_wrong_config` | Controller Manager Misconfiguration | Hard |
| `etcd_backup_restore` | etcd backup and restore | Hard |
| `etcd_wrong_ip` | Fix API server → etcd communication | Medium |
| `kubeadm_cert_renewal` | Kubeadm Certificate Expired | Hard |
| `kubelet_stopped` | Fix stopped kubelet service | Medium |
| `missing_crd_dependency` | Custom Resource fails — missing CRD | Hard |
| `namespace_finalizer_stuck` | Namespace stuck in Terminating | Hard |
| `node_cordoned` | Node cordoned — pods cannot schedule | Easy |
| `node_not_ready` | Fix kubelet on NotReady node | Medium |
| `node_pressure` | Clear disk/memory pressure on node | Hard |
| `node_registration_error` | Node Registration Error | Medium |
| `scheduler_not_running` | Debug broken kube-scheduler | Medium |
| `stray_static_pod` | Stray static pod consuming resources | Medium |

### Networking (17 labs)
| ID | Lab | Difficulty |
|----|-----|-----------|
| `external_ip_not_assigned` | External IP Not Assigned | Medium |
| `ingress_broken` | Fix Ingress configuration | Medium |
| `ingress_tls_missing` | Ingress TLS Secret Missing | Medium |
| `loadbalancer_wrong_protocol` | LoadBalancer Wrong Protocol | Medium |
| `multi_container_pod` | Fix multi-container pod communication | Medium |
| `network_policy_audit_mode` | NetworkPolicy in Audit Mode | Medium |
| `network_policy_blocking` | Fix NetworkPolicy blocking traffic | Medium |
| `network_policy_egress_blocked` | NetworkPolicy Blocks Egress | Medium |
| `networkpolicy_egress_dns_blocked` | NetworkPolicy blocks DNS resolution | Hard |
| `pod_network_connectivity` | Pod-to-Pod Network Connectivity | Hard |
| `service_clusterip_not_working` | ClusterIP Service Not Responding | Medium |
| `service_loadbalancer_pending` | LoadBalancer Service stuck Pending | Easy |
| `service_no_endpoints` | Fix Service with no endpoints | Medium |
| `service_wrong_selector` | Fix Service selector not matching pods | Easy |
| `service_wrong_targetport` | Service points to wrong targetPort | Easy |

### Workloads (54 labs)
| ID | Lab | Difficulty |
|----|-----|-----------|
| `pod_crashloop` | Debug CrashLoopBackOff | Easy |
| `image_pull_backoff` | Fix image name typo | Easy |
| `deployment_replicas_mismatch` | Fix readiness probe for full replicas | Medium |
| `deployment_rolling_update_stuck` | Fix stuck rolling update | Medium |
| `hpa_not_working` | Fix HPA target reference | Medium |
| `oomkilled_limits` | Fix pods OOMKilled by low memory limits | Easy |
| `liveness_probe_wrong` | Fix wrong liveness probe port | Medium |
| `readiness_probe_wrong` | Fix wrong readiness probe path | Medium |
| `statefulset_broken` | Fix StatefulSet configuration | Medium |
| `cronjob_failed` | Fix broken CronJob image | Medium |
| ... and 44 more | | |

### Storage (12 labs)
| ID | Lab | Difficulty |
|----|-----|-----------|
| `pvc_pending` | Debug PVC stuck in Pending | Medium |
| `pv_not_binding` | Fix PersistentVolume not binding to PVC | Medium |
| `storageclass_wrong_provisioner` | StorageClass Wrong Provisioner | Medium |
| `volume_mount_conflict` | Volume Mount Path Conflict | Medium |
| `csi_driver_not_installed` | CSI Driver Not Installed | Hard |
| ... and 7 more | | |

### Security (12 labs)
| ID | Lab | Difficulty |
|----|-----|-----------|
| `secret_missing` | Create missing Secret for pod | Easy |
| `pod_security_context` | Fix pod securityContext misconfiguration | Medium |
| `rbac_permission_denied` | Fix missing Role permissions | Medium |
| `cert_expiration` | Check certificate expiration | Hard |
| ... and 8 more | | |

</details>

## Configuration

```yaml
# cka-lab-runner.yaml
cluster:
  provider: kind      # kind | k3d | minikube (auto-detected)
  name: cka-lab
  k8sVersion: v1.35.0

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
│   ├── cli/                 Terminal output formatting, themes, TUI
│   │   ├── printer.go       Colored output, banners, lab details
│   │   ├── tui.go           Interactive Bubble Tea lab selector
│   │   ├── theme.go         Dark/light theme support
│   │   ├── exam.go          Exam simulation mode
│   │   └── completion.go    Shell completion generation
│   ├── cluster/             Cluster providers (kind/k3d/minikube)
│   ├── config/              Configuration management
│   ├── labs/                All 348 lab implementations
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

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines and lab ideas.

## Credits

Originally based on [**cka-lab-runner**](https://github.com/CuriousLearner/cka-lab-runner) by [**CuriousLearner**](https://github.com/CuriousLearner).

**K8S-Lab-Everything** extends it with:
- 348+ hands-on labs covering CKA, CKAD, and CKS
- Cross-platform support (macOS, Linux, Windows)
- Interactive TUI lab picker with search and filtering
- OTA auto-update system
- Exam simulation mode
- Progress tracking with streaks and ratings
- Shell completion for bash/zsh/fish/powershell
- Dark/light theme support
- Markdown and certificate export

## Author

**Monesh Ram** — [GitHub](https://github.com/WhoisMonesh)

## License

MIT License — see [LICENSE](LICENSE)
