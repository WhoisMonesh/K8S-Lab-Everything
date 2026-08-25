package labs

import (
	"context"
	"fmt"
)

// Difficulty represents the difficulty level of a lab
type Difficulty string

const (
	DifficultyEasy   Difficulty = "easy"
	DifficultyMedium Difficulty = "medium"
	DifficultyHard   Difficulty = "hard"
)

// Category represents the type/category of a lab
type Category string

const (
	CategoryControlPlane Category = "control-plane"
	CategoryNetworking   Category = "networking"
	CategoryScheduling   Category = "scheduling"
	CategoryDNS          Category = "dns"
	CategoryStorage      Category = "storage"
	CategoryWorkloads    Category = "workloads"
	CategoryRBAC         Category = "rbac"
	CategorySecurity     Category = "security"
)

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
	Difficulty    Difficulty
	EstimatedTime int
	Tags          []string
}

// BaseLab provides default implementations for optional Lab methods
type BaseLab struct{}

func (b *BaseLab) EstimatedTime() int { return 20 }
func (b *BaseLab) Tags() []string     { return []string{} }
func (b *BaseLab) Prepare(_ context.Context, _ string) error { return nil }
func (b *BaseLab) VerifyBroken(_ context.Context, _ string) error { return nil }
func (b *BaseLab) Verify(_ context.Context, _ string) error {
	return fmt.Errorf("verify not implemented for this lab")
}

// GetInfo returns the metadata for a lab
func GetInfo(lab Lab) Info {
	return Info{
		ID:            lab.ID(),
		Title:         lab.Title(),
		Category:      lab.Category(),
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
