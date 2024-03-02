package roles

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/jsii-runtime-go"
)

func CreateInvocationRole(stack awscdk.Stack, sendEmail_handler awslambda.Function) awsiam.Role {
	role := awsiam.NewRole(stack, jsii.String("Invoke-Role"), &awsiam.RoleProps{
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("lambda.amazonaws.com"), &awsiam.ServicePrincipalOpts{}),
	})

	role.AddToPolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Actions:   &[]*string{jsii.String("lambda:InvokeFunction")},
		Resources: &[]*string{jsii.String(*sendEmail_handler.FunctionArn())},
	}))

	return role
}
