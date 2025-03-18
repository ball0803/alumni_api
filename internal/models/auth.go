package models

import "github.com/golang-jwt/jwt/v5"

type Claims struct {
	UserID       string `json:"user_id"`
	Role         string `json:"role"`
	DepartmentID string `json:"department_id,omitempty"`
	AdmitYear    int    `json:"admit_year,omitempty"`
	jwt.RegisteredClaims
}

type Verify struct {
	UserID            string `json:"user_id,omitempty" mapstructure:"user_id" validate:"required,uuid4"`
	VerificationToken string `json:"token,omitempty" mapstructure:"token"`
	jwt.RegisteredClaims
}

type ResetPassword struct {
	Password string `json:"password,omitempty" mapstructure:"password" validate:"required,min=8"`
	ResetJWT string `json:"token,omitempty" mapstructure:"token"`
}

type ChangeEmail struct {
	Email string `json:"email,omitempty" mapstructure:"email" validate:"required,email"`
}

type LoginResponse struct {
	UserID    string `json:"user_id,omitempty" mapstructure:"user_id" validate:"required,uuid4"`
	Password  string `json:"user_password,omitempty" mapstructure:"user_password" validate:"required,min=8"`
	Role      string `json:"role,omitempty" mapstructure:"role" validate:"required,oneof=student alumni staff visitor"`
	AdmitYear int16  `json:"admit_year,omitempty" mapstructure:"admit_year" validate:"gte=1950,lte=2100"`
}

type LoginRequest struct {
	Username string `json:"username" mapstructure:"username" validate:"required"`
	Password string `json:"password,omitempty" mapstructure:"password" validate:"required,min=8"`
}

type ReqistryRequest struct {
	Username string `json:"username" mapstructure:"username" validate:"required"`
	Password string `json:"password,omitempty" mapstructure:"password" validate:"required,min=8"`
}
