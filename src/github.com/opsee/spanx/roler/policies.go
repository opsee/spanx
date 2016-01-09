package roler

const (
	RoleName   = "OpseeRole"
	PolicyName = "OpseePolicy"

	UserPolicy = `{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "iam:CreateRole",
                "iam:GetUser",
                "iam:PutRolePolicy"
            ],
            "Resource": "*"
        }
    ]
}`

	AssumeRolePolicy = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "AWS": "933693344490"
      },
      "Action": "sts:AssumeRole",
      "Condition": {
        "StringEquals": {
          "sts:ExternalId": "%s"
        }
      }
    },
    {
      "Effect": "Allow",
      "Principal": {
        "Service": ["ec2.amazonaws.com"]
      },
      "Action": [ "sts:AssumeRole" ]
    }
  ]
}
`
	Policy = `{
 "Version": "2012-10-17",
 "Statement": [
  {
   "Effect": "Allow",
   "Action": [
    "autoscaling:DescribeLoadBalancers",
    "autoscaling:DescribeAutoScalingGroups",
    "cloudformation:CreateStack",
    "cloudformation:DeleteStack",
    "cloudformation:ListStackResources",
    "ec2:CreateTags",
    "ec2:DeleteTags",
    "ec2:AuthorizeSecurityGroupIngress",
    "ec2:AuthorizeSecurityGroupEgress",
    "ec2:RevokeSecurityGroupIngress",
    "ec2:RevokeSecurityGroupEgress",
    "ec2:StartInstances",
    "ec2:RunInstances",
    "ec2:StopInstances",
    "ec2:RebootInstances",
    "ec2:TerminateInstances",
    "ec2:DescribeAccountAttributes",
    "ec2:DescribeImages",
    "ec2:DescribeSecurityGroups",
    "ec2:CreateSecurityGroup",
    "ec2:DeleteSecurityGroup",
    "ec2:DescribeSubnets",
    "ec2:DescribeVpcs",
    "ec2:DescribeInstances",
    "ec2:DescribeInternetGateways",
    "ec2:DescribeRouteTables",
    "elasticloadbalancing:DescribeLoadBalancers",
    "iam:AddRoleToInstanceProfile",
    "iam:RemoveRoleFromInstanceProfile",
    "iam:CreateRole",
    "iam:CreateRolePolicy",
    "iam:CreateInstanceProfile",
    "iam:PutRolePolicy",
    "iam:DeleteRole",
    "iam:DeleteRolePolicy",
    "iam:DeleteInstanceProfile",
    "iam:PassRole",
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
    "rds:DescribeDBInstances",
    "rds:DescribeDBSecurityGroups"
   ],
   "Resource": "*"
  }
 ]
}`
)
