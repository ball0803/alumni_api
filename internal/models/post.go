package models

import "time"

var AllowRangeType = []string{
	"event",
	"mentorship",
	"survey",
}

type Post struct {
	Title      string    `json:"title,omitempty" mapstructure:"title" validate:"required,min=3,max=100"`
	Content    string    `json:"content,omitempty" mapstructure:"content" validate:"required,min=10,max=500"`
	PostType   string    `json:"post_type,omitempty" mapstructure:"post_type" validate:"required,oneof=event story job mentorship showcase announcement discussion survey"`
	StartDate  time.Time `json:"start_date,omitempty" mapstructure:"start_date" validate:"omitempty"`
	EndDate    time.Time `json:"end_date,omitempty" mapstructure:"end_date" validate:"omitempty"`
	MediaURL   []string  `json:"media_url,omitempty" mapstructure:"media_url" validate:"omitempty,dive,url"`
	Visibility string    `json:"visibility,omitempty" mapstructure:"visibility" validate:"required,oneof=alumnus admin all"`
}

type UpdatePostRequest struct {
	Title      string    `json:"title,omitempty" mapstructure:"title" validate:"omitempty,min=3,max=50"`
	Content    string    `json:"content,omitempty" mapstructure:"content" validate:"omitempty,min=10,max=500"`
	PostType   string    `json:"post_type,omitempty" mapstructure:"post_type" validate:"omitempty,oneof=event story job mentorship showcase announcement discussion survey"`
	StartDate  time.Time `json:"start_date,omitempty" mapstructure:"start_date" validate:"omitempty"`
	EndDate    time.Time `json:"end_date,omitempty" mapstructure:"end_date" validate:"omitempty"`
	MediaURL   []string  `json:"media_url,omitempty" mapstructure:"media_url" validate:"omitempty,dive,url"`
	Visibility string    `json:"visibility,omitempty" mapstructure:"visibility" validate:"omitempty,oneof=alumnus admin all"`
}

type Comment struct {
	CommentID       string    `json:"comment_id"`
	Content         string    `json:"content"`
	CreatedAt       int64     `json:"created_timestamp"`
	UserID          string    `json:"user_id"`
	Username        string    `json:"username"`
	Name            string    `json:"name,omitempty"`
	ProfilePicture  string    `json:"profile_picture,omitempty"`
	ParentCommentID *string   `json:"parent_comment_id,omitempty"`
	Replies         []Comment `json:"replies,omitempty"`
}

type CommentRequest struct {
	Comment string `json:"comment,omitempty" mapstructure:"comment" validate:"required,max=200"`
}

type Like struct {
	like_id  string
	post_id  string
	user_id  string
	datetime string
}

type Report struct {
	ID         string `json:"id,omitempty" mapstructure:"id" validate:"required,uuid4"`
	UserID     string `json:"user_id,omitempty" mapstructure:"user_id" validate:"required,uuid4"`
	Status     string `json:"status,omitempty" mapstructure:"status" validate:"omitempty,oneof=pending reviewed resolved"`
	Type       string `json:"type,omitempty" mapstructure:"type" validate:"required,oneof=post comment user"`
	Category   string `json:"category,omitempty" mapstructure:"category" validate:"required,oneof=spam harassment hate_speech misinformation other"`
	Additional string `json:"additional,omitempty" mapstructure:"additional" validate:"omitempty,max=200"`
}
