package dto

import "blog/internal/models/entities"

type CreatePostRequest struct {
	AuthorId       string `json:"-"`
	IdempotencyKey string `json:"idempotency_key"`
	Title          string `json:"title"`
	Content        string `json:"content"`
}

type CreatePostResponse struct {
	Message string `json:"message"`
}
type EditPostRequest struct {
	AuthorId string `json:"-"`
	PostId   string `json:"-"`
	Title    string `json:"title"`
	Content  string `json:"content"`
}

type EditPostResponse struct {
	Message string `json:"message"`
}

type PublishPostRequest struct {
	AuthorId string `json:"-"`
	PostId   string `json:"-"`
	Status   string `json:"status"`
}

type PublishPostResponse struct {
	Message string `json:"message"`
}

type GetPostsByIdRequest struct {
	AuthorId string `json:"-"`
}

type GetPostsResponse struct {
	Posts []entities.Post `json:"posts"`
}
