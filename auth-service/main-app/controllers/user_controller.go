package controllers

import (
	db "auth-service/db/sqlconfig"
	"auth-service/lib/configs"
	network "auth-service/lib/net"
	"auth-service/lib/security"
	"auth-service/lib/utils"
	model "auth-service/main-app/models"
	"auth-service/main-app/responses"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
)

var validate = validator.New()

type DIDRequestBody struct {
	Email string `json:"email"`
}

// ^ Login :
//
//	@Summary		Login route
//	@Description	Allows users to login into their account.
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Param			Body	body		model.Login					true	"User's email , password and role"
//	@Success		200		{object}	responses.UserResponse_doc	"Successful response"
//	@Failure		400		{object}	responses.ErrorResponse_doc	"Invalid JSON data"
//	@Failure		400		{object}	responses.ErrorResponse_doc	"Please provide the required credentials"
//	@Failure		401		{object}	responses.ErrorResponse_doc	"Invalid Credentials : Password does not match"
//	@Failure		404		{object}	responses.ErrorResponse_doc	"Email is not registered"
//	@Failure		404		{object}	responses.ErrorResponse_doc	"Email is not registered with the specified role. Registered Role : <role>"
//	@Failure		422		{object}	responses.ErrorResponse_doc	"Email already registered, please verify your email address"
//	@Failure		500		{object}	responses.ErrorResponse_doc	"Internal server error"
//	@Router			/auth/login [post]
func Login(r *gin.Context) {
	r.Writer.Header().Set("Access-Control-Allow-Headers", "*")
	r.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	var req model.Login

	//* Checking for invalid json format
	if err := r.BindJSON(&req); err != nil {
		network.RespondWithError(r, http.StatusBadRequest, "Invalid JSON data")
		return
	}

	//* Validating if all the fields are present
	if validationErr := validate.Struct(&req); validationErr != nil {
		network.RespondWithError(r, http.StatusBadRequest, "Please provide the required credentials.")
		return
	}

	queries := db.New(configs.CONN)
	//* Checking whether the user is registered
	user, userErr := queries.GetUserByEmail(ctx, req.Email)
	if userErr != nil {
		if strings.Contains(userErr.Error(), "no rows in result set") {
			network.RespondWithError(r, http.StatusNotFound, "Email is not registered.")
			return
		}
		log.Println(userErr)
		configs.NotifyAdmin(userErr)
		network.RespondWithError(r, http.StatusInternalServerError, "Internal server error : "+userErr.Error())
		return
	}

	//* Checking Password
	securityErr := security.CheckPassword(req.Password, user.Password)
	if securityErr != nil {
		network.RespondWithError(r, http.StatusUnauthorized, "Invalid Credentials : Password does not match")
		return
	}

	if user.Role != req.Role {
		network.RespondWithError(r, http.StatusNotFound, "Email is not registered with the specified role. Registered Role : "+user.Role)
		return
	}

	//* Checking for verification of the user
	if !user.Isverified {
		network.RespondWithError(r, http.StatusOK, "Email already is not verified. Please verify your email address.")
		return
	}

	//* Generating Token
	token, genJWTErr := security.GenerateJWT()
	if genJWTErr != nil {
		network.RespondWithError(r, http.StatusInternalServerError, "Internal Server Error : "+genJWTErr.Error())
		return
	}

	r.JSON(http.StatusOK, responses.UserResponse{Message: "success", Data: map[string]interface{}{"token": "Bearer " + token}})
}

// ^ Register :
//
//	@Summary		Register route
//	@Description	Allows users to create a new account.
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Param			user	body		model.Register				true	"User email, password and role"
//	@Success		201		{object}	responses.UserResponse_doc	"Successful response"
//	@Failure		400		{object}	responses.ErrorResponse_doc	"Invalid JSON data, Invalid Email"
//	@Failure		401		{object}	responses.ErrorResponse_doc	"Invalid Credentials"
//	@Failure		409		{object}	responses.ErrorResponse_doc	"Email is already registered. Please login"
//	@Failure		422		{object}	responses.ErrorResponse_doc	"Please provide the required credentials"
//	@Failure		500		{object}	responses.ErrorResponse_doc	"Internal Server Error, Error in inserting the document"
//	@Router			/auth/register [post]
func Register(r *gin.Context) {
	r.Writer.Header().Set("Access-Control-Allow-Headers", "*")
	r.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	var user model.User_Model
	defer cancel()
	var queries *db.Queries
	go func() {
		queries = db.New(configs.CONN)
	}()

	//* Checking for invalid json format
	if invalidJsonErr := r.BindJSON(&user); invalidJsonErr != nil {
		network.RespondWithError(r, http.StatusBadRequest, "Invalid JSON data")
		return
	}

	//* Validating if all the fields are present
	if validationErr := validate.Struct(&user); validationErr != nil {
		network.RespondWithError(r, http.StatusUnprocessableEntity, "Please provide the required credentials")
		return
	}

	//* Hashing Password
	hashedPass, hashPassErr := security.HashPassword(user.Password)
	if hashPassErr != nil {
		network.RespondWithError(r, http.StatusInternalServerError, "Internal Server Error : "+hashPassErr.Error())
		return
	}

	//* Generating OTP
	otp, genOtpErr := utils.GenerateOTP()
	if genOtpErr != nil {
		network.RespondWithError(r, http.StatusInternalServerError, "Internal Server Error : "+genOtpErr.Error())
		return
	}

	//* Creating User
	_, insertDBErr := queries.CreateUser(ctx, db.CreateUserParams{
		Email:    user.Email,
		Password: hashedPass,
		Otp:      otp,
		Role:     user.Role,
	})

	//* Checking for errors while inserting in the DB
	if insertDBErr != nil {
		if strings.HasPrefix(insertDBErr.Error(), "ERROR: duplicate key") {
			network.RespondWithError(r, http.StatusConflict, "Email is already registered. Please login")
			return
		} else if strings.Contains(insertDBErr.Error(), "\"valid_email\"") {
			network.RespondWithError(r, http.StatusBadRequest, "Invalid Email")
			return
		}

		log.Println(insertDBErr)
		go func() {
			sendEmailErrAdm := configs.NotifyAdmin(insertDBErr)
			log.Println(sendEmailErrAdm)
		}()
		network.RespondWithError(r, http.StatusInternalServerError, insertDBErr.Error()+"  : Error in inserting the document")
		return
	}

	//* Sending OTP
	if sendEmailErr := network.SendOTP(user.Email, otp); sendEmailErr != nil {
		network.RespondWithError(r, http.StatusInternalServerError, "Internal Server Error : "+sendEmailErr.Error())
		return
	}

	r.JSON(http.StatusCreated, responses.UserResponse{Message: "OTP has been sent to your email"})
}

// ^ Validation :
//
//	@Summary		Validation route
//	@Description	Allows users to validate OTP and complete the registration process.
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Param			Body	body		model.OTP					true	"User's email and otp"
//	@Success		200		{object}	responses.UserResponse_doc	"Email is already verified. Please login."
//	@Failure		400		{object}	responses.ErrorResponse_doc	"Invalid JSON data, Invalid Email"
//	@Failure		404		{object}	responses.ErrorResponse_doc	"Email is not registered. Please register to continue"
//	@Failure		401		{object}	responses.ErrorResponse_doc	"Invalid OTP"
//	@Failure		422		{object}	responses.ErrorResponse_doc	"Please provide the required credentials"
//	@Failure		500		{object}	responses.ErrorResponse_doc	"Internal Server Error"
//	@Router			/auth/otp [post]
func ValidateOTP(r *gin.Context) {
	r.Writer.Header().Set("Access-Control-Allow-Headers", "*")
	r.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	ctx := context.Background()
	var req model.OTP
	//* Checking for invalid json format
	if err := r.BindJSON(&req); err != nil {
		network.RespondWithError(r, http.StatusBadRequest, "Invalid JSON data")
		return
	}

	//* Validating if all the fields are present
	if validationErr := validate.Struct(&req); validationErr != nil {
		network.RespondWithError(r, http.StatusBadRequest, "Please provide the required credentials.")
		return
	}

	queries := db.New(configs.CONN)

	//* Checking whether user exists or not
	user, getUserErr := queries.GetUserByEmail(ctx, req.Email)
	if getUserErr != nil {
		network.RespondWithError(r, http.StatusNotFound, "Email is not registered. Please register to continue.")
		return
	}

	//* Checking if user is already verified
	if user.Isverified {
		network.RespondWithError(r, http.StatusOK, "Email is already verified. Please login.")
		return
	}

	//* Validating OTP
	if user.Otp != req.OTP {
		network.RespondWithError(r, http.StatusUnauthorized, "Invalid OTP")
		return
	}

	//* Updating user to be verified
	updateUserErr := queries.UpdateUser(ctx, req.Email)
	if updateUserErr != nil {
		network.RespondWithError(r, http.StatusInternalServerError, "Internal Server Error : "+updateUserErr.Error())
		return
	}

	//* Creating Wallet
	requestBody := DIDRequestBody{
		Email: user.Email,
	}
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Println("Error marshaling request body:", err)
		network.RespondWithError(r, http.StatusInternalServerError, "Internal Server Error : "+err.Error())
		return
	}
	res, didErr := http.Post(os.Getenv("WALLET_URL"), "application/json", bytes.NewBuffer(jsonBody))
	if didErr != nil {
		log.Println(didErr)
		network.RespondWithError(r, http.StatusInternalServerError, "Unable to generate Wallet")
		return
	}
	var res_suc SuccessResponse
	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}
	err = json.Unmarshal(body, &res_suc)
	if err != nil {
		log.Println("Error unmarshaling response body:", err)
		return
	}

	log.Println(res_suc)

	//* Generating Token
	token, tokenErr := security.GenerateJWT()
	if tokenErr != nil {
		network.RespondWithError(r, http.StatusInternalServerError, "Internal Server Error : "+tokenErr.Error())
		return
	}

	r.JSON(http.StatusOK, responses.UserResponse{Message: "success", Data: map[string]interface{}{"token": "Bearer " + token}})
}

type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
