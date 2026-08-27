package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	if cfg.Cluster.Provider != "kind" {
		t.Errorf("expected provider 'kind', got '%s'", cfg.Cluster.Provider)
	}

	if cfg.Cluster.Name != "cka-lab" {
		t.Errorf("expected name 'cka-lab', got '%s'", cfg.Cluster.Name)
	}

	if cfg.Cluster.KubernetesVersion != "v1.30.0" {
		t.Errorf("expected version 'v1.30.0', got '%s'", cfg.Cluster.KubernetesVersion)
	}

	if cfg.Labs.DefaultNamespace != "lab" {
		t.Errorf("expected namespace 'lab', got '%s'", cfg.Labs.DefaultNamespace)
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.yaml")

	// Create a config
	original := Default()
	original.Cluster.Name = "test-cluster"

	// Save it
	if err := Save(original, configPath); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatalf("config file was not created")
	}

	// Load it back
	loaded, err := Load(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Verify contents
	if loaded.Cluster.Name != "test-cluster" {
		t.Errorf("expected name 'test-cluster', got '%s'", loaded.Cluster.Name)
	}

	if loaded.Cluster.Provider != original.Cluster.Provider {
		t.Errorf("provider mismatch: expected '%s', got '%s'", original.Cluster.Provider, loaded.Cluster.Provider)
	}
}

func TestLoadNonExistentFile(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("expected error when loading nonexistent file")
	}
}
