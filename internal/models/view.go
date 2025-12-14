package models

type GetPostsByIdRequest struct {
	UserId string
}

type GetPostsResponse struct {
	Posts []Post `json:"posts"`
}
