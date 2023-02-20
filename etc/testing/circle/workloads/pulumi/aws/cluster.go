package main

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	awseks "github.com/pulumi/pulumi-aws/sdk/v5/go/aws/eks"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/iam"
	"github.com/pulumi/pulumi-awsx/sdk/go/awsx/ec2"
	"github.com/pulumi/pulumi-eks/sdk/go/eks"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	storagev1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/storage/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

// Create a new EKS cluster; this leverages the pulumi aws crosswalk library.
// aws crosswalk (awsx) is a higher level library that makes it easier to
// create AWS resources since it created all the associated networking.
func DeployCluster(ctx *pulumi.Context) (*kubernetes.Provider, *iam.Role, error) {
	cfg := config.New(ctx, "")
	minClusterSize, err := cfg.TryInt("minClusterSize")
	if err != nil {
		minClusterSize = 1
	}
	maxClusterSize, err := cfg.TryInt("maxClusterSize")
	if err != nil {
		maxClusterSize = 3
	}
	desiredClusterSize, err := cfg.TryInt("desiredClusterSize")
	if err != nil {
		desiredClusterSize = 1
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
			"Project":     pulumi.String("Feature Testing"),
			"Service":     pulumi.String("CI"),
			"Owner":       pulumi.String("pachyderm-ci"),
			"Team":        pulumi.String("Core"),
			"Environment": pulumi.String(ctx.Stack()),
		},
	})
	if err != nil {
		return nil, nil, errors.WithStack(fmt.Errorf("error creating VPC: %w", err))
	}

	// Create a new EKS cluster
	eksClusterName := fmt.Sprintf("eks-%s-cluster", ctx.Stack())
	eksCluster, err := eks.NewCluster(ctx, eksClusterName, &eks.ClusterArgs{
		// Put the cluster in the new VPC created earlier
		VpcId:              eksVpc.VpcId,
		CreateOidcProvider: pulumi.Bool(true),
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
			"Project":     pulumi.String("Feature Testing"),
			"Service":     pulumi.String("CI"),
			"Owner":       pulumi.String("pachyderm-ci"),
			"Team":        pulumi.String("Core"),
			"Environment": pulumi.String(ctx.Stack()),
		},
	})
	if err != nil {
		return nil, nil, errors.WithStack(fmt.Errorf("error creating EKS cluster: %w", err))
	}

	k8sProvider, err := kubernetes.NewProvider(ctx, "k8sprovider", &kubernetes.ProviderArgs{
		Kubeconfig: eksCluster.KubeconfigJson,
	})

	if err != nil {
		return nil, nil, errors.WithStack(fmt.Errorf("error creating k8s provider: %w", err))
	}

	clusterOidcProviderUrl := eksCluster.Core.OidcProvider().Url()

	assumeRolePolicyDocument := pulumi.All(clusterOidcProviderUrl, eksCluster.Core.OidcProvider().Arn()).ApplyT(func(args []interface{}) (string, error) {
		url := args[0].(string)
		arn := args[1].(string)

		urlTrimmed := strings.TrimPrefix(url, "https://")
		urlAud := fmt.Sprintf("%s:aud", urlTrimmed)
		urlSub := fmt.Sprintf("%s:sub", urlTrimmed)

		assumeRolePolicy, _ := iam.GetPolicyDocument(ctx, &iam.GetPolicyDocumentArgs{
			Statements: []iam.GetPolicyDocumentStatement{
				{
					Sid: &[]string{"allowK8sServiceAccount"}[0],
					Actions: []string{
						"sts:AssumeRoleWithWebIdentity",
					},
					Principals: []iam.GetPolicyDocumentStatementPrincipal{
						{
							Identifiers: []string{
								arn,
							},
							Type: "Federated",
						},
					},
					Conditions: []iam.GetPolicyDocumentStatementCondition{
						{
							Test:     "StringEquals",
							Variable: urlAud,
							Values: []string{
								"sts.amazonaws.com",
							},
						},
						{
							Test:     "StringEquals",
							Variable: urlSub,
							Values: []string{
								"system:serviceaccount:kube-system:ebs-csi-controller-sa",
							},
						},
					},
				},
			},
		})
		if err != nil {
			return "", errors.WithStack(fmt.Errorf("error creating assume role policy: %w", err))
		}
		return assumeRolePolicy.Json, nil
	})

	saRoleName := fmt.Sprintf("ebscsi-%s-sa-role", ctx.Stack())
	saRole, err := iam.NewRole(ctx, saRoleName, &iam.RoleArgs{
		AssumeRolePolicy: assumeRolePolicyDocument,
		Tags: pulumi.StringMap{
			"Project":     pulumi.String("Feature Testing"),
			"Service":     pulumi.String("CI"),
			"Owner":       pulumi.String("pachyderm-ci"),
			"Team":        pulumi.String("Core"),
			"Environment": pulumi.String(ctx.Stack()),
		},
	})

	if err != nil {
		return nil, nil, errors.WithStack(fmt.Errorf("error creating service account role: %w", err))
	}

	_, err = iam.NewRolePolicyAttachment(ctx, "attach-ebs-csi-policy", &iam.RolePolicyAttachmentArgs{
		Role:      saRole,
		PolicyArn: pulumi.String("arn:aws:iam::aws:policy/service-role/AmazonEBSCSIDriverPolicy"),
	})

	if err != nil {
		return nil, nil, errors.WithStack(fmt.Errorf("error attaching policy to service account role: %w", err))
	}

	_, err = awseks.NewAddon(ctx, "aws-ebs-csi-driver", &awseks.AddonArgs{
		ClusterName:           eksCluster.EksCluster.Name(),
		AddonName:             pulumi.String("aws-ebs-csi-driver"),
		ServiceAccountRoleArn: saRole.Arn,
		Tags: pulumi.StringMap{
			"Project":     pulumi.String("Feature Testing"),
			"Service":     pulumi.String("CI"),
			"Owner":       pulumi.String("pachyderm-ci"),
			"Team":        pulumi.String("Core"),
			"Environment": pulumi.String(ctx.Stack()),
		},
	})
	if err != nil {
		return nil, nil, errors.WithStack(fmt.Errorf("error creating EBS CSI driver addon: %w", err))
	}

	_, err = storagev1.NewStorageClass(ctx, "gp3", &storagev1.StorageClassArgs{
		Provisioner: pulumi.String("ebs.csi.aws.com"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("gp3"),
		},
		Parameters: pulumi.StringMap{
			"type":   pulumi.String("gp3"),
			"fsType": pulumi.String("ext4"),
		},
	}, pulumi.Provider(k8sProvider))

	if err != nil {
		return nil, nil, errors.WithStack(fmt.Errorf("error creating storage class: %w", err))
	}

	// Export some values in case they are needed elsewhere
	ctx.Export("kubeconfig", eksCluster.Kubeconfig)
	ctx.Export("vpcId", eksVpc.VpcId)

	return k8sProvider, saRole, nil
}
