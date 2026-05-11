package yaml

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// SaddleConfig represents the expected structure of a Saddle YAML configuration
type SaddleConfig struct {
	Clusters map[string]ClusterConfig `yaml:"clusters"`
}

type ClusterConfig struct {
	Provider   ProviderConfig   `yaml:"provider"`
	Kubernetes KubernetesConfig `yaml:"kubernetes"`
	Rancher    RancherConfig    `yaml:"rancher,omitempty"`
	SSH        SSHConfig        `yaml:"ssh"`
	Cluster    ClusterSettings  `yaml:"cluster"`
}

type ProviderConfig struct {
	Type   string                 `yaml:"type"`
	Config map[string]interface{} `yaml:"config"`
}

type KubernetesConfig struct {
	Distribution string                 `yaml:"distribution"`
	Config       map[string]interface{} `yaml:"config"`
}

type RancherConfig struct {
	Version           string `yaml:"version,omitempty"`
	Deploy            bool   `yaml:"deploy,omitempty"`
	Prime             bool   `yaml:"prime,omitempty"`
	BootstrapPassword string `yaml:"bootstrap_password,omitempty"`
}

type SSHConfig struct {
	KeyName        string `yaml:"key_name"`
	PrivateKeyPath string `yaml:"private_key_path"`
	User           string `yaml:"user"`
}

type ClusterSettings struct {
	NodePrefix    string `yaml:"node_prefix"`
	InstanceCount int    `yaml:"instance_count"`
}

// Validate checks if the given YAML content is valid according to the Saddle schema
func Validate(content string) error {
	// Remove potential markdown code fences if present
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```yaml")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var config SaddleConfig
	if err := yaml.Unmarshal([]byte(content), &config); err != nil {
		return fmt.Errorf("invalid YAML syntax: %w", err)
	}

	// Validate required top-level structure
	if len(config.Clusters) == 0 {
		return fmt.Errorf("'clusters' map must contain at least one cluster")
	}

	// Validate each cluster
	for clusterName, cluster := range config.Clusters {
		// Validate provider
		if cluster.Provider.Type == "" {
			return fmt.Errorf("cluster '%s': provider.type is required", clusterName)
		}

		validProviders := map[string]bool{
			"aws":   true,
			"azure": true,
			"gcp":   true,
		}
		if !validProviders[cluster.Provider.Type] {
			return fmt.Errorf("cluster '%s': provider.type must be one of: aws, azure, gcp (got '%s')", clusterName, cluster.Provider.Type)
		}

		if cluster.Provider.Config == nil || len(cluster.Provider.Config) == 0 {
			return fmt.Errorf("cluster '%s': provider.config is required and must not be empty", clusterName)
		}

		// Validate kubernetes
		if cluster.Kubernetes.Distribution == "" {
			return fmt.Errorf("cluster '%s': kubernetes.distribution is required", clusterName)
		}

		validDistros := map[string]bool{
			"rke2": true,
			"k3s":  true,
			"eks":  true,
			"aks":  true,
			"gke":  true,
		}
		if !validDistros[cluster.Kubernetes.Distribution] {
			return fmt.Errorf("cluster '%s': kubernetes.distribution must be one of: rke2, k3s, eks, aks, gke (got '%s')", clusterName, cluster.Kubernetes.Distribution)
		}

		if cluster.Kubernetes.Config == nil || len(cluster.Kubernetes.Config) == 0 {
			return fmt.Errorf("cluster '%s': kubernetes.config is required and must not be empty", clusterName)
		}

		// Validate SSH
		if cluster.SSH.KeyName == "" {
			return fmt.Errorf("cluster '%s': ssh.key_name is required", clusterName)
		}
		if cluster.SSH.PrivateKeyPath == "" {
			return fmt.Errorf("cluster '%s': ssh.private_key_path is required", clusterName)
		}
		if cluster.SSH.User == "" {
			return fmt.Errorf("cluster '%s': ssh.user is required", clusterName)
		}

		// Validate cluster settings
		if cluster.Cluster.NodePrefix == "" {
			return fmt.Errorf("cluster '%s': cluster.node_prefix is required", clusterName)
		}
		if cluster.Cluster.InstanceCount <= 0 {
			return fmt.Errorf("cluster '%s': cluster.instance_count must be greater than 0", clusterName)
		}
	}

	return nil
}
