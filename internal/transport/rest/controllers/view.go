package controllers

import (
	"blog/internal/models"
	"blog/pkg/consts"
	"encoding/json"
	"log"
	"net/http"
)

type ViewPostsService interface {
	ViewPostsById(raws *models.GetPostsByIdRequest) (*models.GetPostsResponse, error)
	ViewAllPosts() (*models.GetPostsResponse, error)
}

type ViewPostsController struct {
	srv ViewPostsService
}

func NewViewController(srv ViewPostsService) *ViewPostsController {
	return &ViewPostsController{
		srv: srv,
	}
}

// ViewPosts godoc
// @Summary Просмотр постов
// @Tags Просмотр постов
// @Accept json
// @Produce json
// @Param Authorization header string true "Токен авторизации"
// @Success 200 {object} models.GetPostsResponse
// @Router /api/posts [get]
func (c *ViewPostsController) ViewPosts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctxUser := ctx.Value(consts.CtxUser)
	user, ok := ctxUser.(*models.User)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	switch user.Role {
	case consts.AuthorRole:
		c.AuthorView(w, r)
	case consts.ReaderRole:
		c.ReaderView(w, r)
	default:
		w.WriteHeader(http.StatusForbidden)
		return
	}
}

func (c *ViewPostsController) AuthorView(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctxUser := ctx.Value(consts.CtxUser)
	user, _ := ctxUser.(*models.User)

	var posts models.GetPostsByIdRequest
	posts.UserId = user.UserId
	response, err := c.srv.ViewPostsById(&posts)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
}

func (c *ViewPostsController) ReaderView(w http.ResponseWriter, r *http.Request) {
	response, err := c.srv.ViewAllPosts()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
}
