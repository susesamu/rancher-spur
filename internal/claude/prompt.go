package claude

import (
	"fmt"
	"strings"

	"github.com/suse/rancher/rancher-spur/internal/config"
)

const systemPrompt = `You are an infrastructure automation expert. Your role is to generate valid YAML configurations for the Saddle provisioning tool.

Output Requirements:
- ONLY output valid YAML, no explanations or markdown
- Follow this exact schema:

clusters:
  <cluster-name>:
    provider:
      type: aws
      config:
        access_key: <placeholder-or-value>
        secret_key: <placeholder-or-value>
        region: us-west-2
        ami: <aws-ami-id>
        instance_type: t3.xlarge
        security_group_id: <sg-id>
        subnet_id: <subnet-id>
    kubernetes:
      distribution: rke2
      config:
        version: <k8s-version>
        deploy_rancher: true
        rancher_version: <rancher-version>
        rancher_bootstrap_password: admin
        rancher_prime: false
        rancher_debug: false
    rancher:
      version: <rancher-version>
      deploy: true
      prime: false
      bootstrap_password: admin
    ssh:
      key_name: <ssh-key-name>
      private_key_path: <path-to-private-key>
      user: ubuntu
    cluster:
      node_prefix: <prefix>
      instance_count: 3

Rules:
- Set cluster name as "repro-<lowercase-issue-id>" (e.g., for SURE-11610 use "repro-sure-11610")
- Extract Rancher version from the "environment" field if present (look for patterns like "Rancher version: 2.13.5" or "2.14.1")
- Use placeholder credentials: PLACEHOLDER_ACCESS_KEY, PLACEHOLDER_SECRET_KEY
- Generate reasonable dummy values for security_group_id (format: sg-XXXXXXXXXX) and subnet_id (format: subnet-XXXXXXXXXX)
- For AMI, use a common Ubuntu AMI for us-west-2: ami-0a3e3ef8596692376
- Default instance_type: t3.xlarge
- Default region: us-west-2
- Default kubernetes.distribution: rke2
- Default ssh.user: ubuntu
- Default cluster.instance_count: 3
- If Rancher version is found in environment, use it; otherwise use 2.12.0
- For kubernetes.config.version, use a reasonable RKE2 version like v1.33.7+rke2r1 or infer from context
- Generate a meaningful node_prefix based on the issue (e.g., "sure-11610" for issue SURE-11610)`

// buildUserPrompt constructs the user prompt from issue data
func buildUserPrompt(input *PromptInput, awsConfig *config.AWSConfig, sshConfig *config.SSHConfig) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Generate Saddle YAML for Jira issue %s:\n\n", input.IssueID))
	sb.WriteString(fmt.Sprintf("Summary: %s\n\n", input.Summary))

	if input.Description != "" {
		sb.WriteString(fmt.Sprintf("Description:\n%s\n\n", input.Description))
	}

	if input.Environment != "" {
		sb.WriteString(fmt.Sprintf("Environment: %s\n\n", input.Environment))
	}

	// Include AWS config if provided
	if awsConfig != nil && awsConfig.AccessKey != "" {
		sb.WriteString("\nAWS Configuration (use these values instead of placeholders):\n")
		if awsConfig.AccessKey != "" {
			sb.WriteString(fmt.Sprintf("- Access Key: %s\n", awsConfig.AccessKey))
		}
		if awsConfig.SecretKey != "" {
			sb.WriteString(fmt.Sprintf("- Secret Key: %s\n", awsConfig.SecretKey))
		}
		if awsConfig.Region != "" {
			sb.WriteString(fmt.Sprintf("- Region: %s\n", awsConfig.Region))
		}
		if awsConfig.InstanceType != "" {
			sb.WriteString(fmt.Sprintf("- Instance Type: %s\n", awsConfig.InstanceType))
		}
		if awsConfig.SecurityGroupID != "" {
			sb.WriteString(fmt.Sprintf("- Security Group ID: %s\n", awsConfig.SecurityGroupID))
		}
		if awsConfig.SubnetID != "" {
			sb.WriteString(fmt.Sprintf("- Subnet ID: %s\n", awsConfig.SubnetID))
		}
		if awsConfig.AMI != "" {
			sb.WriteString(fmt.Sprintf("- AMI: %s\n", awsConfig.AMI))
		}
		sb.WriteString("\n")
	}

	// Include SSH config if provided
	if sshConfig != nil && sshConfig.KeyName != "" {
		sb.WriteString("\nSSH Configuration (use these values):\n")
		if sshConfig.KeyName != "" {
			sb.WriteString(fmt.Sprintf("- Key Name: %s\n", sshConfig.KeyName))
		}
		if sshConfig.PrivateKeyPath != "" {
			sb.WriteString(fmt.Sprintf("- Private Key Path: %s\n", sshConfig.PrivateKeyPath))
		}
		if sshConfig.User != "" {
			sb.WriteString(fmt.Sprintf("- User: %s\n", sshConfig.User))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// buildRetryPrompt constructs a retry prompt with error feedback
func buildRetryPrompt(originalPrompt, validationError string) string {
	return fmt.Sprintf(`The previous YAML output was invalid:
%s

Please regenerate valid YAML following the schema exactly.

Original request:
%s`, validationError, originalPrompt)
}
