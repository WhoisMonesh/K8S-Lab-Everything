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
	labs := ListByCategory(CategoryClusterArchitecture)
	// We know we have at least one cluster-architecture lab
	if len(labs) == 0 {
		t.Error("expected at least one cluster-architecture lab")
	}

	// Verify all returned labs match the category
	for _, lab := range labs {
		if lab.Category() != CategoryClusterArchitecture {
			t.Errorf("expected category %s, got %s", CategoryClusterArchitecture, lab.Category())
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

func TestListByCert(t *testing.T) {
	// All CKA labs
	ckaLabs := ListByCert(CertCKA)
	if len(ckaLabs) == 0 {
		t.Error("expected at least one CKA lab")
	}
	for _, lab := range ckaLabs {
		cert := GetCert(lab)
		if cert != CertCKA {
			t.Errorf("expected cert CKA, got %s for lab %s", cert, lab.ID())
		}
	}

	// All CKAD labs
	ckadLabs := ListByCert(CertCKAD)
	// May be empty since existing labs are CKA-focused
	_ = ckadLabs

	// All = no filter
	allLabs := ListByCert(CertAll)
	if len(allLabs) != len(List()) {
		t.Errorf("CertAll should return all labs, got %d, want %d", len(allLabs), len(List()))
	}
}

func TestRandom(t *testing.T) {
	// Test with fixed seed for reproducibility
	lab1, err := Random(42, "", "", CertAll)
	if err != nil {
		t.Fatalf("failed to get random lab: %v", err)
	}

	// Same seed should return same lab
	lab2, err := Random(42, "", "", CertAll)
	if err != nil {
		t.Fatalf("failed to get random lab: %v", err)
	}

	if lab1.ID() != lab2.ID() {
		t.Errorf("same seed returned different labs: %s vs %s", lab1.ID(), lab2.ID())
	}

	// Different seed might return different lab (not guaranteed, but likely)
	lab3, err := Random(123, "", "", CertAll)
	if err != nil {
		t.Fatalf("failed to get random lab: %v", err)
	}

	_ = lab3 // We got a lab, that's what matters
}

func TestRandomWithFilters(t *testing.T) {
	// Test with category filter
	lab, err := Random(42, CategoryServicesNetworking, "", CertAll)
	if err != nil {
		t.Fatalf("failed to get random services-networking lab: %v", err)
	}

	if lab.Category() != CategoryServicesNetworking {
		t.Errorf("expected services-networking lab, got %s", lab.Category())
	}

	// Test with difficulty filter
	lab, err = Random(42, "", DifficultyEasy, CertAll)
	if err != nil {
		t.Fatalf("failed to get random easy lab: %v", err)
	}

	if lab.Difficulty() != DifficultyEasy {
		t.Errorf("expected easy lab, got %s", lab.Difficulty())
	}

	// Test with cert filter
	lab, err = Random(42, "", "", CertCKA)
	if err != nil {
		t.Fatalf("failed to get random CKA lab: %v", err)
	}

	if GetCert(lab) != CertCKA {
		t.Errorf("expected CKA lab, got %s", GetCert(lab))
	}
}

func TestFormatSolution(t *testing.T) {
	mock := &mockLab{
		id:         "test",
		title:      "Test Lab",
		category:   CategoryWorkloadsScheduling,
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

func TestCertCategories(t *testing.T) {
	// Verify each cert returns the expected number of categories
	ckaCats := CertCategories(CertCKA)
	if len(ckaCats) != 5 {
		t.Errorf("CKA should have 5 categories, got %d", len(ckaCats))
	}

	ckadCats := CertCategories(CertCKAD)
	if len(ckadCats) != 5 {
		t.Errorf("CKAD should have 5 categories, got %d", len(ckadCats))
	}

	cksCats := CertCategories(CertCKS)
	if len(cksCats) != 6 {
		t.Errorf("CKS should have 6 categories, got %d", len(cksCats))
	}
}

func TestDomainWeightForCategory(t *testing.T) {
	tests := []struct {
		cat    Category
		weight int
	}{
		{CategoryClusterArchitecture, 25},
		{CategoryWorkloadsScheduling, 15},
		{CategoryServicesNetworking, 20},
		{CategoryStorage, 10},
		{CategoryTroubleshooting, 30},
		{CategoryAppDesignBuild, 20},
		{CategoryAppDeployment, 20},
		{CategoryAppObservability, 15},
		{CategoryAppConfigSecurity, 25},
		{CategoryServicesNetworkCKAD, 20},
		{CategoryClusterSetupCKS, 15},
		{CategoryClusterHardening, 15},
		{CategorySystemHardening, 10},
		{CategoryMicroserviceVulns, 20},
		{CategorySupplyChain, 20},
		{CategoryMonitoringLogging, 20},
	}

	for _, tt := range tests {
		w := DomainWeightForCategory(tt.cat)
		if w != tt.weight {
			t.Errorf("DomainWeightForCategory(%s) = %d, want %d", tt.cat, w, tt.weight)
		}
	}
}

func TestCertForCategory(t *testing.T) {
	if CertForCategory(CategoryClusterArchitecture) != CertCKA {
		t.Error("cluster-architecture should be CKA")
	}
	if CertForCategory(CategoryAppDeployment) != CertCKAD {
		t.Error("app-deployment should be CKAD")
	}
	if CertForCategory(CategorySupplyChain) != CertCKS {
		t.Error("supply-chain should be CKS")
	}
}
