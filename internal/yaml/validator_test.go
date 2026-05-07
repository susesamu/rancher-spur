package yaml

import (
	"testing"
)

func TestValidate_ValidYAML(t *testing.T) {
	validYAML := `
cluster:
  name: test-cluster
  nodes:
    - role: control-plane
      instance_type: t3.medium
      count: 1
    - role: worker
      instance_type: t3.large
      count: 3
  networking:
    plugin: calico
  applications:
    - name: nginx
      version: 1.21.0
    - name: redis
      version: 6.2.6
`

	err := Validate(validYAML)
	if err != nil {
		t.Errorf("expected no error for valid YAML, got: %v", err)
	}
}

func TestValidate_MinimalValidYAML(t *testing.T) {
	minimalYAML := `
cluster:
  name: minimal-cluster
  nodes:
    - role: control-plane
      instance_type: t3.small
  networking:
    plugin: flannel
`

	err := Validate(minimalYAML)
	if err != nil {
		t.Errorf("expected no error for minimal valid YAML, got: %v", err)
	}
}

func TestValidate_WithMarkdownCodeFence(t *testing.T) {
	yamlWithFence := "```yaml\ncluster:\n  name: test\n  nodes:\n    - role: worker\n      instance_type: t3.medium\n  networking:\n    plugin: cilium\n```"

	err := Validate(yamlWithFence)
	if err != nil {
		t.Errorf("expected no error when stripping markdown fences, got: %v", err)
	}
}

func TestValidate_MissingClusterName(t *testing.T) {
	invalidYAML := `
cluster:
  nodes:
    - role: control-plane
      instance_type: t3.medium
  networking:
    plugin: calico
`

	err := Validate(invalidYAML)
	if err == nil {
		t.Error("expected error for missing cluster.name, got none")
	}
}

func TestValidate_NoNodes(t *testing.T) {
	invalidYAML := `
cluster:
  name: test-cluster
  networking:
    plugin: calico
`

	err := Validate(invalidYAML)
	if err == nil {
		t.Error("expected error for missing nodes, got none")
	}
}

func TestValidate_NodeMissingRole(t *testing.T) {
	invalidYAML := `
cluster:
  name: test-cluster
  nodes:
    - instance_type: t3.medium
  networking:
    plugin: calico
`

	err := Validate(invalidYAML)
	if err == nil {
		t.Error("expected error for node missing role, got none")
	}
}

func TestValidate_NodeInvalidRole(t *testing.T) {
	invalidYAML := `
cluster:
  name: test-cluster
  nodes:
    - role: invalid-role
      instance_type: t3.medium
  networking:
    plugin: calico
`

	err := Validate(invalidYAML)
	if err == nil {
		t.Error("expected error for invalid node role, got none")
	}
}

func TestValidate_NodeMissingInstanceType(t *testing.T) {
	invalidYAML := `
cluster:
  name: test-cluster
  nodes:
    - role: worker
  networking:
    plugin: calico
`

	err := Validate(invalidYAML)
	if err == nil {
		t.Error("expected error for node missing instance_type, got none")
	}
}

func TestValidate_MissingNetworkingPlugin(t *testing.T) {
	invalidYAML := `
cluster:
  name: test-cluster
  nodes:
    - role: control-plane
      instance_type: t3.medium
  networking: {}
`

	err := Validate(invalidYAML)
	if err == nil {
		t.Error("expected error for missing networking.plugin, got none")
	}
}

func TestValidate_InvalidNetworkingPlugin(t *testing.T) {
	invalidYAML := `
cluster:
  name: test-cluster
  nodes:
    - role: control-plane
      instance_type: t3.medium
  networking:
    plugin: invalid-plugin
`

	err := Validate(invalidYAML)
	if err == nil {
		t.Error("expected error for invalid networking.plugin, got none")
	}
}

func TestValidate_ApplicationMissingName(t *testing.T) {
	invalidYAML := `
cluster:
  name: test-cluster
  nodes:
    - role: control-plane
      instance_type: t3.medium
  networking:
    plugin: calico
  applications:
    - version: 1.0.0
`

	err := Validate(invalidYAML)
	if err == nil {
		t.Error("expected error for application missing name, got none")
	}
}

func TestValidate_MalformedYAML(t *testing.T) {
	malformedYAML := `
cluster:
  name: test
  nodes:
    - role: control-plane
      instance_type: t3.medium
    invalid yaml syntax here
`

	err := Validate(malformedYAML)
	if err == nil {
		t.Error("expected error for malformed YAML, got none")
	}
}
