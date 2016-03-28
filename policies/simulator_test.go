package policies

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSimulator(t *testing.T) {
	assert := assert.New(t)

	responses, err := Simulate([]string{
		"iam:*",
	})

	assert.Error(err)
	if len(responses) != 1 {
		assert.FailNow("template error: ", err.Error())
	}
	assert.Equal("implicitDeny", responses[0].Decision)

	// role and stack stuff only
	_, err = SimulateOnResources([]string{
		"autoscaling:CreateAutoScalingGroup",
		"autoscaling:CreateLaunchConfiguration",
		"autoscaling:CreateOrUpdateTags",
		"autoscaling:DeleteAutoScalingGroup",
		"autoscaling:DeleteLaunchConfiguration",
		"autoscaling:UpdateAutoScalingGroup",
	}, []string{
		"arn:aws:autoscaling:us-west-2:933693344490:autoScalingGroup:58631a25-b3b4-4bed-bfcc-ac7860e53d39:autoScalingGroupName/opsee-stack-7de20034-8356-11e5-ae31-abab0449cea4-OpseeGroup-8Z1DX7BMC2YW",
	})

	assert.NoError(err)

	// role and stack stuff only
	_, err = SimulateOnResources([]string{
		"cloudformation:CreateStack",
		"cloudformation:DeleteStack",
		"cloudformation:UpdateStack",
	}, []string{
		"arn:aws:cloudformation:us-west-2:933693344490:stack/opsee-stack-7de20034-8356-11e5-ae31-abab0449cea4",
	})

	// role and stack stuff only
	_, err = SimulateOnResources([]string{
		"iam:AddRoleToInstanceProfile",
		"iam:CreateInstanceProfile",
		"iam:DeleteInstanceProfile",
		"iam:RemoveRoleFromInstanceProfile",
	}, []string{
		"arn:aws:iam::933693344490:instance-profile/opsee-stack-f537221a-6ba3-11e5-ac08-f717ba66ea33-OpseeInstanceProfile-RHVVG3TFETY8",
	})

	assert.NoError(err)

	_, err = Simulate([]string{
		"autoscaling:DescribeAutoScalingGroups",
		"autoscaling:DescribeAutoScalingInstances",
		"autoscaling:DescribeLaunchConfigurations",
		"autoscaling:DescribeLoadBalancers",
		"cloudfront:ListDistributions",
		"cloudfront:ListStreamingDistributions",
		"cloudfront:GetDistribution",
		"cloudfront:GetDistributionConfig",
		"cloudfront:GetStreamingDistribution",
		"cloudfront:GetStreamingDistributionConfig",
		"cloudwatch:DescribeAlarms",
		"cloudwatch:DescribeAlarmsForMetric",
		"cloudwatch:GetMetricStatistics",
		"cloudwatch:ListMetrics",
		"ec2:AuthorizeSecurityGroupIngress",
		"ec2:CreateSecurityGroup",
		"ec2:CreateTags",
		"ec2:DeleteSecurityGroup",
		"ec2:DescribeAccountAttributes",
		"ec2:DescribeAvailabilityZones",
		"ec2:DescribeImageAttribute",
		"ec2:DescribeImages",
		"ec2:DescribeInstanceAttribute",
		"ec2:DescribeInstances",
		"ec2:DescribeInstanceStatus",
		"ec2:DescribeInternetGateways",
		"ec2:DescribeNatGateways",
		"ec2:DescribeNetworkAcls",
		"ec2:DescribeRegions",
		"ec2:DescribeRouteTables",
		"ec2:DescribeSecurityGroups",
		"ec2:DescribeSubnets",
		"ec2:DescribeTags",
		"ec2:DescribeVpcAttribute",
		"ec2:DescribeVpcs",
		"ec2:ReportInstanceStatus",
		"ec2:RevokeSecurityGroupIngress",
		"ec2:RebootInstances",
		"ec2:RunInstances",
		"ec2:StartInstances",
		"ec2:StopInstances",
		"ec2:TerminateInstances",
		"ecs:DeregisterTaskDefinition",
		"ecs:DescribeClusters",
		"ecs:DescribeContainerInstances",
		"ecs:DescribeServices",
		"ecs:DescribeTaskDefinition",
		"ecs:DescribeTasks",
		"ecs:ListClusters",
		"ecs:ListContainerInstances",
		"ecs:ListServices",
		"ecs:ListTaskDefinitionFamilies",
		"ecs:ListTaskDefinitions",
		"ecs:ListTasks",
		"ecs:RegisterTaskDefinition",
		"ecs:RunTask",
		"ecs:StartTask",
		"ecs:StopTask",
		"ecs:UpdateService",
		"elasticloadbalancing:DescribeInstanceHealth",
		"elasticloadbalancing:DescribeLoadBalancerAttributes",
		"elasticloadbalancing:DescribeLoadBalancers",
		"route53:GetHostedZone",
		"route53:ListHostedZones",
		"route53:ListResourceRecordSets",
		"route53domains:GetDomainDetail",
		"route53domains:ListDomains",
		"rds:DescribeAccountAttributes",
		"rds:DescribeDBClusters",
		"rds:DescribeDBInstances",
		"rds:DescribeDBLogFiles",
		"rds:DescribeDBSecurityGroups",
		"rds:DescribeDBSubnetGroups",
		"sns:CreateTopic",
		"sns:DeleteTopic",
		"sns:Subscribe",
		"sns:Unsubscribe",
		"sns:Publish",
		"sqs:CreateQueue",
		"sqs:DeleteQueue",
		"sqs:DeleteMessage",
		"sqs:ReceiveMessage",
		"sqs:GetQueueAttributes",
		"sqs:SetQueueAttributes",
	})

	assert.NoError(err)
}
