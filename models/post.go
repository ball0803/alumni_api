package models

type Post struct {
	Title      string `json:"title,omitempty" mapstructure:"title" validate:"required,min=3,max=50"`
	Content    string `json:"content,omitempty" mapstructure:"content" validate:"required,min=10,max=500"`
	PostType   string `json:"post_type,omitempty" mapstructure:"post_type" validate:"required,oneof=event story job mentorship showcase donation_campaign announcement discussion survey"`
	MediaURL   string `json:"media_url,omitempty" mapstructure:"media_url" validate:"omitempty,url"`
	Visibility string `json:"visibility,omitempty" mapstructure:"visibility" validate:"required,oneof=alumnus admin all"`
}

type UpdatePostRequest struct {
	Title      string `json:"title,omitempty" mapstructure:"title" validate:"omitempty,min=3,max=50"`
	Content    string `json:"content,omitempty" mapstructure:"content" validate:"omitempty,min=10,max=500"`
	PostType   string `json:"post_type,omitempty" mapstructure:"post_type" validate:"omitempty,oneof=event story job mentorship showcase donation_campaign announcement discussion survey"`
	MediaURL   string `json:"media_url,omitempty" mapstructure:"media_url" validate:"omitempty,url"`
	Visibility string `json:"visibility,omitempty" mapstructure:"visibility" validate:"omitempty,oneof=alumnus admin all"`
}

type Comment struct {
	comment_id       string
	post_id          string
	user_id          string
	comment          string
	created_datetime string
	reply_id         string
	updated_datetime string
}

type Like struct {
	like_id  string
	post_id  string
	user_id  string
	datetime string
}
