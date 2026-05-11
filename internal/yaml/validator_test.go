package yaml

import (
	"testing"
)

func TestValidate_ValidYAML(t *testing.T) {
	validYAML := `
clusters:
  repro-sure-11610:
    provider:
      type: aws
      config:
        access_key: AKIATEST123
        secret_key: secretkey123
        region: us-west-2
        ami: ami-0a3e3ef8596692376
        instance_type: t3.xlarge
        security_group_id: sg-0c1663c340fac1acd
        subnet_id: subnet-066d7c2f2bea54812
    kubernetes:
      distribution: rke2
      config:
        version: v1.33.7+rke2r1
        deploy_rancher: true
        rancher_version: 2.13.5
        rancher_bootstrap_password: admin
        rancher_prime: false
        rancher_debug: false
    rancher:
      version: 2.13.5
      deploy: true
      prime: false
      bootstrap_password: admin
    ssh:
      key_name: suse-aws-key
      private_key_path: ~/SUSE/keys/suse-aws-key
      user: ubuntu
    cluster:
      node_prefix: sure-11610
      instance_count: 3
`

	err := Validate(validYAML)
	if err != nil {
		t.Errorf("expected no error for valid YAML, got: %v", err)
	}
}

func TestValidate_MinimalValidYAML(t *testing.T) {
	minimalYAML := `
clusters:
  test-cluster:
    provider:
      type: aws
      config:
        region: us-east-1
    kubernetes:
      distribution: rke2
      config:
        version: v1.30.0+rke2r1
    ssh:
      key_name: my-key
      private_key_path: /path/to/key
      user: ubuntu
    cluster:
      node_prefix: test
      instance_count: 1
`

	err := Validate(minimalYAML)
	if err != nil {
		t.Errorf("expected no error for minimal valid YAML, got: %v", err)
	}
}

func TestValidate_WithMarkdownCodeFence(t *testing.T) {
	yamlWithFence := "```yaml\nclusters:\n  test:\n    provider:\n      type: aws\n      config:\n        region: us-west-2\n    kubernetes:\n      distribution: rke2\n      config:\n        version: v1.30.0+rke2r1\n    ssh:\n      key_name: test\n      private_key_path: /test\n      user: ubuntu\n    cluster:\n      node_prefix: test\n      instance_count: 1\n```"

	err := Validate(yamlWithFence)
	if err != nil {
		t.Errorf("expected no error when stripping markdown fences, got: %v", err)
	}
}

func TestValidate_NoClusters(t *testing.T) {
	invalidYAML := `
clusters: {}
`

	err := Validate(invalidYAML)
	if err == nil {
		t.Error("expected error for empty clusters map, got none")
	}
}

func TestValidate_MissingProviderType(t *testing.T) {
	invalidYAML := `
clusters:
  test-cluster:
    provider:
      config:
        region: us-west-2
    kubernetes:
      distribution: rke2
      config:
        version: v1.30.0+rke2r1
    ssh:
      key_name: test
      private_key_path: /test
      user: ubuntu
    cluster:
      node_prefix: test
      instance_count: 1
`

	err := Validate(invalidYAML)
	if err == nil {
		t.Error("expected error for missing provider.type, got none")
	}
}

func TestValidate_InvalidProviderType(t *testing.T) {
	invalidYAML := `
clusters:
  test-cluster:
    provider:
      type: invalid-provider
      config:
        region: us-west-2
    kubernetes:
      distribution: rke2
      config:
        version: v1.30.0+rke2r1
    ssh:
      key_name: test
      private_key_path: /test
      user: ubuntu
    cluster:
      node_prefix: test
      instance_count: 1
`

	err := Validate(invalidYAML)
	if err == nil {
		t.Error("expected error for invalid provider type, got none")
	}
}

func TestValidate_EmptyProviderConfig(t *testing.T) {
	invalidYAML := `
clusters:
  test-cluster:
    provider:
      type: aws
      config: {}
    kubernetes:
      distribution: rke2
      config:
        version: v1.30.0+rke2r1
    ssh:
      key_name: test
      private_key_path: /test
      user: ubuntu
    cluster:
      node_prefix: test
      instance_count: 1
`

	err := Validate(invalidYAML)
	if err == nil {
		t.Error("expected error for empty provider config, got none")
	}
}

func TestValidate_MissingKubernetesDistribution(t *testing.T) {
	invalidYAML := `
clusters:
  test-cluster:
    provider:
      type: aws
      config:
        region: us-west-2
    kubernetes:
      config:
        version: v1.30.0+rke2r1
    ssh:
      key_name: test
      private_key_path: /test
      user: ubuntu
    cluster:
      node_prefix: test
      instance_count: 1
`

	err := Validate(invalidYAML)
	if err == nil {
		t.Error("expected error for missing kubernetes.distribution, got none")
	}
}

func TestValidate_InvalidKubernetesDistribution(t *testing.T) {
	invalidYAML := `
clusters:
  test-cluster:
    provider:
      type: aws
      config:
        region: us-west-2
    kubernetes:
      distribution: invalid-distro
      config:
        version: v1.30.0+rke2r1
    ssh:
      key_name: test
      private_key_path: /test
      user: ubuntu
    cluster:
      node_prefix: test
      instance_count: 1
`

	err := Validate(invalidYAML)
	if err == nil {
		t.Error("expected error for invalid kubernetes distribution, got none")
	}
}

func TestValidate_EmptyKubernetesConfig(t *testing.T) {
	invalidYAML := `
clusters:
  test-cluster:
    provider:
      type: aws
      config:
        region: us-west-2
    kubernetes:
      distribution: rke2
      config: {}
    ssh:
      key_name: test
      private_key_path: /test
      user: ubuntu
    cluster:
      node_prefix: test
      instance_count: 1
`

	err := Validate(invalidYAML)
	if err == nil {
		t.Error("expected error for empty kubernetes config, got none")
	}
}

func TestValidate_MissingSSHKeyName(t *testing.T) {
	invalidYAML := `
clusters:
  test-cluster:
    provider:
      type: aws
      config:
        region: us-west-2
    kubernetes:
      distribution: rke2
      config:
        version: v1.30.0+rke2r1
    ssh:
      private_key_path: /test
      user: ubuntu
    cluster:
      node_prefix: test
      instance_count: 1
`

	err := Validate(invalidYAML)
	if err == nil {
		t.Error("expected error for missing ssh.key_name, got none")
	}
}

func TestValidate_MissingSSHPrivateKeyPath(t *testing.T) {
	invalidYAML := `
clusters:
  test-cluster:
    provider:
      type: aws
      config:
        region: us-west-2
    kubernetes:
      distribution: rke2
      config:
        version: v1.30.0+rke2r1
    ssh:
      key_name: test
      user: ubuntu
    cluster:
      node_prefix: test
      instance_count: 1
`

	err := Validate(invalidYAML)
	if err == nil {
		t.Error("expected error for missing ssh.private_key_path, got none")
	}
}

func TestValidate_MissingSSHUser(t *testing.T) {
	invalidYAML := `
clusters:
  test-cluster:
    provider:
      type: aws
      config:
        region: us-west-2
    kubernetes:
      distribution: rke2
      config:
        version: v1.30.0+rke2r1
    ssh:
      key_name: test
      private_key_path: /test
    cluster:
      node_prefix: test
      instance_count: 1
`

	err := Validate(invalidYAML)
	if err == nil {
		t.Error("expected error for missing ssh.user, got none")
	}
}

func TestValidate_MissingNodePrefix(t *testing.T) {
	invalidYAML := `
clusters:
  test-cluster:
    provider:
      type: aws
      config:
        region: us-west-2
    kubernetes:
      distribution: rke2
      config:
        version: v1.30.0+rke2r1
    ssh:
      key_name: test
      private_key_path: /test
      user: ubuntu
    cluster:
      instance_count: 1
`

	err := Validate(invalidYAML)
	if err == nil {
		t.Error("expected error for missing cluster.node_prefix, got none")
	}
}

func TestValidate_InvalidInstanceCount(t *testing.T) {
	invalidYAML := `
clusters:
  test-cluster:
    provider:
      type: aws
      config:
        region: us-west-2
    kubernetes:
      distribution: rke2
      config:
        version: v1.30.0+rke2r1
    ssh:
      key_name: test
      private_key_path: /test
      user: ubuntu
    cluster:
      node_prefix: test
      instance_count: 0
`

	err := Validate(invalidYAML)
	if err == nil {
		t.Error("expected error for instance_count <= 0, got none")
	}
}

func TestValidate_MalformedYAML(t *testing.T) {
	malformedYAML := `
clusters:
  test-cluster:
    provider:
      type: aws
      invalid yaml syntax here
`

	err := Validate(malformedYAML)
	if err == nil {
		t.Error("expected error for malformed YAML, got none")
	}
}

func TestValidate_MultipleValidDistributions(t *testing.T) {
	distributions := []string{"rke2", "k3s", "eks", "aks", "gke"}

	for _, distro := range distributions {
		yamlContent := `
clusters:
  test-cluster:
    provider:
      type: aws
      config:
        region: us-west-2
    kubernetes:
      distribution: ` + distro + `
      config:
        version: v1.30.0
    ssh:
      key_name: test
      private_key_path: /test
      user: ubuntu
    cluster:
      node_prefix: test
      instance_count: 1
`

		err := Validate(yamlContent)
		if err != nil {
			t.Errorf("expected no error for distribution '%s', got: %v", distro, err)
		}
	}
}

func TestValidate_MultipleValidProviders(t *testing.T) {
	providers := []string{"aws", "azure", "gcp"}

	for _, provider := range providers {
		yamlContent := `
clusters:
  test-cluster:
    provider:
      type: ` + provider + `
      config:
        region: us-west-2
    kubernetes:
      distribution: rke2
      config:
        version: v1.30.0
    ssh:
      key_name: test
      private_key_path: /test
      user: ubuntu
    cluster:
      node_prefix: test
      instance_count: 1
`

		err := Validate(yamlContent)
		if err != nil {
			t.Errorf("expected no error for provider '%s', got: %v", provider, err)
		}
	}
}
