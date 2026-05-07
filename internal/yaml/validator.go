package yaml

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// SaddleConfig represents the expected structure of a Saddle YAML configuration
type SaddleConfig struct {
	Cluster ClusterConfig `yaml:"cluster"`
}

type ClusterConfig struct {
	Name       string              `yaml:"name"`
	Nodes      []NodeConfig        `yaml:"nodes"`
	Networking NetworkingConfig    `yaml:"networking"`
	Applications []ApplicationConfig `yaml:"applications,omitempty"`
}

type NodeConfig struct {
	Role         string `yaml:"role"`
	InstanceType string `yaml:"instance_type"`
	Count        int    `yaml:"count,omitempty"`
}

type NetworkingConfig struct {
	Plugin string `yaml:"plugin"`
}

type ApplicationConfig struct {
	Name    string                 `yaml:"name"`
	Version string                 `yaml:"version,omitempty"`
	Config  map[string]interface{} `yaml:"config,omitempty"`
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

	// Validate required fields
	if config.Cluster.Name == "" {
		return fmt.Errorf("cluster.name is required")
	}

	if len(config.Cluster.Nodes) == 0 {
		return fmt.Errorf("cluster.nodes must contain at least one node")
	}

	// Validate each node
	for i, node := range config.Cluster.Nodes {
		if node.Role == "" {
			return fmt.Errorf("cluster.nodes[%d].role is required", i)
		}
		if node.Role != "control-plane" && node.Role != "worker" {
			return fmt.Errorf("cluster.nodes[%d].role must be 'control-plane' or 'worker', got '%s'", i, node.Role)
		}
		if node.InstanceType == "" {
			return fmt.Errorf("cluster.nodes[%d].instance_type is required", i)
		}
	}

	// Validate networking
	if config.Cluster.Networking.Plugin == "" {
		return fmt.Errorf("cluster.networking.plugin is required")
	}

	validPlugins := map[string]bool{
		"calico":  true,
		"flannel": true,
		"cilium":  true,
	}
	if !validPlugins[config.Cluster.Networking.Plugin] {
		return fmt.Errorf("cluster.networking.plugin must be one of: calico, flannel, cilium (got '%s')", config.Cluster.Networking.Plugin)
	}

	// Validate applications (if present)
	for i, app := range config.Cluster.Applications {
		if app.Name == "" {
			return fmt.Errorf("cluster.applications[%d].name is required", i)
		}
	}

	return nil
}
