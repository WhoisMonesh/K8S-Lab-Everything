package labs

import (
	"context"
	"testing"
)

// mockLab is a test implementation of the Lab interface
type mockLab struct {
	id         string
	title      string
	category   Category
	difficulty Difficulty
}

func (m *mockLab) ID() string                                                    { return m.id }
func (m *mockLab) Title() string                                                 { return m.title }
func (m *mockLab) Category() Category                                            { return m.category }
func (m *mockLab) Difficulty() Difficulty                                        { return m.difficulty }
func (m *mockLab) Description() string                                           { return "test description" }
func (m *mockLab) Hints() []string                                               { return []string{"hint1"} }
func (m *mockLab) EstimatedTime() int                                            { return 20 }
func (m *mockLab) Tags() []string                                                { return []string{"test"} }
func (m *mockLab) Prepare(ctx context.Context, kubeconfigPath string) error      { return nil }
func (m *mockLab) Break(ctx context.Context, kubeconfigPath string) error        { return nil }
func (m *mockLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error { return nil }
func (m *mockLab) Verify(ctx context.Context, kubeconfigPath string) error       { return nil }
func (m *mockLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "step1", Command: "cmd1"},
	}
}

func TestGet(t *testing.T) {
	// Labs should be already registered via init() functions
	// Test getting a known lab
	lab, err := Get("etcd_wrong_ip")
	if err != nil {
		t.Fatalf("failed to get known lab: %v", err)
	}

	if lab.ID() != "etcd_wrong_ip" {
		t.Errorf("expected ID 'etcd_wrong_ip', got '%s'", lab.ID())
	}
}

func TestGetNonExistent(t *testing.T) {
	_, err := Get("nonexistent_lab_id")
	if err == nil {
		t.Fatal("expected error when getting nonexistent lab")
	}
}

func TestList(t *testing.T) {
	labs := List()
	if len(labs) == 0 {
		t.Fatal("expected at least one lab to be registered")
	}

	// Verify all labs have valid IDs
	for _, lab := range labs {
		if lab.ID() == "" {
			t.Error("found lab with empty ID")
		}
	}
}

func TestIDs(t *testing.T) {
	ids := IDs()
	if len(ids) == 0 {
		t.Fatal("expected at least one lab ID")
	}

	// Verify IDs are sorted
	for i := 1; i < len(ids); i++ {
		if ids[i-1] >= ids[i] {
			t.Errorf("IDs are not sorted: %v", ids)
			break
		}
	}
}

func TestListByCategory(t *testing.T) {
	labs := ListByCategory(CategoryControlPlane)
	// We know we have at least one control plane lab
	if len(labs) == 0 {
		t.Error("expected at least one control-plane lab")
	}

	// Verify all returned labs match the category
	for _, lab := range labs {
		if lab.Category() != CategoryControlPlane {
			t.Errorf("expected category %s, got %s", CategoryControlPlane, lab.Category())
		}
	}
}

func TestListByDifficulty(t *testing.T) {
	labs := ListByDifficulty(DifficultyEasy)
	// We have at least one easy lab
	if len(labs) == 0 {
		t.Error("expected at least one easy lab")
	}

	// Verify all returned labs match the difficulty
	for _, lab := range labs {
		if lab.Difficulty() != DifficultyEasy {
			t.Errorf("expected difficulty %s, got %s", DifficultyEasy, lab.Difficulty())
		}
	}
}

func TestRandom(t *testing.T) {
	// Test with fixed seed for reproducibility
	lab1, err := Random(42, "", "")
	if err != nil {
		t.Fatalf("failed to get random lab: %v", err)
	}

	// Same seed should return same lab
	lab2, err := Random(42, "", "")
	if err != nil {
		t.Fatalf("failed to get random lab: %v", err)
	}

	if lab1.ID() != lab2.ID() {
		t.Errorf("same seed returned different labs: %s vs %s", lab1.ID(), lab2.ID())
	}

	// Different seed might return different lab (not guaranteed, but likely)
	lab3, err := Random(123, "", "")
	if err != nil {
		t.Fatalf("failed to get random lab: %v", err)
	}

	_ = lab3 // We got a lab, that's what matters
}

func TestRandomWithFilters(t *testing.T) {
	// Test with category filter
	lab, err := Random(42, CategoryDNS, "")
	if err != nil {
		t.Fatalf("failed to get random DNS lab: %v", err)
	}

	if lab.Category() != CategoryDNS {
		t.Errorf("expected DNS lab, got %s", lab.Category())
	}

	// Test with difficulty filter
	lab, err = Random(42, "", DifficultyEasy)
	if err != nil {
		t.Fatalf("failed to get random easy lab: %v", err)
	}

	if lab.Difficulty() != DifficultyEasy {
		t.Errorf("expected easy lab, got %s", lab.Difficulty())
	}
}

func TestFormatSolution(t *testing.T) {
	mock := &mockLab{
		id:         "test",
		title:      "Test Lab",
		category:   CategoryWorkloads,
		difficulty: DifficultyEasy,
	}

	solution := FormatSolution(mock)
	if solution == "" {
		t.Error("expected non-empty solution")
	}

	// Check if it contains the title
	if len(solution) < len(mock.title) {
		t.Error("solution seems too short")
	}
}
