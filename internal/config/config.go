package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const DefaultConfigFile = "cka-lab-runner.yaml"

// Config represents the main configuration for cka-lab-runner
type Config struct {
	Cluster ClusterConfig `yaml:"cluster"`
	Labs    LabsConfig    `yaml:"labs"`
}

// ClusterConfig holds cluster-related settings
type ClusterConfig struct {
	// Provider specifies the cluster provider (kind, k3d, minikube)
	Provider string `yaml:"provider"`
	// Name is the cluster name
	Name string `yaml:"name"`
	// KubernetesVersion specifies the Kubernetes version to use
	KubernetesVersion string `yaml:"k8sVersion"`
}

// LabsConfig holds lab-related settings
type LabsConfig struct {
	// DefaultNamespace is the namespace to use for lab resources
	DefaultNamespace string `yaml:"defaultNamespace"`
}

// Default returns a config with sensible defaults
func Default() *Config {
	return &Config{
		Cluster: ClusterConfig{
			Provider:          "kind",
			Name:              "cka-lab",
			KubernetesVersion: "v1.35.0",
		},
		Labs: LabsConfig{
			DefaultNamespace: "lab",
		},
	}
}

// Load reads the config from the specified file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	return &cfg, nil
}

// Save writes the config to the specified file
func Save(cfg *Config, path string) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	// Create parent directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	// Write with helpful comments
	fullData := []byte(configHeader + string(data))
	if err := os.WriteFile(path, fullData, 0644); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	return nil
}

const configHeader = `# cka-lab-runner configuration file
#
# This file configures the local Kubernetes cluster and lab environment
# for CKA (Certified Kubernetes Administrator) practice.

`
