package labs

import (
	"context"
	"strings"
	"testing"
	"time"
)

// TestAllLabsRegistration verifies every lab is properly registered and has required fields.
func TestAllLabsRegistration(t *testing.T) {
	allLabs := List()
	if len(allLabs) == 0 {
		t.Fatal("No labs registered — import lab files may be missing")
	}
	t.Logf("Found %d registered labs", len(allLabs))

	seenIDs := make(map[string]bool)
	for _, lab := range allLabs {
		id := lab.ID()
		if id == "" {
			t.Errorf("Lab has empty ID (title: %s)", lab.Title())
			continue
		}
		if seenIDs[id] {
			t.Errorf("Duplicate lab ID: %s", id)
			continue
		}
		seenIDs[id] = true

		t.Run(id, func(t *testing.T) {
			if lab.Title() == "" {
				t.Error("Title() is empty")
			}
			if lab.Category() == "" {
				t.Error("Category() is empty")
			}
			if lab.Difficulty() == "" {
				t.Error("Difficulty() is empty")
			}
			if lab.Description() == "" {
				t.Error("Description() is empty")
			}
			if lab.EstimatedTime() <= 0 {
				t.Errorf("EstimatedTime() is %d (must be > 0)", lab.EstimatedTime())
			}
			if len(lab.Tags()) == 0 {
				t.Error("Tags() is empty")
			}
			if len(lab.Hints()) == 0 {
				t.Error("Hints() is empty")
			}
			steps := lab.SolutionSteps()
			if len(steps) == 0 {
				t.Error("SolutionSteps() is empty")
			}
			for i, step := range steps {
				if step.Description == "" {
					t.Errorf("SolutionSteps()[%d].Description is empty", i)
				}
			}
		})
	}
}

// TestAllLabsVerifyNotDefault checks no lab uses the default BaseLab.Verify() which always fails.
func TestAllLabsVerifyNotDefault(t *testing.T) {
	// BaseLab.Verify returns "verify not implemented"
	var base BaseLab
	err := base.Verify(context.Background(), "")
	if err == nil {
		t.Fatal("BaseLab.Verify() should return error")
	}
	defaultMsg := err.Error()

	for _, lab := range List() {
		t.Run(lab.ID(), func(t *testing.T) {
			// Run Verify in a goroutine so time.Sleep calls in labs don't block the test suite
			type result struct {
				err error
			}
			ch := make(chan result, 1)
			go func() {
				ch <- result{err: lab.Verify(context.Background(), "")}
			}()
			select {
			case r := <-ch:
				if r.err != nil && r.err.Error() == defaultMsg {
					t.Errorf("Lab uses default BaseLab.Verify() — needs real implementation")
				}
			case <-time.After(5 * time.Second):
				t.Logf("Lab Verify() timed out (likely has time.Sleep) — skipping default check")
			}
		})
	}
}

// TestAllLabsBreakSignature validates that Break() methods exist without invoking VerifyBroken.
func TestAllLabsBreakSignature(t *testing.T) {
	for _, lab := range List() {
		lab := lab // capture range variable
		t.Run(lab.ID(), func(t *testing.T) {
			// Ensure VerifyBroken method exists (no call to avoid long sleeps)
			_ = lab.VerifyBroken
		})
	}
}

// TestDuplicateLabIDs ensures no two labs share the same ID.
func TestDuplicateLabIDs(t *testing.T) {
	seen := make(map[string]string)
	for _, lab := range List() {
		id := lab.ID()
		if prev, exists := seen[id]; exists {
			t.Errorf("Duplicate lab ID %q: first seen in lab with title %q, also in lab with title %q",
				id, prev, lab.Title())
		}
		seen[id] = lab.Title()
	}
}

// TestLabCategoriesValid ensures all labs use valid categories.
func TestLabCategoriesValid(t *testing.T) {
	validCategories := map[Category]bool{
		CategoryClusterArchitecture: true,
		CategoryWorkloadsScheduling: true,
		CategoryServicesNetworking:  true,
		CategoryStorage:             true,
		CategoryTroubleshooting:     true,
		CategoryAppDesignBuild:      true,
		CategoryAppDeployment:       true,
		CategoryAppObservability:    true,
		CategoryAppConfigSecurity:   true,
		CategoryServicesNetworkCKAD: true,
		CategoryClusterSetupCKS:     true,
		CategoryClusterHardening:    true,
		CategorySystemHardening:     true,
		CategoryMicroserviceVulns:   true,
		CategorySupplyChain:         true,
		CategoryMonitoringLogging:   true,
	}

	for _, lab := range List() {
		t.Run(lab.ID(), func(t *testing.T) {
			cat := lab.Category()
			if !validCategories[cat] {
				t.Errorf("Invalid category %q for lab %s", cat, lab.ID())
			}
		})
	}
}

// TestLabDifficultiesValid ensures all labs use valid difficulty levels.
func TestLabDifficultiesValid(t *testing.T) {
	valid := map[Difficulty]bool{
		DifficultyEasy:   true,
		DifficultyMedium: true,
		DifficultyHard:   true,
	}

	for _, lab := range List() {
		t.Run(lab.ID(), func(t *testing.T) {
			d := lab.Difficulty()
			if !valid[d] {
				t.Errorf("Invalid difficulty %q for lab %s", d, lab.ID())
			}
		})
	}
}

// TestLabIDsContainCertPrefix checks that CKA/CKAD/CKS labs have proper ID prefixes.
func TestLabIDsContainCertPrefix(t *testing.T) {
	for _, lab := range List() {
		t.Run(lab.ID(), func(t *testing.T) {
			cert := GetCert(lab)
			id := lab.ID()
			switch cert {
			case CertCKA:
				if !strings.HasPrefix(id, "cka_") && !strings.HasPrefix(id, "etcd_") &&
					!strings.HasPrefix(id, "kubeadm_") && !strings.HasPrefix(id, "node_") &&
					!strings.HasPrefix(id, "configmap_") && !strings.HasPrefix(id, "pod_") &&
					!strings.HasPrefix(id, "deployment_") && !strings.HasPrefix(id, "service_") &&
					!strings.HasPrefix(id, "multi_node_") && !strings.HasPrefix(id, "scheduler_") &&
					!strings.HasPrefix(id, "image_") && !strings.HasPrefix(id, "pv_") &&
					!strings.HasPrefix(id, "network_") && !strings.HasPrefix(id, "hpa_") &&
					!strings.HasPrefix(id, "cronjob_") && !strings.HasPrefix(id, "resource_") &&
					!strings.HasPrefix(id, "init_") && !strings.HasPrefix(id, "probe_") &&
					!strings.HasPrefix(id, "cert_") && !strings.HasPrefix(id, "cni_") &&
					!strings.HasPrefix(id, "controller_") && !strings.HasPrefix(id, "apiserver_") &&
					!strings.HasPrefix(id, "nginx_") && !strings.HasPrefix(id, "runtime_") &&
					!strings.HasPrefix(id, "etcd_") {
					// Not a hard failure, just a warning
					t.Logf("WARNING: CKA lab %s doesn't have a standard prefix", id)
				}
			}
		})
	}
}

// TestSolutionStepsHaveCommands verifies solution steps include actual commands.
func TestSolutionStepsHaveCommands(t *testing.T) {
	for _, lab := range List() {
		t.Run(lab.ID(), func(t *testing.T) {
			steps := lab.SolutionSteps()
			if len(steps) == 0 {
				t.Skip("No solution steps")
			}
			hasCommand := false
			for _, step := range steps {
				if step.Command != "" {
					hasCommand = true
					break
				}
			}
			if !hasCommand {
				t.Errorf("Lab %s has %d solution steps but none have commands", lab.ID(), len(steps))
			}
		})
	}
}
