package models

type Post struct {
	post_id          string
	user_id          string
	title            string
	content          string
	post_type        string
	media_url        string
	visibility       string
	created_datetime string
	updated_datetime string
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
