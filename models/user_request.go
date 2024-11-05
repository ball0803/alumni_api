package models

type UserRequest struct {
	ID string `json:"id" validate:"required,uuid4"`
}
