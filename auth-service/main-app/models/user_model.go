package models

type User_Model struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
	OTP      string `json:"otp"`
	Role     string `json:"role" validate:"oneof=Issuer Verifier User"`
}

type OTP struct {
	Email string `json:"email" validate:"required"`
	OTP   string `json:"otp" validate:"required"`
}
type Register struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
	Role     string `json:"role" validate:"oneof=Issuer Verifier User"`
}

type Login struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
	Role     string `json:"role" validate:"oneof=Issuer Verifier User"`
}
