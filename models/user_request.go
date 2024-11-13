package models

import (
	"time"
)

type UserRequest struct {
	ID string `json:"id" validate:"required,uuid4"`
}

type CreateUserRequest struct {
	UserID         string    `json:"user_id,omitempty" mapstructure:"user_id" validate:"required,uuid4"`
	Username       string    `json:"username,omitempty" mapstructure:"username" validate:"required,min=3,max=50"`
	Password       string    `json:"user_password,omitempty" mapstructure:"user_password" validate:"required,min=8"`
	FirstName      string    `json:"first_name,omitempty" mapstructure:"first_name" validate:"required,min=2,max=50"`
	LastName       string    `json:"last_name,omitempty" mapstructure:"last_name" validate:"required,min=2,max=50"`
	FirstNameEng   string    `json:"first_name_eng,omitempty" mapstructure:"first_name_eng" validate:"omitempty,min=2,max=50"`
	LastNameEng    string    `json:"last_name_eng,omitempty" mapstructure:"last_name_eng" validate:"omitempty,min=2,max=50"`
	Gender         string    `json:"gender,omitempty" mapstructure:"gender" validate:"oneof=male female other"`
	DOB            time.Time `json:"dob,omitempty" mapstructure:"dob" validate:"required"`
	ProfilePicture string    `json:"profile_picture,omitempty" mapstructure:"profile_picture" validate:"omitempty,url"`
	Role           string    `json:"role,omitempty" mapstructure:"role" validate:"required,oneof=student alumnus staff"`
}

type UserFriendRequest struct {
	UserID string `json:"user_id,omitempty" mapstructure:"user_id" validate:"required,uuid4"`
}

type UserRequestFilter struct {
	StudentType string `json:"studentType,omitempty" mapstructure:"studentType" validate:"omitempty"`
	Field       string `json:"field,omitempty" mapstructure:"field" validate:"omitempty"`
}

type StudentInfoRequest struct {
	Faculty     string `json:"faculty,omitempty" mapstructure:"faculty" validate:"required"`
	Department  string `json:"department,omitempty" mapstructure:"department" validate:"required"`
	Field       string `json:"field,omitempty" mapstructure:"field" validate:"required"`
	StudentType string `json:"studentType,omitempty" mapstructure:"studentType" validate:"required"`
}

type UserRequestCompany struct {
	Companies []Company `json:"companies" validate:"required"`
}

type UserCompanyUpdateRequest struct {
	Position string `json:"position,omitempty" mapstructure:"position" validate:"required,max=100"`
}
