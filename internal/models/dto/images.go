package dto

import (
	"io"
	"mime/multipart"
)

//TODO: проверить везде json форму

type AddImageToPostRequest struct {
	PostId  string `json:"-"`
	File    io.Reader
	Handler *multipart.FileHeader
}
type AddImageToPostResponse struct {
	ImageId string `json:"image_id"`
}

type DeleteImageFromPostRequest struct {
	PostId  string `json:"-"`
	ImageId string `json:"-"`
}

type DeleteImageFromPostResponse struct {
	PostId string `json:"post_id"`
}
