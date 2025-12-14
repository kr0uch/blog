package entities

import "time"

type Image struct {
	ImageId   string    `json:"image_id"`
	PostId    string    `json:"post_id"`
	ImageURL  string    `json:"image_url"`
	CreatedAt time.Time `json:"created_at"`
}
