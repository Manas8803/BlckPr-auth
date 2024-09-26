package network

import (
	"auth-service/main-app/responses"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

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

// EmailData represents the structure for the email content
type EmailData struct {
	Email   string    `json:"email"`
	Message EmailBody `json:"message"`
}

// EmailBody represents the structure for the email body
type EmailBody struct {
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// SendOTP sends an OTP to the specified email using a POST API call
func SendOTP(email string, otp string) error {
	// Set the subject and body of the email
	emailData := EmailData{
		Email: email,
		Message: EmailBody{
			Subject: "OTP Verification",
			Body:    fmt.Sprintf("<p>Your OTP for verification is: <strong>%s</strong></p>", otp), // Create a simple OTP email body
		},
	}

	// Convert emailData to JSON
	payloadBytes, err := json.Marshal(emailData)
	if err != nil {
		return fmt.Errorf("failed to marshal email data: %v", err)
	}

	// Prepare the API request
	req, err := http.NewRequest("POST", "https://q648rhgza1.execute-api.ap-south-1.amazonaws.com/prod/", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}
	defer resp.Body.Close()

	// Check for response errors
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("email API responded with status: %v", resp.Status)
	}

	return nil
}
