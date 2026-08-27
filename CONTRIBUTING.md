# Contributing to CKA Lab Runner

Thank you for your interest in contributing to cka-lab-runner! This document provides guidelines and instructions for contributing.

## Ways to Contribute

1. **Add new labs** - The most valuable contribution!
2. **Improve documentation** - Help others understand the tool
3. **Fix bugs** - Report or fix issues you find
4. **Add features** - CLI features, new verification methods, etc.
5. **Write tests** - Improve code coverage

## Adding a New Lab

Adding labs is the easiest and most impactful way to contribute. Here's a complete guide.

### Step 1: Choose a Scenario

Pick a realistic CKA exam scenario:
- Control plane issues (API server, scheduler, controller-manager, etcd)
- Networking problems (CNI, NetworkPolicies, Services)
- DNS issues (CoreDNS configuration, resolution)
- Storage (PV/PVC issues, StorageClasses)
- RBAC (permissions, roles, service accounts)
- Workload issues (Deployments, StatefulSets, DaemonSets)
- Cluster upgrades
- Backup and restore

### Step 2: Determine Category and Difficulty

**Categories:**
- `CategoryControlPlane` - API server, etcd, scheduler, controller-manager
- `CategoryNetworking` - CNI, NetworkPolicies, Services, Ingress
- `CategoryDNS` - CoreDNS issues
- `CategoryScheduling` - Scheduler, taints, tolerations, affinity
- `CategoryStorage` - PV, PVC, StorageClass
- `CategoryWorkloads` - Pods, Deployments, StatefulSets, Jobs
- `CategoryRBAC` - Roles, RoleBindings, ServiceAccounts
- `CategorySecurity` - PodSecurityPolicies, SecurityContext

**Difficulty Levels:**
- `DifficultyEasy` - Single issue, obvious symptoms (10-15 min)
- `DifficultyMedium` - Multiple steps or less obvious cause (15-30 min)
- `DifficultyHard` - Complex debugging, multiple components (30+ min)

### Step 3: Create the Lab File

Create a new file in `internal/labs/` named `lab_<your_scenario>.go`:

```go
package labs

import (
	"context"
	"fmt"
	"time"
)

// Register the lab in init()
func init() {
	Register(&YourLabName{})
}

type YourLabName struct {
	BaseLab  // Embeds BaseLab for default implementations
}

func (l *YourLabName) ID() string {
	// Unique ID using snake_case
	return "your_scenario_name"
}

func (l *YourLabName) Title() string {
	// Human-readable title
	return "Your Scenario Title"
}

func (l *YourLabName) Category() Category {
	// Choose appropriate category
	return CategoryWorkloads
}

func (l *YourLabName) Difficulty() Difficulty {
	// Choose appropriate difficulty
	return DifficultyMedium
}

func (l *YourLabName) Description() string {
	// Problem statement shown to the user
	// Should describe symptoms, not the root cause
	return `Brief description of what's broken.
Users will see this and symptoms.

Your task: Fix the issue described above.`
}

func (l *YourLabName) Hints() []string {
	// Progressive hints (general to specific)
	return []string{
		"General hint about where to look",
		"More specific hint about what component",
		"Hint about which commands to use",
		"Very specific hint about the root cause",
	}
}

func (l *YourLabName) EstimatedTime() int {
	// Estimated completion time in minutes
	return 20
}

func (l *YourLabName) Tags() []string {
	// Searchable tags for this lab
	return []string{"tag1", "tag2", "troubleshooting"}
}

func (l *YourLabName) Prepare(ctx context.Context, kubeconfigPath string) error {
	// Optional: Set up any baseline state
	// Example: Create a namespace, deploy workloads

	// Wait for cluster to be ready
	for i := 0; i < 30; i++ {
		_, err := kubectl(ctx, kubeconfigPath, "get", "nodes")
		if err == nil {
			return nil
		}
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("cluster did not become ready in time")
}

func (l *YourLabName) Break(ctx context.Context, kubeconfigPath string) error {
	// Apply the broken scenario
	// This is where you introduce the problem

	// Example 1: Apply a broken manifest
	brokenManifest := `apiVersion: v1
kind: ConfigMap
metadata:
  name: broken-config
  namespace: default
data:
  config: "invalid configuration"
`
	if err := kubectlApply(ctx, kubeconfigPath, brokenManifest); err != nil {
		return fmt.Errorf("applying broken manifest: %w", err)
	}

	// Example 2: Modify a file in a kind node
	nodeName, err := getControlPlaneNode(ctx, kubeconfigPath)
	if err != nil {
		return err
	}

	_, err = dockerExec(ctx, nodeName, "sed", "-i", "s/working/broken/g", "/etc/some/config")
	if err != nil {
		return fmt.Errorf("breaking config: %w", err)
	}

	return nil
}

func (l *YourLabName) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	// Optional: Verify the lab is actually broken
	// This helps catch issues in the Break() implementation

	time.Sleep(10 * time.Second)

	// Example: Check that pods are failing
	output, _ := kubectl(ctx, kubeconfigPath, "get", "pods", "-o", "jsonpath={.items[*].status.phase}")
	// Could verify output contains "CrashLoopBackOff" or similar

	return nil
}

func (l *YourLabName) Verify(ctx context.Context, kubeconfigPath string) error {
	// Check if the user fixed the issue correctly
	// This enables the 'lab verify' command

	output, err := kubectl(ctx, kubeconfigPath, "get", "pods", "-o", "jsonpath={.items[*].status.phase}")
	if err != nil {
		return fmt.Errorf("failed to check pods: %w", err)
	}

	if !strings.Contains(output, "Running") {
		return fmt.Errorf("pods are not running yet")
	}

	return nil
}

func (l *YourLabName) SolutionSteps() []SolutionStep {
	// Step-by-step solution
	// Each step should be actionable
	return []SolutionStep{
		{
			Description: "Check the symptom",
			Command:     "kubectl get pods",
			Notes:       "Notice the pods are in a bad state",
		},
		{
			Description: "Investigate the root cause",
			Command:     "kubectl describe pod <pod-name>",
			Notes:       "Look for error messages in events",
		},
		{
			Description: "Check logs",
			Command:     "kubectl logs <pod-name>",
			Notes:       "Error message should indicate the problem",
		},
		{
			Description: "Fix the issue",
			Command:     "kubectl edit configmap broken-config",
			Notes:       "Update the configuration to be valid",
		},
		{
			Description: "Verify the fix",
			Command:     "kubectl get pods",
			Notes:       "Pods should now be running",
		},
	}
}
```

**Important Notes:**
- **BaseLab**: By embedding `BaseLab`, your lab automatically gets default implementations for `EstimatedTime()` (returns 20), `Tags()` (returns empty array), and `Verify()` (returns error). You can override any of these by implementing the method in your lab struct.
- **Optional Methods**: `Prepare()` and `VerifyBroken()` are optional. If you don't need them, you can omit them entirely.
- **Verify() Method**: Implement this if you want users to be able to use `cka-lab-runner lab verify <lab-id>` to check their fix automatically. If not implemented, users will see an error message saying verification is not available for this lab.

### Step 4: Test Your Lab

```bash
# Build the binary
go build -o cka-lab-runner ./cmd/cka-lab-runner

# Create a cluster
./cka-lab-runner init
./cka-lab-runner up

# Test your lab
./cka-lab-runner lab list | grep your_scenario_name
./cka-lab-runner lab run your_scenario_name

# Verify it breaks as expected
# Then manually fix it following your solution

# View the solution
./cka-lab-runner lab solution your_scenario_name

# Clean up
./cka-lab-runner down
```

### Step 5: Add Tests (Optional)

Add a test in `internal/labs/registry_test.go`:

```go
func TestYourLabRegistered(t *testing.T) {
	lab, err := Get("your_scenario_name")
	if err != nil {
		t.Fatalf("lab not registered: %v", err)
	}

	if lab.Category() != CategoryWorkloads {
		t.Errorf("unexpected category: %s", lab.Category())
	}

	if lab.Difficulty() != DifficultyMedium {
		t.Errorf("unexpected difficulty: %s", lab.Difficulty())
	}

	steps := lab.SolutionSteps()
	if len(steps) < 3 {
		t.Error("solution should have at least 3 steps")
	}
}
```

### Step 6: Submit a Pull Request

1. Fork the repository
2. Create a feature branch: `git checkout -b lab/your-scenario-name`
3. Commit your changes: `git commit -am "feat: add your_scenario_name lab"`
4. Push to the branch: `git push origin lab/your-scenario-name`
5. Create a Pull Request

## Lab Best Practices

### DO:
- ✅ Make scenarios realistic (things that happen in real CKA exams)
- ✅ Provide progressive hints (general → specific)
- ✅ Write clear, step-by-step solutions
- ✅ Test the lab multiple times before submitting
- ✅ Use descriptive IDs and titles
- ✅ Include helpful notes in solution steps
- ✅ Make the broken state obvious when debugging
- ✅ Add variety (different components, different issue types)

### DON'T:
- ❌ Make the problem impossible to diagnose
- ❌ Break multiple unrelated things in one lab
- ❌ Give away the answer in the description
- ❌ Create labs that require external resources
- ❌ Make labs that take >45 minutes to fix
- ❌ Forget to test the Break() method thoroughly
- ❌ Use overly complex scenarios for "easy" labs

## Helper Functions Available

In your lab implementation, you have access to:

### kubectl Helpers
```go
// Execute kubectl command
output, err := kubectl(ctx, kubeconfigPath, "get", "pods")

// Apply a YAML manifest
err := kubectlApply(ctx, kubeconfigPath, yamlString)
```

### Docker Helpers (for kind clusters)
```go
// Execute command in a container
output, err := dockerExec(ctx, containerName, "ls", "/etc/kubernetes")

// Copy files to/from container
err := dockerCp(ctx, localPath, "container:/remote/path")
```

### Utility Functions
```go
// Get the control plane node name
nodeName, err := getControlPlaneNode(ctx, kubeconfigPath)
```

## Lab Ideas Needed

We need labs for these scenarios:

### High Priority
- [x] RBAC permission denied
- [x] PersistentVolume/PersistentVolumeClaim issues
- [x] Node NotReady (kubelet stopped)
- [ ] Taint/toleration scheduling issue
- [ ] Resource quota exceeded
- [x] ImagePullBackOff
- [x] Cluster upgrade simulation
- [x] Backup and restore etcd

### Medium Priority
- [x] Ingress not routing traffic
- [ ] Service selector mismatch
- [ ] Init container failure
- [ ] ConfigMap/Secret mounting issues
- [x] DaemonSet not scheduling on all nodes
- [x] StatefulSet pod not starting
- [ ] Horizontal Pod Autoscaler not scaling
- [ ] Node affinity preventing scheduling

### Advanced
- [ ] CNI plugin failure
- [x] Certificate expiration
- [ ] Controller manager not running
- [ ] Multiple control plane nodes down
- [ ] Split-brain scenario

### New Ideas Welcome!
- [ ] Job failures and debugging
- [ ] CronJob scheduling issues
- [ ] Pod Security Standards violations
- [ ] Resource limits causing OOMKilled
- [ ] Liveness/readiness probe misconfiguration

## Code Style

- Follow standard Go conventions
- Run `go fmt` before committing
- Run `go vet` to catch issues
- Keep functions focused and small
- Add comments for complex logic
- Use meaningful variable names

## Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/labs/
```

## Documentation

When adding features:
- Update README.md if user-facing
- Update this CONTRIBUTING.md if it affects contributors
- Add inline comments for complex code

## Questions?

- Open an issue for questions
- Check existing issues for similar questions
- Join discussions in pull requests

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

Thank you for contributing to cka-lab-runner! 🎉
