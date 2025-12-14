package models

// TODO: почистить от ненужных

type CreatePostRequest struct {
	AuthorId       string `json:"author_id"`
	IdempotencyKey string `json:"idempotency_key"`
	Title          string `json:"title"`
	Content        string `json:"content"`
}

type CreatePostResponse struct {
	PostId string `json:"post_id"`
}

type AddImageToPostRequest struct {
	PostId string
}

type AddImageToPostResponse struct {
	PostId string `json:"post_id"`
}

type EditPostRequest struct {
	AuthorId string `json:"-"`
	PostId   string `json:"-"`
	Title    string `json:"title"`
	Content  string `json:"content"`
}

type EditPostResponse struct {
	PostId string `json:"post_id"`
}

type DeleteImageFromPostRequest struct {
	PostId  string `json:"-"`
	ImageId string `json:"-"`
}

type DeleteImageFromPostResponse struct {
	PostId string `json:"post_id"`
}

type PublishPostRequest struct {
	AuthorId string `json:"-"`
	PostId   string `json:"-"`
	Status   string `json:"status"`
}

type PublishPostResponse struct {
	PostId string `json:"post_id"`
}
