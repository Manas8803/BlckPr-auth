package network

import (
	"auth-service/main-app/models"
	"auth-service/main-app/responses"
	"encoding/json"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/gin-gonic/gin"
)

type Payload_Body struct {
	Body string `json:"body"`
}

func RespondWithError(ctx *gin.Context, statusCode int, message string) {
	ctx.JSON(statusCode, responses.UserResponse{
		Message: message,
	})
}

func SendOTP(email string, otp string) error {
	sess, err := session.NewSession()
	if err != nil {
		log.Println("Error in creating session : ", err.Error())
		return err
	}

	client := lambda.New(sess)
	data, err := json.Marshal(models.OTP{Email: email, OTP: otp})
	if err != nil {
		log.Println("Error in marshalling data : ", err.Error())
	}

	body := Payload_Body{Body: string(data)}

	payload, err := json.Marshal(body)
	if err != nil {
		log.Println("Error in marshalling payload : ", err.Error())
	}

	input := &lambda.InvokeInput{
		FunctionName:   aws.String(os.Getenv("SEND_TO_EMAIL_ARN")),
		Payload:        payload,
		InvocationType: aws.String("Event"),
	}

	result, err := client.Invoke(input)
	if err != nil {
		log.Println("Error invoking Lambda function:", err)
	} else {
		log.Println("Lambda function invoked successfully:", result)
	}
	return err
}
