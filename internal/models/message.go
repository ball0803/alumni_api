package models

import (
	"alumni_api/pkg/customtypes"
)

type Message struct {
	MessageID       string                        `json:"message_id,omitempty" mapstructure:"message_id" validate:"omitempty,uuid4"`
	SenderID        string                        `json:"sender_id,omitempty" mapstructure:"sender_id" validate:"required,uuid4"`
	ReceiverID      string                        `json:"receiver_id,omitempty" mapstructure:"receiver_id" validate:"required,uuid4"`
	Content         customtypes.Encrypted[string] `json:"content,omitempty" mapstructure:"content" validate:"required"`
	Attachment      customtypes.Encrypted[string] `json:"attachment,omitempty" mapstructure:"attachment,squash" validate:"omitempty,url"`
	CreatedDatetime string                        `json:"created_datetime,omitempty" mapstructure:"created_datetime" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	UpdatedDatetime string                        `json:"updated_datetime,omitempty" mapstructure:"updated_datetime" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
}

type ReplyMessage struct {
	MessageID       string                        `json:"message_id,omitempty" mapstructure:"message_id" validate:"omitempty,uuid4"`
	SenderID        string                        `json:"sender_id,omitempty" mapstructure:"sender_id" validate:"required,uuid4"`
	ReceiverID      string                        `json:"receiver_id,omitempty" mapstructure:"receiver_id" validate:"required,uuid4"`
	ReplyID         string                        `json:"reply_id,omitempty" mapstructure:"reply_id" validate:"required,uuid4"`
	Content         customtypes.Encrypted[string] `json:"content,omitempty" mapstructure:"content" validate:"required"`
	Attachment      customtypes.Encrypted[string] `json:"attachment,omitempty" mapstructure:"attachment,squash" validate:"omitempty,url"`
	CreatedDatetime string                        `json:"created_datetime,omitempty" mapstructure:"created_datetime" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	UpdatedDatetime string                        `json:"updated_datetime,omitempty" mapstructure:"updated_datetime" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
}

type EditMessage struct {
	MessageID  string                        `json:"message_id,omitempty" mapstructure:"message_id" validate:"required,uuid4"`
	Content    customtypes.Encrypted[string] `json:"content,omitempty" mapstructure:"content" validate:"required"`
	Attachment customtypes.Encrypted[string] `json:"attachment,omitempty" mapstructure:"attachment,squash" validate:"omitempty,url"`
}

type DeleteMessage struct {
	MessageID string `json:"message_id,omitempty" mapstructure:"message_id" validate:"required,uuid4"`
}
