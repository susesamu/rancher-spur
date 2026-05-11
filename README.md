# Spur

[![Go Version](https://img.shields.io/badge/go-1.23-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)

Spur is a Go-based CLI tool that automates environment reproduction from Jira issues. It streamlines the process of recreating environments for debugging, testing, and incident reproduction by integrating:

- **Jira API** → Fetch issue details and environment information
- **Claude API (via gcloud)** → Generate infrastructure-as-code YAML specifications
- **Saddle CLI** → Provision AWS/RKE2/Rancher environments from generated YAML

## Features

- Automated environment setup from Jira issue descriptions
- AI-powered YAML generation using Claude Sonnet 4.6 via Google Cloud CLI
- Automatic retry logic for invalid YAML with error feedback
- Support for AWS provider with RKE2 and Rancher
- Dry-run mode for YAML preview without provisioning
- Verbose logging for debugging
- Configuration via environment variables or config file

## Prerequisites

- Go 1.23 or later
- Jira account with API Bearer token
- Google Cloud CLI (`gcloud`) installed and authenticated
- Saddle CLI (for provisioning environments)

## Installation

### From Source

```bash
git clone https://github.com/suse/rancher/rancher-spur.git
cd rancher-spur
make build
```

This creates a `spur` binary in the current directory.

### Install to GOPATH

```bash
make install
```

This installs `spur` to `$GOPATH/bin`, making it available system-wide.

## Configuration

Spur requires configuration for Jira API access. Google Cloud authentication is handled through `gcloud auth login`.

### Required Environment Variables

```bash
export SPUR_JIRA_URL="https://jira.suse.com"
export SPUR_JIRA_BEARER_TOKEN="your-jira-bearer-token"
```

### Optional Configuration

**Claude Model (defaults to claude-sonnet-4-6):**
```bash
export SPUR_CLAUDE_MODEL="claude-sonnet-4-6"
```

**AWS Credentials (optional - uses placeholders if not set):**
```bash
export SPUR_AWS_ACCESS_KEY="AKIA..."
export SPUR_AWS_SECRET_KEY="..."
export SPUR_AWS_REGION="us-west-2"  # Default: us-west-2
export SPUR_AWS_INSTANCE_TYPE="t3.xlarge"  # Default: t3.xlarge
export SPUR_AWS_SECURITY_GROUP_ID="sg-..."
export SPUR_AWS_SUBNET_ID="subnet-..."
export SPUR_AWS_AMI="ami-..."
```

**SSH Configuration (optional):**
```bash
export SPUR_SSH_KEY_NAME="your-aws-key"
export SPUR_SSH_PRIVATE_KEY_PATH="~/.ssh/your-key.pem"
export SPUR_SSH_USER="ubuntu"  # Default: ubuntu
```

### Config File (Optional)

Create a config file at `~/.spur/config.yaml`:

```yaml
jira:
  url: https://jira.suse.com
  bearer_token: your-bearer-token

claude:
  model: claude-sonnet-4-6

aws:
  access_key: AKIA...
  secret_key: ...
  region: us-west-2
  instance_type: t3.xlarge
  security_group_id: sg-...
  subnet_id: subnet-...
  ami: ami-...

ssh:
  key_name: your-aws-key
  private_key_path: ~/.ssh/your-key.pem
  user: ubuntu
```

**Note:** Environment variables take precedence over config file values.

### Getting Credentials

**Jira Bearer Token:**
1. Go to your Jira instance
2. Generate a Personal Access Token or API token
3. Use it as the Bearer token

**Google Cloud Authentication:**
```bash
gcloud auth login
```

This authenticates you for Claude API access via Google Cloud's Vertex AI.

## Usage

### Basic Usage

```bash
spur reproduce SURE-11610
```

This will:
1. Fetch issue `SURE-11610` from Jira
2. Generate a Saddle YAML configuration using Claude via gcloud
3. Save the YAML to `SURE-11610.yaml`
4. Execute `saddle create SURE-11610.yaml` to provision the environment

### Dry Run Mode

Generate YAML without provisioning:

```bash
spur reproduce SURE-11610 --dry-run
```

### Custom Output File

```bash
spur reproduce SURE-11610 --output production-env.yaml
```

### Verbose Logging

```bash
spur reproduce SURE-11610 --verbose
```

This shows detailed logs including:
- Jira issue details (summary, description, environment)
- YAML generation time
- Saddle execution output

### Combined Flags

```bash
spur reproduce SURE-11610 --output staging-env.yaml --dry-run --verbose
```

## YAML Schema

Spur generates YAML configurations following the Saddle schema for AWS/RKE2/Rancher environments:

```yaml
clusters:
  repro-sure-11610:  # Cluster name: repro-<lowercase-issue-id>
    provider:
      type: aws
      config:
        access_key: PLACEHOLDER_ACCESS_KEY  # Or from SPUR_AWS_ACCESS_KEY
        secret_key: PLACEHOLDER_SECRET_KEY  # Or from SPUR_AWS_SECRET_KEY
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
        rancher_version: 2.13.5  # Extracted from Jira environment field
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
      private_key_path: ~/.ssh/suse-aws-key.pem
      user: ubuntu
    cluster:
      node_prefix: sure-11610
      instance_count: 3
```

### Schema Fields

**clusters:**
- Top-level map with cluster names as keys
- Cluster name format: `repro-<lowercase-issue-id>`

**provider:**
- `type`: Cloud provider (`aws`, `azure`, `gcp`)
- `config`: Provider-specific configuration (credentials, region, instance type, etc.)

**kubernetes:**
- `distribution`: Kubernetes distribution (`rke2`, `k3s`, `eks`, `aks`, `gke`)
- `config`: Distribution-specific configuration including version and Rancher deployment options

**rancher:** (optional)
- Rancher-specific deployment configuration
- `version`: Rancher version to deploy
- `deploy`: Whether to deploy Rancher
- `bootstrap_password`: Initial admin password

**ssh:**
- `key_name`: AWS SSH key pair name
- `private_key_path`: Path to private SSH key file
- `user`: SSH user for instance access

**cluster:**
- `node_prefix`: Prefix for node names
- `instance_count`: Number of instances to create

## Error Handling

### Common Errors

**Jira authentication failure:**
```
Error: failed to fetch Jira issue: authentication failed: check your Jira Bearer token
```
→ Verify `SPUR_JIRA_BEARER_TOKEN` is correct

**Issue not found:**
```
Error: failed to fetch Jira issue: issue SURE-999 not found
```
→ Check that the issue ID is correct and you have access to it

**gcloud not authenticated:**
```
Error: Claude API call failed: gcloud CLI not found in PATH
```
→ Install gcloud CLI and run `gcloud auth login`

**Invalid YAML generation:**
```
Error: failed to generate YAML: generated YAML is invalid after 2 attempts
```
→ The issue description may lack environment details. Try adding more context to the Jira issue.

**Saddle not installed:**
```
Error: saddle execution failed: saddle CLI not found in PATH
```
→ Install Saddle CLI and ensure it's in your PATH

## Development

### Running Tests

```bash
make test
```

This runs all unit tests with race detection and coverage reporting.

### Linting

```bash
make lint
```

Requires [golangci-lint](https://golangci-lint.run/usage/install/) to be installed.

### Cleaning Build Artifacts

```bash
make clean
```

Removes the `spur` binary, generated YAML files, and coverage reports.

## Architecture

```
spur/
├── cmd/                    # CLI commands
│   ├── root.go            # Root command setup
│   └── reproduce.go       # Reproduce command implementation
├── internal/
│   ├── jira/              # Jira API v2 client with Bearer token auth
│   ├── claude/            # Claude API via gcloud CLI
│   ├── saddle/            # Saddle CLI executor
│   ├── config/            # Configuration management
│   └── yaml/              # YAML validation for Saddle schema
└── main.go                # Entry point
```

### Key Packages

- **cmd**: Cobra-based CLI command structure
- **internal/jira**: Jira REST API v2 client with Bearer token authentication
- **internal/claude**: Claude API integration via gcloud CLI with retry logic
- **internal/saddle**: Saddle CLI execution wrapper
- **internal/config**: Viper-based configuration (env vars + file)
- **internal/yaml**: YAML validation against Saddle schema (AWS/RKE2/Rancher)

## How It Works

1. **Fetch Issue**: Connects to Jira API v2 using Bearer token authentication and retrieves issue details including summary, description, and environment field.

2. **Generate YAML**: Sends issue data to Claude API via gcloud CLI with a structured prompt that instructs Claude to generate valid Saddle YAML. Attempts to extract Rancher version from the environment field. If validation fails, automatically retries with error feedback (max 2 retries).

3. **Validate YAML**: Checks the generated YAML against the Saddle schema, ensuring all required fields are present and valid for AWS/RKE2/Rancher provisioning.

4. **Provision**: Executes `saddle create <yaml-file>` to provision the environment, streaming output to the user in real-time.

## Troubleshooting

### Issue: Claude generates invalid YAML

**Solution:** 
- Add more detailed environment information to your Jira issue
- Include specific Rancher version, instance requirements, and application details
- Use the `--verbose` flag to see the validation errors

### Issue: Saddle execution fails

**Solution:**
- Verify Saddle is properly configured
- Check that you have AWS credentials configured
- Run `saddle create <file>` manually to see detailed error messages

### Issue: gcloud authentication errors

**Solution:**
- Run `gcloud auth login` to authenticate
- Verify you have access to Vertex AI in your GCP project
- Check that the Claude model name is correct

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/my-feature`
3. Make your changes and add tests
4. Run tests: `make test`
5. Run linter: `make lint`
6. Commit your changes: `git commit -am 'Add feature'`
7. Push to the branch: `git push origin feature/my-feature`
8. Open a Pull Request

## License

Apache 2.0

## Support

For issues and questions:
- GitHub Issues: https://github.com/suse/rancher/rancher-spur/issues
- Documentation: See this README

## Changelog

### v2.0.0 - Breaking Changes
- **Jira API**: Migrated from API v3 to v2 with Bearer token authentication
- **Claude Integration**: Changed from Anthropic SDK to gcloud CLI for Vertex AI
- **YAML Schema**: Complete rewrite to support real Saddle schema for AWS/RKE2/Rancher
- **Configuration**: Removed `SPUR_JIRA_USER` and `SPUR_CLAUDE_API_KEY`, added AWS/SSH config options

### Migration Guide from v1.x

**Environment Variables:**
- Change `SPUR_JIRA_TOKEN` → `SPUR_JIRA_BEARER_TOKEN`
- Remove `SPUR_JIRA_USER` (no longer needed)
- Remove `SPUR_CLAUDE_API_KEY` (use `gcloud auth login` instead)
- Add AWS/SSH config if you want specific values instead of placeholders

**YAML Schema:**
- Old YAML files from v1.x are not compatible with v2.0
- New schema focuses on AWS/RKE2/Rancher environments
- See YAML Schema section above for new format

## Roadmap

- [ ] Support for Azure and GCP providers
- [ ] Custom Jira field mapping configuration
- [ ] YAML template customization
- [ ] Caching of Claude responses
- [ ] K3s support alongside RKE2
- [ ] Interactive mode for manual YAML corrections
- [ ] CI/CD pipeline integration examples
