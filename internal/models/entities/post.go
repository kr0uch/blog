package entities

import "time"

type Post struct {
	PostId         string    `json:"post_id"`
	AuthorId       string    `json:"author_id"`
	IdempotencyKey string    `json:"idempotency_key"`
	Title          string    `json:"title"`
	Content        string    `json:"content"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Images         []Image   `json:"images"`
}
