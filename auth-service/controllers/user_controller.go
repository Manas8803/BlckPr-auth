package controllers

import (
	"auth-service/auth"
	"auth-service/configs"
	db "auth-service/db/sqlconfig"
	model "auth-service/models"
	"auth-service/responses"
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
)

var validate = validator.New()

// ^ Login :
//
//	@Summary		Login route
//	@Description	Allows users to login into their account.
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Param			Body	body		model.Login					true	"User's email and password"
//	@Success		200		{object}	responses.UserResponse_doc	"Successful response"
//	@Failure		400		{object}	responses.ErrorResponse_doc	"Invalid JSON data"
//	@Failure		400		{object}	responses.ErrorResponse_doc	"Please provide with sufficient credentials"
//	@Failure		401		{object}	responses.ErrorResponse_doc	"Invalid Credentials"
//	@Failure		404		{object}	responses.ErrorResponse_doc	"User is not registered"
//	@Failure		422		{object}	responses.ErrorResponse_doc	"Email already registered, please verify your email address"
//	@Failure		500		{object}	responses.ErrorResponse_doc	"Internal server error"
//	@Router			/auth/login [post]
func Login(r *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	var req model.Login

	//* Checking for invalid json format
	if err := r.BindJSON(&req); err != nil {
		responses.RespondWithError(r, http.StatusBadRequest, "Invalid JSON data")
		return
	}

	//* Validating if all the fields are present
	if validationErr := validate.Struct(&req); validationErr != nil {
		responses.RespondWithError(r, http.StatusBadRequest, "Please provide the required credentials.")
		return
	}

	queries := db.New(configs.CONN)
	//* Checking whether the user is registered
	user, userErr := queries.GetUserByEmail(ctx, req.Email)
	if userErr != nil {
		if strings.Contains(userErr.Error(), "no rows in result set") {
			responses.RespondWithError(r, http.StatusNotFound, "User is not registered.")
			return
		}
		log.Println(userErr)
		configs.NotifyAdmin(userErr)
		responses.RespondWithError(r, http.StatusInternalServerError, "Internal server error : "+userErr.Error())
		return
	}

	//* Checking for verification of the user
	if !user.Isverified {

		responses.RespondWithError(r, http.StatusUnprocessableEntity, "Email is already registered. Please verify your email address using the OTP sent to your registered email.")
		return
	}

	//* Verifying password
	credentialsError := model.CheckPassword(req.Password, user.Password)
	if credentialsError != nil {
		responses.RespondWithError(r, http.StatusUnauthorized, "Invalid Credentials")
		return
	}

	//* Generating Token
	token, genJWTErr := auth.GenerateJWT()
	if genJWTErr != nil {
		responses.RespondWithError(r, http.StatusInternalServerError, "Internal Server Error : "+genJWTErr.Error())
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
//	@Param			user	body		model.Register				true	"User name, email, password"
//	@Success		201		{object}	responses.UserResponse_doc	"Successful response"
//	@Failure		400		{object}	responses.ErrorResponse_doc	"Invalid JSON data, Invalid Email"
//	@Failure		401		{object}	responses.ErrorResponse_doc	"Invalid Credentials"
//	@Failure		409		{object}	responses.ErrorResponse_doc	"User already exists"
//	@Failure		422		{object}	responses.ErrorResponse_doc	"Please provide with sufficient credentials"
//	@Failure		500		{object}	responses.ErrorResponse_doc	"Internal Server Error, Error in inserting the document"
//	@Router			/auth/register [post]
func Register(r *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	var user model.User
	defer cancel()
	var queries *db.Queries
	go func() {
		queries = db.New(configs.CONN)
	}()

	//* Checking for invalid json format
	if invalidJsonErr := r.BindJSON(&user); invalidJsonErr != nil {
		responses.RespondWithError(r, http.StatusBadRequest, "Invalid JSON data")
		return
	}

	//* Validating if all the fields are present
	if validationErr := validate.Struct(&user); validationErr != nil {
		responses.RespondWithError(r, http.StatusUnprocessableEntity, "Please provide the required credentials")
		return
	}

	//* Hashing Password
	if hashPassErr := user.HashPassword(user.Password); hashPassErr != nil {
		responses.RespondWithError(r, http.StatusInternalServerError, "Internal Server Error : "+hashPassErr.Error())
		return
	}

	//* Generating OTP
	if genOtpErr := user.GenerateOTP(); genOtpErr != nil {
		responses.RespondWithError(r, http.StatusInternalServerError, "Internal Server Error : "+genOtpErr.Error())
		return
	}

	//* Creating User
	_, insertDBErr := queries.CreateUser(ctx, db.CreateUserParams{
		Name:     user.Name,
		Email:    user.Email,
		Password: user.Password,
		Otp:      user.OTP,
	})

	//* Checking for errors while inserting in the DB
	if insertDBErr != nil {
		if strings.HasPrefix(insertDBErr.Error(), "ERROR: duplicate key") {
			responses.RespondWithError(r, http.StatusConflict, "User already exists")
			return
		} else if strings.Contains(insertDBErr.Error(), "\"valid_email\"") {
			responses.RespondWithError(r, http.StatusBadRequest, "Invalid Email")
			return
		}

		log.Println(insertDBErr)
		go func() {
			sendEmailErrAdm := configs.NotifyAdmin(insertDBErr)
			log.Println(sendEmailErrAdm)
		}()
		responses.RespondWithError(r, http.StatusInternalServerError, insertDBErr.Error()+"  : Error in inserting the document")
		return
	}

	//* Sending OTP
	if sendEmailErr := model.SendOTP(user.Email, user.OTP); sendEmailErr != nil {
		responses.RespondWithError(r, http.StatusInternalServerError, "Internal Server Error : "+sendEmailErr.Error())
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
//	@Param			Body	body		model.OTP					true	"User's email address and otp"
//	@Success		200		{object}	responses.UserResponse_doc	"Successful response, User already verified. Please login."
//	@Failure		400		{object}	responses.ErrorResponse_doc	"Invalid JSON data, Invalid Email"
//	@Failure		404		{object}	responses.ErrorResponse_doc	"User does not exist. Please register to generate OTP."
//	@Failure		401		{object}	responses.ErrorResponse_doc	"Invalid OTP"
//	@Failure		422		{object}	responses.ErrorResponse_doc	"Please provide with sufficient credentials"
//	@Failure		500		{object}	responses.ErrorResponse_doc	"Internal Server Error"
//	@Router			/auth/otp [post]
func ValidateOTP(r *gin.Context) {
	ctx := context.Background()
	var req model.OTP
	//* Checking for invalid json format
	if err := r.BindJSON(&req); err != nil {
		responses.RespondWithError(r, http.StatusBadRequest, "Invalid JSON data")
		return
	}

	//* Validating if all the fields are present
	if validationErr := validate.Struct(&req); validationErr != nil {
		responses.RespondWithError(r, http.StatusBadRequest, "Please provide the required credentials.")
		return
	}

	queries := db.New(configs.CONN)

	//* Checking whether user exists or not
	user, getUserErr := queries.GetUserByEmail(ctx, req.Email)
	if getUserErr != nil {
		responses.RespondWithError(r, http.StatusNotFound, "User does not exist. Please register to generate OTP.")
		return
	}

	//* Checking if user is already verified
	if user.Isverified {
		responses.RespondWithError(r, http.StatusOK, "User already verified. Please login.")
		return
	}

	//* Validating OTP
	if user.Otp != req.OTP {
		responses.RespondWithError(r, http.StatusUnauthorized, "Invalid OTP")
		return
	}

	//* Updating user to be verified
	log.Println("HELLO: updating user")
	updateUserErr := queries.UpdateUser(ctx, req.Email)
	if updateUserErr != nil {
		responses.RespondWithError(r, http.StatusInternalServerError, "Internal Server Error : "+updateUserErr.Error())
		return
	}

	//* Generating Token
	token, tokenErr := auth.GenerateJWT()
	if tokenErr != nil {
		responses.RespondWithError(r, http.StatusInternalServerError, "Internal Server Error : "+tokenErr.Error())
		return
	}

	r.JSON(http.StatusOK, responses.UserResponse{Message: "success", Data: map[string]interface{}{"token": "Bearer " + token}})
}
