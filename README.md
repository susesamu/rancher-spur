# Spur

[![Go Version](https://img.shields.io/badge/go-1.23-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)

Spur is a Go-based CLI tool that automates environment reproduction from Jira issues. It streamlines the process of recreating environments for debugging, testing, and incident reproduction by integrating:

- **Jira API** → Fetch issue details and environment information
- **Claude API** → Generate infrastructure-as-code YAML specifications
- **Saddle CLI** → Provision the environment from generated YAML

## Features

- Automated environment setup from Jira issue descriptions
- AI-powered YAML generation using Claude Sonnet 4.6
- Automatic retry logic for invalid YAML with error feedback
- Dry-run mode for YAML preview without provisioning
- Verbose logging for debugging
- Configuration via environment variables or config file

## Prerequisites

- Go 1.23 or later
- Jira account with API access
- Claude API key (from Anthropic)
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

Spur requires configuration for Jira and Claude API access. You can configure it using environment variables or a config file.

### Environment Variables (Recommended)

```bash
export SPUR_JIRA_URL="https://your-company.atlassian.net"
export SPUR_JIRA_USER="your-email@example.com"
export SPUR_JIRA_TOKEN="your-jira-api-token"
export SPUR_CLAUDE_API_KEY="your-claude-api-key"
export SPUR_CLAUDE_MODEL="claude-sonnet-4-6"  # Optional, defaults to claude-sonnet-4-6
```

### Config File (Optional)

Create a config file at `~/.spur/config.yaml`:

```yaml
jira:
  url: https://your-company.atlassian.net
  username: your-email@example.com
  token: your-jira-api-token

claude:
  api_key: your-claude-api-key
  model: claude-sonnet-4-6  # Optional
```

**Note:** Environment variables take precedence over config file values.

### Getting API Credentials

**Jira API Token:**
1. Go to https://id.atlassian.com/manage-profile/security/api-tokens
2. Click "Create API token"
3. Copy the generated token

**Claude API Key:**
1. Sign up at https://console.anthropic.com
2. Navigate to API Keys section
3. Create a new API key

## Usage

### Basic Usage

```bash
spur reproduce JIRA-123
```

This will:
1. Fetch issue `JIRA-123` from Jira
2. Generate a Saddle YAML configuration using Claude
3. Save the YAML to `JIRA-123.yaml`
4. Execute `saddle create JIRA-123.yaml` to provision the environment

### Dry Run Mode

Generate YAML without provisioning:

```bash
spur reproduce JIRA-123 --dry-run
```

### Custom Output File

```bash
spur reproduce JIRA-123 --output my-environment.yaml
```

### Verbose Logging

```bash
spur reproduce JIRA-123 --verbose
```

This shows detailed logs including:
- Jira issue details (summary, environment, labels, components)
- YAML generation time
- Saddle execution output

### Combined Flags

```bash
spur reproduce JIRA-123 --output staging-env.yaml --dry-run --verbose
```

## YAML Schema

Spur generates YAML configurations following the Saddle schema:

```yaml
cluster:
  name: repro-JIRA-123
  nodes:
    - role: control-plane
      instance_type: t3.medium
      count: 1
    - role: worker
      instance_type: t3.large
      count: 3
  networking:
    plugin: calico  # Options: calico, flannel, cilium
  applications:
    - name: nginx
      version: 1.21.0
      config:
        replicas: 3
```

### Schema Fields

- **cluster.name**: Unique cluster identifier
- **cluster.nodes**: Array of node configurations
  - **role**: `control-plane` or `worker`
  - **instance_type**: Instance type (e.g., `t3.medium`)
  - **count**: Number of nodes (optional, default: 1)
- **cluster.networking.plugin**: Network plugin (`calico`, `flannel`, or `cilium`)
- **cluster.applications**: Array of applications to deploy (optional)
  - **name**: Application name
  - **version**: Application version (optional)
  - **config**: Application-specific configuration (optional)

## Error Handling

### Common Errors

**Jira authentication failure:**
```
Error: failed to fetch Jira issue: authentication failed: check your Jira credentials
```
→ Verify `SPUR_JIRA_USER` and `SPUR_JIRA_TOKEN`

**Issue not found:**
```
Error: failed to fetch Jira issue: issue JIRA-999 not found
```
→ Check that the issue ID is correct and you have access to it

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
│   ├── jira/              # Jira API client
│   ├── claude/            # Claude API integration
│   ├── saddle/            # Saddle CLI executor
│   ├── config/            # Configuration management
│   └── yaml/              # YAML validation
└── main.go                # Entry point
```

### Key Packages

- **cmd**: Cobra-based CLI command structure
- **internal/jira**: Jira REST API v3 client with Basic Auth
- **internal/claude**: Claude API integration with retry logic
- **internal/saddle**: Saddle CLI execution wrapper
- **internal/config**: Viper-based configuration (env vars + file)
- **internal/yaml**: YAML validation against Saddle schema

## How It Works

1. **Fetch Issue**: Connects to Jira API using Basic Auth and retrieves issue details including summary, description, environment field, labels, and components.

2. **Generate YAML**: Sends issue data to Claude API with a structured prompt that instructs Claude to generate valid Saddle YAML. If validation fails, automatically retries with error feedback (max 2 retries).

3. **Validate YAML**: Checks the generated YAML against the Saddle schema, ensuring all required fields are present and valid.

4. **Provision**: Executes `saddle create <yaml-file>` to provision the environment, streaming output to the user in real-time.

## Troubleshooting

### Issue: Claude generates invalid YAML

**Solution:** 
- Add more detailed environment information to your Jira issue
- Include specific instance types, networking requirements, and application versions
- Use the `--verbose` flag to see the validation errors

### Issue: Saddle execution fails

**Solution:**
- Verify Saddle is properly configured
- Check that you have necessary cloud provider credentials
- Run `saddle create <file>` manually to see detailed error messages

### Issue: Rate limiting from APIs

**Solution:**
- Wait a few minutes before retrying
- For Jira: Check your API rate limits in Atlassian settings
- For Claude: Check your API tier limits

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

## Roadmap

- [ ] Support for custom Jira field mapping
- [ ] YAML template customization
- [ ] Caching of Claude responses
- [ ] Multi-cloud provider support
- [ ] Interactive mode for manual corrections
- [ ] CI/CD pipeline integration examples
