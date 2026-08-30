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
	// Workers is the number of worker nodes (0 = single-node, default)
	Workers int `yaml:"workers,omitempty"`
}

// LabsConfig holds lab-related settings
type LabsConfig struct {
	// DefaultNamespace is the namespace to use for lab resources
	DefaultNamespace string `yaml:"defaultNamespace"`
	// IsolateNamespaces spawns each lab in its own namespaced workspace
	IsolateNamespaces bool `yaml:"isolate,omitempty"`
	// AutoBreak re-applies the broken scenario when `up` runs (used with lab resume)
	AutoBreak bool `yaml:"autoBreak,omitempty"`
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

// Set updates a config value by dot-separated path (e.g. "cluster.provider", "labs.isolate")
// and returns the updated config. Shared keys map to the `name` cluster field.
func Set(cfg *Config, key, value string) error {
	switch key {
	case "cluster.provider":
		cfg.Cluster.Provider = value
	case "cluster.name", "name":
		cfg.Cluster.Name = value
	case "cluster.k8sVersion", "cluster.version", "version":
		cfg.Cluster.KubernetesVersion = value
	case "cluster.workers", "workers":
		var w int
		if _, err := fmt.Sscanf(value, "%d", &w); err != nil {
			return fmt.Errorf("workers must be an integer: %w", err)
		}
		cfg.Cluster.Workers = w
	case "labs.defaultNamespace", "labs.namespace":
		cfg.Labs.DefaultNamespace = value
	case "labs.isolate":
		b := value == "true" || value == "1" || value == "yes"
		cfg.Labs.IsolateNamespaces = b
	case "labs.autoBreak", "autobreak":
		b := value == "true" || value == "1" || value == "yes"
		cfg.Labs.AutoBreak = b
	default:
		return fmt.Errorf("unknown config key %q (try cluster.provider, cluster.name, cluster.k8sVersion, cluster.workers, labs.defaultNamespace, labs.isolate)", key)
	}
	return nil
}

// Get returns the value of a config key by dot-separated path.
func Get(cfg *Config, key string) (string, error) {
	switch key {
	case "cluster.provider":
		return cfg.Cluster.Provider, nil
	case "cluster.name", "name":
		return cfg.Cluster.Name, nil
	case "cluster.k8sVersion", "cluster.version", "version":
		return cfg.Cluster.KubernetesVersion, nil
	case "cluster.workers", "workers":
		return fmt.Sprintf("%d", cfg.Cluster.Workers), nil
	case "labs.defaultNamespace", "labs.namespace":
		return cfg.Labs.DefaultNamespace, nil
	case "labs.isolate":
		return fmt.Sprintf("%t", cfg.Labs.IsolateNamespaces), nil
	case "labs.autoBreak", "autobreak":
		return fmt.Sprintf("%t", cfg.Labs.AutoBreak), nil
	default:
		return "", fmt.Errorf("unknown config key %q", key)
	}
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
