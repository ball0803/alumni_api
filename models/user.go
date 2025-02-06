package models

import (
	"alumni_api/customtypes"
	"time"
)

type UserProfile struct {
	UserID         string      `json:"user_id,omitempty" mapstructure:"user_id" validate:"required,uuid4"`
	Username       string      `json:"username,omitempty" mapstructure:"username" validate:"required,min=3,max=50"`
	Password       string      `json:"user_password,omitempty" mapstructure:"user_password" validate:"required,min=8"`
	FirstName      string      `json:"first_name,omitempty" mapstructure:"first_name" validate:"required,min=2,max=50"`
	LastName       string      `json:"last_name,omitempty" mapstructure:"last_name" validate:"required,min=2,max=50"`
	FirstNameEng   string      `json:"first_name_eng,omitempty" mapstructure:"first_name_eng" validate:"omitempty,min=2,max=50"`
	LastNameEng    string      `json:"last_name_eng,omitempty" mapstructure:"last_name_eng" validate:"omitempty,min=2,max=50"`
	Gender         string      `json:"gender,omitempty" mapstructure:"gender" validate:"oneof=male female other"`
	DOB            time.Time   `json:"DOB,omitempty" mapstructure:"DOB" validate:"required"`
	ProfilePicture string      `json:"profile_picture,omitempty" mapstructure:"profile_picture" validate:"omitempty,url"`
	Role           string      `json:"role,omitempty" mapstructure:"role" validate:"required,oneof=student alumnus staff"`
	StudentInfo    StudentInfo `json:"student_info,omitempty" mapstructure:"student_info,squash" validate:"required"`
	Companies      []Company   `json:"companies,omitempty" mapstructure:"companies" validate:"omitempty,dive"`
	ContactInfo    Contact     `json:"contact_info,omitempty" mapstructure:"contact_info,squash" validate:"omitempty"`
}

type StudentInfo struct {
	CollegeInfo    CollegeInfo                    `json:"college_info,omitempty" mapstructure:"college_info,squash" validate:"omitempty"`
	EducationLevel string                         `json:"education_level,omitempty" mapstructure:"education_level,omitempty" validate:"omitempty"`
	AdmitYear      customtypes.Encrypted[int16]   `json:"admit_year,omitempty" mapstructure:"admit_year,omitempty" validate:"omitempty,gte=2510,lte=2600"`
	GraduateYear   customtypes.Encrypted[int16]   `json:"graduate_year,omitempty" mapstructure:"graduate_year,omitempty" validate:"omitempty,gte=2530"`
	GPAX           customtypes.Encrypted[float32] `json:"gpax,omitempty" mapstructure:"gpax,omitempty" validate:"omitempty,gte=0.0,lte=4.0"`
}

type CollegeInfo struct {
	Faculty     string `json:"faculty,omitempty" mapstructure:"faculty,omitempty" validate:"required"`
	Department  string `json:"department,omitempty" mapstructure:"department,omitempty" validate:"required"`
	Field       string `json:"field,omitempty" mapstructure:"field,omitempty" validate:"required"`
	StudentType string `json:"student_type,omitempty" mapstructure:"student_type,omitempty" validate:"required"`
}

type Company struct {
	Company  customtypes.Encrypted[string] `json:"name,omitempty" mapstructure:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Address  customtypes.Encrypted[string] `json:"address,omitempty" mapstructure:"address,omitempty" validate:"omitempty,max=200"`
	Position customtypes.Encrypted[string] `json:"position,omitempty" mapstructure:"position,omitempty" validate:"omitempty,max=100"`
}

type Contact struct {
	Email    customtypes.Encrypted[string] `json:"email,omitempty" mapstructure:"email,squash" validate:"omitempty,email"`
	Github   customtypes.Encrypted[string] `json:"github,omitempty" mapstructure:"github,squash" validate:"omitempty,url"`
	Linkedin customtypes.Encrypted[string] `json:"linkedin,omitempty" mapstructure:"linkedin,squash" validate:"omitempty,url"`
	Facebook customtypes.Encrypted[string] `json:"facebook,omitempty" mapstructure:"facebook,squash" validate:"omitempty,url"`
	Phone    customtypes.Encrypted[string] `json:"phone,omitempty" mapstructure:"phone,squash" validate:"omitempty,e164"`
}

type Friend struct {
	UserID         string `json:"user_id,omitempty" mapstructure:"user_id" validate:"required,uuid4"`
	CreateDatetime string `json:"create_datetime,omitempty" mapstructure:"create_datetime" validate:"required,datetime=2006-01-02T15:04:05Z07:00"`
}
