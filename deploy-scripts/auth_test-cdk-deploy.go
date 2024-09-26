package main

import (
	"log"
	"os"

	"github.com/Manas8803/authTest-cdk-deploy/deploy/roles"
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/joho/godotenv"
)

type AuthTestProps struct {
	awscdk.StackProps
}

func LamdaStack(scope constructs.Construct, id string, props *AuthTestProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	sendEmail_handler := awslambda.NewFunction(stack, jsii.String("email-service"), &awslambda.FunctionProps{
		Code:    awslambda.Code_FromAsset(jsii.String("../email-service"), nil),
		Runtime: awslambda.Runtime_PROVIDED_AL2023(),
		Handler: jsii.String("main"),
		Timeout: awscdk.Duration_Seconds(jsii.Number(10)),
		Environment: &map[string]*string{
			"EMAIL":    jsii.String(os.Getenv("EMAIL")),
			"PASSWORD": jsii.String(os.Getenv("PASSWORD")),
		},
	})

	invoke_role := roles.CreateInvocationRole(stack, sendEmail_handler)

	auth_handler := awslambda.NewFunction(stack, jsii.String("auth-service"), &awslambda.FunctionProps{
		Code:    awslambda.Code_FromAsset(jsii.String("../auth-service"), nil),
		Runtime: awslambda.Runtime_PROVIDED_AL2023(),
		Handler: jsii.String("main"),
		Timeout: awscdk.Duration_Seconds(jsii.Number(10)),
		Role:    invoke_role,
		Environment: &map[string]*string{
			"SQLURI":            jsii.String(os.Getenv("SQLURI")),
			"JWT_SECRET_KEY":    jsii.String(os.Getenv("JWT_SECRET_KEY")),
			"JWT_LIFETIME":      jsii.String(os.Getenv("JWT_LIFETIME")),
			"EMAIL":             jsii.String(os.Getenv("EMAIL")),
			"PASSWORD":          jsii.String(os.Getenv("PASSWORD")),
			"ADMIN":             jsii.String(os.Getenv("ADMIN")),
			"WALLET_URL":        jsii.String(os.Getenv("WALLET_URL")),
			"SEND_TO_EMAIL_ARN": jsii.String(*sendEmail_handler.FunctionArn()),
		},
	})

	awsapigateway.NewLambdaRestApi(stack, jsii.String("BlckPr-auth-1"), &awsapigateway.LambdaRestApiProps{
		Handler: auth_handler,
		DefaultCorsPreflightOptions: &awsapigateway.CorsOptions{
			AllowOrigins: awsapigateway.Cors_ALL_ORIGINS(),
			AllowMethods: awsapigateway.Cors_ALL_METHODS(),
			AllowHeaders: awsapigateway.Cors_DEFAULT_HEADERS(),
		},
	})

	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	LamdaStack(app, "BlckPr-auth-Stack", &AuthTestProps{
		awscdk.StackProps{
			Env: env(),
		},
	})

	app.Synth(nil)
}

func env() *awscdk.Environment {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatalln("Error loading .env file : ", err)
	}

	return &awscdk.Environment{
		Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
		Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
	}
}
