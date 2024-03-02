package model

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"golang.org/x/crypto/bcrypt"
)

type UserResponse struct {
	Status  int                    `json:"status"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

type User struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
	OTP      string `json:"otp"`
}

type OTP struct {
	Email string `json:"email" validate:"required"`
	OTP   string `json:"otp" validate:"required"`
}
type Register struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type Login struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type Payload_Body struct {
	Body string `json:"body"`
}

func (user *User) HashPassword(password string) error {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 8)
	if err != nil {
		return err
	}
	user.Password = string(bytes)
	return nil
}

func CheckPassword(providedPassword string, userPassword string) error {
	err := bcrypt.CompareHashAndPassword([]byte(userPassword), []byte(providedPassword))
	if err != nil {
		return err
	}
	return nil
}

func (user *User) GenerateOTP() error {
	randomBytes := make([]byte, 2)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return err
	}

	otp := fmt.Sprintf("%06d", int(randomBytes[0])<<8|int(randomBytes[1])%1000000)
	user.OTP = otp

	return nil
}

func SendOTP(email string, otp string) error {
	sess, err := session.NewSession()
	if err != nil {
		log.Println("Error in creating session : ", err.Error())
		return err
	}

	client := lambda.New(sess)
	data, err := json.Marshal(OTP{Email: email, OTP: otp})
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
