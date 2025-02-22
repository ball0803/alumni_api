package models

import (
	"alumni_api/pkg/customtypes"
	"time"
)

type UserRequest struct {
	ID string `json:"id" validate:"required,uuid4"`
}

type UserFulltextSearch struct {
	Name string `json:"name,omitempty" mapstructure:"name,omitempty" validate:"omitempty,min=2,max=50"`
	Mode string `json:"mode,omitempty" mapstructure:"mode,omitempty" validate:"omitempty,oneof=contain fuzzy exact"`
}

type UpdateUserProfileRequest struct {
	Username       string    `json:"username,omitempty" mapstructure:"username,omitempty" validate:"omitempty,min=3,max=50"`
	Password       string    `json:"user_password,omitempty" mapstructure:"user_password,omitempty" validate:"omitempty,min=8"`
	FirstName      string    `json:"first_name,omitempty" mapstructure:"first_name,omitempty" validate:"omitempty,min=2,max=50"`
	LastName       string    `json:"last_name,omitempty" mapstructure:"last_name,omitempty" validate:"omitempty,min=2,max=50"`
	FirstNameEng   string    `json:"first_name_eng,omitempty" mapstructure:"first_name_eng,omitempty" validate:"omitempty,min=2,max=50"`
	LastNameEng    string    `json:"last_name_eng,omitempty" mapstructure:"last_name_eng,omitempty" validate:"omitempty,min=2,max=50"`
	Gender         string    `json:"gender,omitempty" mapstructure:"gender,omitempty" validate:"omitempty,oneof=male female other"`
	DOB            time.Time `json:"dob,omitempty" mapstructure:"dob,omitempty" validate:"omitempty"`
	ProfilePicture string    `json:"profile_picture,omitempty" mapstructure:"profile_picture,omitempty" validate:"omitempty,url"`
	Role           string    `json:"role,omitempty" mapstructure:"role,omitempty" validate:"omitempty,oneof=student alumnus staff"`

	StudentInfo StudentInfo `json:"student_info,omitempty" mapstructure:"student_info,squash" validate:"omitempty"`
	Companies   []Company   `json:"companies,omitempty" mapstructure:"companies" validate:"omitempty,dive"`
	ContactInfo Contact     `json:"contact_info,omitempty" mapstructure:"contact_info,squash" validate:"omitempty"`
}

type CreateUserRequest struct {
	UserID         string    `json:"user_id,omitempty" mapstructure:"user_id" validate:"omitempty,uuid4"`
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

type UserRequestCompany struct {
	Companies []Company `json:"companies" validate:"required"`
}

type UserCompanyUpdateRequest struct {
	Position customtypes.Encrypted[string] `json:"position,omitempty" mapstructure:"position,omitempty" validate:"omitempty,max=100"`
}

type UserFOAFRequest struct {
	Degree int8 `json:"degree,omitempty" mapstructure:"degree" validate:"omitempty,min=1,max=5"`
}
