package labs

import (
	"strings"
	"testing"
)

// TestAllLabsRegistrationMetadata validates that every registered lab provides
// complete, non-empty metadata so it can be listed without panicking.
func TestAllLabsRegistrationMetadata(t *testing.T) {
	all := List()
	if len(all) == 0 {
		t.Fatal("no labs registered")
	}
	for _, lab := range all {
		if lab.ID() == "" {
			t.Errorf("lab %T has empty ID", lab)
		}
		if strings.TrimSpace(lab.Title()) == "" {
			t.Errorf("lab %q has empty Title", lab.ID())
		}
		if strings.TrimSpace(lab.Description()) == "" {
			t.Errorf("lab %q has empty Description", lab.ID())
		}
		if lab.Category() == "" {
			t.Errorf("lab %q has empty Category", lab.ID())
		}
		if lab.Difficulty() == "" {
			t.Errorf("lab %q has empty Difficulty", lab.ID())
		}
		if lab.EstimatedTime() <= 0 {
			t.Errorf("lab %q has non-positive EstimatedTime", lab.ID())
		}
	}
}

// TestAllLabsSolutionSteps ensures every lab ships an actionable solution.
func TestAllLabsSolutionSteps(t *testing.T) {
	for _, lab := range List() {
		steps := lab.SolutionSteps()
		if len(steps) == 0 {
			t.Errorf("lab %q has no solution steps", lab.ID())
			continue
		}
		for i, s := range steps {
			if strings.TrimSpace(s.Description) == "" && strings.TrimSpace(s.Command) == "" {
				t.Errorf("lab %q solution step %d is empty", lab.ID(), i)
			}
		}
	}
}

// TestAllLabsMetadataNoPanics ensures metadata accessors never panic at list time.
func TestAllLabsMetadataNoPanics(t *testing.T) {
	for _, lab := range List() {
		_ = GetCert(lab)
		_ = GetDomainWeight(lab)
		_ = GetInfo(lab)
	}
}
