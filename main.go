package main

import (
	"encoding/json"

	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

type FirewallRule struct {
	Name         string            `json:"name"`
	Direction    string            `json:"direction"`
	Priority     int               `json:"priority"`
	SourceRanges []string          `json:"sourceRanges"`
	Allowed      []FirewallAllowed `json:"allowed"`
}

type FirewallAllowed struct {
	Protocol string   `json:"protocol"`
	Ports    []string `json:"ports"`
}

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {

		projectID := config.Require(ctx, "gcp:project")
		region := config.Get(ctx, "gcp:region")
		if region == "" {
			region = "us-central1"
		}
		zone := config.Get(ctx, "gcp:zone")
		if zone == "" {
			zone = "us-central1-a"
		}

		instanceName := config.Get(ctx, "vm:instanceName")
		if instanceName == "" {
			instanceName = "free-tier-vm"
		}

		machineType := config.Get(ctx, "vm:machineType")
		if machineType == "" {
			machineType = "e2-micro"
		}

		sshUser := config.Get(ctx, "vm:sshUser")
		if sshUser == "" {
			sshUser = "gcp-user"
		}

		sshPublicKey := config.RequireSecret(ctx, "vm:sshPublicKey")

		diskSizeGb := config.GetInt(ctx, "vm:diskSizeGb")
		if diskSizeGb == 0 {
			diskSizeGb = 30
		}

		diskType := config.Get(ctx, "vm:diskType")
		if diskType == "" {
			diskType = "pd-standard"
		}

		// Parse firewall rules
		firewallRulesJSON := config.Get(ctx, "vm:firewallRules")
		var firewallRules []FirewallRule
		if firewallRulesJSON != "" {
			if err := json.Unmarshal([]byte(firewallRulesJSON), &firewallRules); err != nil {
				return err
			}
		} else {
			firewallRules = []FirewallRule{
				{
					Name:         "allow-ssh",
					Direction:    "INGRESS",
					Priority:     1000,
					SourceRanges: []string{"0.0.0.0/0"},
					Allowed:      []FirewallAllowed{{Protocol: "tcp", Ports: []string{"22"}}},
				},
			}
		}

		// Create VPC network
		vpcNetwork, err := compute.NewNetwork(ctx, "vpc-network", &compute.NetworkArgs{
			Name:                  pulumi.String("free-tier-vpc"),
			AutoCreateSubnetworks: pulumi.Bool(false),
		})
		if err != nil {
			return err
		}

		// Create subnet
		subnet, err := compute.NewSubnetwork(ctx, "subnet", &compute.SubnetworkArgs{
			Name:        pulumi.String("free-tier-subnet"),
			IpCidrRange: pulumi.String("10.0.1.0/24"),
			Region:      pulumi.String(region),
			Network:     vpcNetwork.ID(),
		})
		if err != nil {
			return err
		}

		// Create firewall rules
		var firewallResources []pulumi.Resource
		for _, rule := range firewallRules {
			var allows compute.FirewallAllowArray
			for _, allowed := range rule.Allowed {
				allows = append(allows, &compute.FirewallAllowArgs{
					Protocol: pulumi.String(allowed.Protocol),
					Ports:    pulumi.ToStringArray(allowed.Ports),
				})
			}

			fw, err := compute.NewFirewall(ctx, rule.Name, &compute.FirewallArgs{
				Name:         pulumi.String(rule.Name),
				Network:      vpcNetwork.Name,
				Direction:    pulumi.String(rule.Direction),
				Priority:     pulumi.Int(rule.Priority),
				SourceRanges: pulumi.ToStringArray(rule.SourceRanges),
				Allows:       allows,
				TargetTags:   pulumi.StringArray{pulumi.String("ssh-enabled")},
			})
			if err != nil {
				return err
			}
			firewallResources = append(firewallResources, fw)
		}

		// Create VM instance
		vmInstance, err := compute.NewInstance(ctx, "vm-instance", &compute.InstanceArgs{
			Name:        pulumi.String(instanceName),
			MachineType: pulumi.String(machineType),
			Zone:        pulumi.String(zone),
			Project:     pulumi.String(projectID),
			Tags:        pulumi.StringArray{pulumi.String("ssh-enabled")},
			BootDisk: &compute.InstanceBootDiskArgs{
				InitializeParams: &compute.InstanceBootDiskInitializeParamsArgs{
					Image: pulumi.String("debian-cloud/debian-12"),
					Size:  pulumi.Int(diskSizeGb),
					Type:  pulumi.String(diskType),
				},
			},
			NetworkInterfaces: compute.InstanceNetworkInterfaceArray{
				&compute.InstanceNetworkInterfaceArgs{
					Subnetwork: subnet.ID(),
					AccessConfigs: compute.InstanceNetworkInterfaceAccessConfigArray{
						&compute.InstanceNetworkInterfaceAccessConfigArgs{},
					},
				},
			},
			Metadata: pulumi.StringMap{
				"ssh-keys": pulumi.Sprintf("%s:%s", sshUser, sshPublicKey),
			},
		}, pulumi.DependsOn(firewallResources))
		if err != nil {
			return err
		}

		// Export outputs
		ctx.Export("vmExternalIp", vmInstance.NetworkInterfaces.Index(pulumi.Int(0)).AccessConfigs().Index(pulumi.Int(0)).NatIp())
		ctx.Export("vmInternalIp", vmInstance.NetworkInterfaces.Index(pulumi.Int(0)).NetworkIp())
		ctx.Export("sshConnectionCommand", pulumi.Sprintf("ssh -i ~/.ssh/id_rsa %s@%s", sshUser, vmInstance.NetworkInterfaces.Index(pulumi.Int(0)).AccessConfigs().Index(pulumi.Int(0)).NatIp()))

		return nil
	})
}
