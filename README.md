# Pulumi GCP Go: Virtual Machine

This project provisions a GCP Compute Engine VM with VPC, subnet, and firewall rules using Pulumi and Go. It demonstrates how to:
- Use the Pulumi GCP provider in a Go program
- Create a VPC network and subnet
- Configure firewall rules for SSH access
- Deploy a Free Tier eligible VM instance

## Providers

- Google Cloud Platform via the Pulumi GCP SDK for Go (`github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp`)

## Resources

- **VPC Network** (`gcp.compute.Network`)
  - Custom VPC with auto subnet creation disabled

- **Subnetwork** (`gcp.compute.Subnetwork`)
  - Regional subnet with configurable CIDR range

- **Firewall Rules** (`gcp.compute.Firewall`)
  - Configurable ingress/egress rules
  - Default allows SSH (port 22)

- **Compute Instance** (`gcp.compute.Instance`)
  - Free Tier eligible e2-micro instance
  - Debian 12 boot disk
  - SSH key metadata

## Outputs

- **vmExternalIp**: External IP address of the VM
- **vmInternalIp**: Internal IP address of the VM
- **sshConnectionCommand**: SSH command to connect to the VM

## When to Use This Project

- You need a simple VM for development or testing
- You want to stay within GCP Free Tier limits
- You need a repeatable VM deployment with networking

## Prerequisites

- Go 1.21+ installed
- A Google Cloud account with billing enabled
- GCP credentials configured for Pulumi (via `gcloud auth application-default login`)
- Compute Engine API enabled

## Usage

1. Install dependencies:
   ```bash
   go mod tidy
   ```

2. Configure your stack:
   ```bash
   cp Pulumi.dev.yaml.example Pulumi.dev.yaml
   pulumi config set gcp:project YOUR_PROJECT_ID
   pulumi config set --secret vm:sshPublicKey "$(cat ~/.ssh/id_rsa.pub)"
   ```

3. Preview and deploy:
   ```bash
   pulumi preview
   pulumi up
   ```

4. Connect to the VM:
   ```bash
   pulumi stack output sshConnectionCommand
   # Or manually:
   ssh gcp-user@<external-ip>
   ```

## Project Layout

```
├── Pulumi.yaml                  Pulumi project definition
├── Pulumi.dev.yaml.example      Template for local dev configuration
├── go.mod                       Go module declaration and dependencies
├── main.go                      Pulumi program defining VM resources
├── .gitignore                   Git ignore rules
└── LICENSE                      MIT License
```

## Configuration

| Name | Description | Default |
|------|-------------|---------|
| `gcp:project` | The Google Cloud project to deploy into | _required_ |
| `gcp:region` | The region to deploy to | `us-central1` |
| `gcp:zone` | The zone to deploy to | `us-central1-a` |
| `vm:instanceName` | Name of the VM instance | `free-tier-vm` |
| `vm:machineType` | Machine type | `e2-micro` |
| `vm:sshUser` | SSH username | `gcp-user` |
| `vm:sshPublicKey` | SSH public key (secret) | _required_ |
| `vm:diskSizeGb` | Boot disk size in GB | `30` |
| `vm:diskType` | Boot disk type | `pd-standard` |
| `vm:firewallRules` | JSON list of firewall rules | See Pulumi.dev.yaml.example |

## Next Steps

- Add startup scripts for automated provisioning
- Configure Cloud NAT for private instances
- Add managed instance groups for scaling
- Set up Cloud Monitoring agent
- Configure OS Login for SSH management

## Getting Help

- Pulumi Documentation: https://www.pulumi.com/docs/
- GCP Provider Reference: https://www.pulumi.com/registry/packages/gcp/
- Community Slack: https://slack.pulumi.com/
- GitHub Issues: https://github.com/pulumi/pulumi/issues