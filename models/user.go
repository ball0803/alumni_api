package models

import (
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
	DOB            time.Time   `json:"dob,omitempty" mapstructure:"dob" validate:"required"`
	ProfilePicture string      `json:"profile_picture,omitempty" mapstructure:"profile_picture" validate:"omitempty,url"`
	Role           string      `json:"role,omitempty" mapstructure:"role" validate:"required,oneof=student alumnus staff"`
	StudentInfo    StudentInfo `json:"student_info,omitempty" mapstructure:"student_info,squash" validate:"required"`
	Companies      []Company   `json:"companies,omitempty" mapstructure:"companies" validate:"omitempty,dive"`
	ContactInfo    Contact     `json:"contact_info,omitempty" mapstructure:"contact_info,squash" validate:"omitempty"`
}

type StudentInfo struct {
	Faculty        string  `json:"faculty,omitempty" mapstructure:"faculty" validate:"required"`
	Department     string  `json:"department,omitempty" mapstructure:"department" validate:"required"`
	Field          string  `json:"field,omitempty" mapstructure:"field" validate:"required"`
	StudentType    string  `json:"student_type,omitempty" mapstructure:"student_type" validate:"required"`
	EducationLevel string  `json:"education_level,omitempty" mapstructure:"education_level" validate:"required"`
	AdmitYear      int16   `json:"admit_year,omitempty" mapstructure:"admit_year" validate:"gte=1950,lte=2100"`
	GraduateYear   int16   `json:"graduate_year,omitempty" mapstructure:"graduate_year" validate:"omitempty,gte=2530"`
	GPAX           float32 `json:"gpax,omitempty" mapstructure:"gpax" validate:"omitempty,gte=0,lte=4.0"`
}

type Company struct {
	Company  string `json:"name,omitempty" mapstructure:"name" validate:"omitempty,min=2,max=100"`
	Address  string `json:"address,omitempty" mapstructure:"address" validate:"omitempty,max=200"`
	Position string `json:"position,omitempty" mapstructure:"position" validate:"omitempty,max=100"`
}

type Contact struct {
	Email    string `json:"email,omitempty" mapstructure:"email" validate:"omitempty,email"`
	Github   string `json:"github,omitempty" mapstructure:"github" validate:"omitempty,url"`
	Linkedin string `json:"linkedin,omitempty" mapstructure:"linkedin" validate:"omitempty,url"`
	Facebook string `json:"facebook,omitempty" mapstructure:"facebook" validate:"omitempty,url"`
	Phone    string `json:"phone,omitempty" mapstructure:"phone" validate:"omitempty,e164"`
}

type Message struct {
	Text            string `json:"text,omitempty" mapstructure:"text" validate:"required"`
	CreatedDatetime string `json:"created_datetime,omitempty" mapstructure:"created_datetime" validate:"required,datetime=2006-01-02T15:04:05Z07:00"`
	UpdatedDatetime string `json:"updated_datetime,omitempty" mapstructure:"updated_datetime" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
}

type Friend struct {
	UserID         string `json:"user_id,omitempty" mapstructure:"user_id" validate:"required,uuid4"`
	CreateDatetime string `json:"create_datetime,omitempty" mapstructure:"create_datetime" validate:"required,datetime=2006-01-02T15:04:05Z07:00"`
}
