package labs

import (
	"context"
	"fmt"
	"os/exec"
)

// Difficulty represents the difficulty level of a lab
type Difficulty string

const (
	DifficultyEasy   Difficulty = "easy"
	DifficultyMedium Difficulty = "medium"
	DifficultyHard   Difficulty = "hard"
)

// Cert represents the CNCF certification a lab maps to
type Cert string

const (
	CertCKA  Cert = "CKA"
	CertCKAD Cert = "CKAD"
	CertCKS  Cert = "CKS"
	CertAll  Cert = ""
)

// Category represents the exam domain category of a lab
type Category string

// CKA Domains
const (
	CategoryClusterArchitecture Category = "cluster-architecture"
	CategoryWorkloadsScheduling Category = "workloads-scheduling"
	CategoryServicesNetworking  Category = "services-networking"
	CategoryStorage             Category = "storage"
	CategoryTroubleshooting     Category = "troubleshooting"
)

// CKAD Domains
const (
	CategoryAppDesignBuild      Category = "app-design-build"
	CategoryAppDeployment       Category = "app-deployment"
	CategoryAppObservability    Category = "app-observability"
	CategoryAppConfigSecurity   Category = "app-config-security"
	CategoryServicesNetworkCKAD Category = "services-networking-ckad"
)

// CKS Domains
const (
	CategoryClusterSetupCKS   Category = "cluster-setup-cks"
	CategoryClusterHardening  Category = "cluster-hardening"
	CategorySystemHardening   Category = "system-hardening"
	CategoryMicroserviceVulns Category = "microservice-vulns"
	CategorySupplyChain       Category = "supply-chain"
	CategoryMonitoringLogging Category = "monitoring-logging"
)

// Legacy category aliases for backward compatibility
const (
	CategoryControlPlane = CategoryClusterArchitecture
	CategoryWorkloads    = CategoryWorkloadsScheduling
	CategoryNetworking   = CategoryServicesNetworking
	CategoryDNS          = CategoryServicesNetworking
	CategoryScheduling   = CategoryWorkloadsScheduling
	CategorySecurity     = CategoryAppConfigSecurity
	CategoryRBAC         = CategoryClusterHardening
)

// CertCategories returns all categories belonging to a certification
func CertCategories(cert Cert) []Category {
	switch cert {
	case CertCKA:
		return []Category{
			CategoryClusterArchitecture,
			CategoryWorkloadsScheduling,
			CategoryServicesNetworking,
			CategoryStorage,
			CategoryTroubleshooting,
		}
	case CertCKAD:
		return []Category{
			CategoryAppDesignBuild,
			CategoryAppDeployment,
			CategoryAppObservability,
			CategoryAppConfigSecurity,
			CategoryServicesNetworkCKAD,
		}
	case CertCKS:
		return []Category{
			CategoryClusterSetupCKS,
			CategoryClusterHardening,
			CategorySystemHardening,
			CategoryMicroserviceVulns,
			CategorySupplyChain,
			CategoryMonitoringLogging,
		}
	default:
		return nil
	}
}

// CertForCategory returns the certification a category belongs to
func CertForCategory(c Category) Cert {
	for _, cat := range CertCategories(CertCKA) {
		if c == cat {
			return CertCKA
		}
	}
	for _, cat := range CertCategories(CertCKAD) {
		if c == cat {
			return CertCKAD
		}
	}
	for _, cat := range CertCategories(CertCKS) {
		if c == cat {
			return CertCKS
		}
	}
	return CertAll
}

// SolutionStep represents a single step in the solution
type SolutionStep struct {
	Description string
	Command     string
	Notes       string
}

// Lab defines the interface that all labs must implement
type Lab interface {
	// ID returns the unique identifier for this lab
	ID() string

	// Title returns the human-readable title
	Title() string

	// Category returns the lab category
	Category() Category

	// Difficulty returns the difficulty level
	Difficulty() Difficulty

	// Description returns the problem statement shown to the user
	Description() string

	// Hints returns optional hints for the user
	Hints() []string

	// EstimatedTime returns the estimated time to complete (in minutes)
	EstimatedTime() int

	// Tags returns searchable tags for this lab
	Tags() []string

	// Prepare ensures the cluster is in a known baseline state (optional)
	Prepare(ctx context.Context, kubeconfigPath string) error

	// Break applies the failure/misconfiguration to the cluster
	Break(ctx context.Context, kubeconfigPath string) error

	// VerifyBroken checks that the lab is actually in a broken state (optional)
	VerifyBroken(ctx context.Context, kubeconfigPath string) error

	// Verify checks if the user has fixed the issue correctly
	Verify(ctx context.Context, kubeconfigPath string) error

	// SolutionSteps returns the step-by-step solution
	SolutionSteps() []SolutionStep
}

// Info holds metadata about a lab
type Info struct {
	ID            string
	Title         string
	Category      Category
	Cert          Cert
	DomainWeight  int
	Difficulty    Difficulty
	EstimatedTime int
	Tags          []string
}

// BaseLab provides default implementations for optional Lab methods
type BaseLab struct{}

func (b *BaseLab) EstimatedTime() int                             { return 20 }
func (b *BaseLab) Tags() []string                                 { return []string{} }
func (b *BaseLab) Prepare(_ context.Context, _ string) error      { return nil }
func (b *BaseLab) VerifyBroken(_ context.Context, _ string) error { return nil }
func (b *BaseLab) Verify(_ context.Context, _ string) error {
	return fmt.Errorf("verify not implemented for this lab")
}

// clusterName is the running cluster's name, injected by the runner so labs
// that interact with node containers (docker exec) can build node names.
var clusterName string

// SetClusterName records the active cluster name (set by the runner).
func SetClusterName(name string) {
	clusterName = name
}

// ControlPlaneNodeName returns the control-plane node container name for the
// active cluster (e.g. "<cluster>-control-plane" for kind).
func ControlPlaneNodeName() string {
	return clusterName + "-control-plane"
}

// NodeName returns the container name for a named node of the active cluster.
func NodeName(name string) string {
	return clusterName + "-" + name
}

// dockerCommand runs a command inside a node container via `docker exec` and
// returns its combined output and error. It is used for node-level operations
// that have no SSH equivalent in kind.
func dockerCommand(container, shellCommand string) (string, error) {
	cmd := exec.CommandContext(context.Background(), "docker", "exec",
		container, "sh", "-c", shellCommand)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), err
	}
	return string(out), nil
}

// GetInfo returns the metadata for a lab
func GetInfo(lab Lab) Info {
	cat := lab.Category()
	return Info{
		ID:            lab.ID(),
		Title:         lab.Title(),
		Category:      cat,
		Cert:          GetCert(lab),
		DomainWeight:  GetDomainWeight(lab),
		Difficulty:    lab.Difficulty(),
		EstimatedTime: lab.EstimatedTime(),
		Tags:          lab.Tags(),
	}
}

// FormatSolution formats the solution steps for display
func FormatSolution(lab Lab) string {
	steps := lab.SolutionSteps()
	if len(steps) == 0 {
		return "No solution steps available."
	}

	result := fmt.Sprintf("Solution for: %s\n", lab.Title())
	result += fmt.Sprintf("═══════════════════════════════════════════════════\n\n")

	for i, step := range steps {
		result += fmt.Sprintf("Step %d: %s\n", i+1, step.Description)
		if step.Command != "" {
			result += fmt.Sprintf("\n  Command:\n  $ %s\n", step.Command)
		}
		if step.Notes != "" {
			result += fmt.Sprintf("\n  Notes: %s\n", step.Notes)
		}
		result += "\n"
	}

	return result
}
