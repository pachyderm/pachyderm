package main

import (
	"fmt"

	"github.com/pulumi/pulumi-awsx/sdk/go/awsx/ec2"
	"github.com/pulumi/pulumi-eks/sdk/go/eks"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

// Create a new EKS cluster; this leverages the pulumi aws crosswalk library.
// aws crosswalk (awsx) is a higher level library that makes it easier to
// create AWS resources since it created all the associated networking.
func DeployCluster(ctx *pulumi.Context) (*eks.Cluster, error) {
	cfg := config.New(ctx, "")
	minClusterSize, err := cfg.TryInt("minClusterSize")
	if err != nil {
		minClusterSize = 3
	}
	maxClusterSize, err := cfg.TryInt("maxClusterSize")
	if err != nil {
		maxClusterSize = 6
	}
	desiredClusterSize, err := cfg.TryInt("desiredClusterSize")
	if err != nil {
		desiredClusterSize = 3
	}
	eksNodeInstanceType, err := cfg.Try("eksNodeInstanceType")
	if err != nil {
		eksNodeInstanceType = "t2.medium"
	}
	vpcNetworkCidr, err := cfg.Try("vpcNetworkCidr")
	if err != nil {
		vpcNetworkCidr = "10.0.0.0/16"
	}

	// Create a new VPC, subnets, and associated infrastructure
	vpcName := fmt.Sprintf("eks-%s-vpc", ctx.Stack())
	eksVpc, err := ec2.NewVpc(ctx, vpcName, &ec2.VpcArgs{
		EnableDnsHostnames: pulumi.Bool(true),
		CidrBlock:          &vpcNetworkCidr,
		Tags: pulumi.StringMap{
			"Project": pulumi.String("Feature Testing"),
			"Service": pulumi.String("CI"),
			"Owner":   pulumi.String("pachyderm-ci"),
			"Team":    pulumi.String("Core"),
		},
	})
	if err != nil {
		return nil, err
	}

	// Create a new EKS cluster
	eksClusterName := fmt.Sprintf("eks-%s-cluster", ctx.Stack())
	eksCluster, err := eks.NewCluster(ctx, eksClusterName, &eks.ClusterArgs{
		// Put the cluster in the new VPC created earlier
		VpcId: eksVpc.VpcId,
		// Public subnets will be used for load balancers
		PublicSubnetIds: eksVpc.PublicSubnetIds,
		// Private subnets will be used for cluster nodes
		PrivateSubnetIds: eksVpc.PrivateSubnetIds,
		// Change configuration values above to change any of the following settings
		InstanceType:    pulumi.String(eksNodeInstanceType),
		DesiredCapacity: pulumi.Int(desiredClusterSize),
		MinSize:         pulumi.Int(minClusterSize),
		MaxSize:         pulumi.Int(maxClusterSize),
		// Do not give the worker nodes a public IP address
		NodeAssociatePublicIpAddress: pulumi.BoolRef(false),
		// Uncomment the next two lines for a private cluster (VPN access required)
		// EndpointPrivateAccess: 	      pulumi.Bool(true),
		// EndpointPublicAccess:         pulumi.Bool(false),
		Tags: pulumi.StringMap{
			"Project": pulumi.String("Feature Testing"),
			"Service": pulumi.String("CI"),
			"Owner":   pulumi.String("pachyderm-ci"),
			"Team":    pulumi.String("Core"),
		},
	})
	if err != nil {
		return nil, err
	}

	// Export some values in case they are needed elsewhere
	ctx.Export("kubeconfig", eksCluster.Kubeconfig)
	ctx.Export("vpcId", eksVpc.VpcId)

	return eksCluster, nil
}
