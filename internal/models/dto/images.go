package dto

import (
	"io"
	"mime/multipart"
)

type AddImageToPostRequest struct {
	PostId   string                `json:"-"`
	AuthorId string                `json:"-"`
	File     io.Reader             `json:"-"`
	Handler  *multipart.FileHeader `json:"-"`
}
type AddImageToPostResponse struct {
	Message string `json:"message"`
}

type DeleteImageFromPostRequest struct {
	PostId   string `json:"-"`
	AuthorId string `json:"-"`
	ImageId  string `json:"-"`
}

type DeleteImageFromPostResponse struct {
	Message string `json:"message"`
}
