package responses

import (
	"github.com/gin-gonic/gin"
)

type UserResponse struct {
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

type UserResponse_doc struct {
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data"`
}

type ErrorResponse_doc struct {
	Message string `json:"message"`
}

func RespondWithError(ctx *gin.Context, statusCode int, message string) {
	ctx.JSON(statusCode, UserResponse{
		Message: message,
	})
}
