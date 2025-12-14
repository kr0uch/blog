package controllers

import (
	"blog/internal/models"
	"blog/pkg/consts"
	"encoding/json"
	"log"
	"net/http"
)

type PostsService interface {
	CreatePost(post *models.CreatePostRequest) (*models.CreatePostResponse, error)
	EditPost(raws *models.EditPostRequest) (*models.EditPostResponse, error)
	PublishPost(post *models.PublishPostRequest) (*models.PublishPostResponse, error)
}

type PostsController struct {
	srv PostsService
}

func NewPostsController(srv PostsService) *PostsController {
	return &PostsController{
		srv: srv,
	}
}

//TODO: коды ответов переделать + ошибки через http.Error

// CreatePost godoc
// @Summary Создать пост
// @Tags Управление постами
// @Accept json
// @Produce json
// @Param request body models.CreatePostRequest true "Данные поста"
// @Param Authorization header string true "Токен авторизации"
// @Success 200 {object} models.CreatePostResponse
// @Router /api/posts [post]
func (c *PostsController) CreatePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctxUser := ctx.Value(consts.CtxUser)
	user, ok := ctxUser.(*models.User)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if user.Role != consts.AuthorRole {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var post models.CreatePostRequest

	err := json.NewDecoder(r.Body).Decode(&post)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println(err)
		return
	}

	post.AuthorId = user.UserId

	response, err := c.srv.CreatePost(&post)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println(err)
		err = json.NewEncoder(w).Encode(models.ErrorResponse{Error: err.Error()})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
		}
		return
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
}

// AddImageToPost godoc
// @Summary Добавить картинку к посту
// @Tags Управление постами
// @Accept json
// @Produce json
// @Param postId path models.AddImageToPostRequest true "ID поста"
// @Param Authorization header string true "Токен авторизации"
// @Success 200 {object} models.AddImageToPostResponse
// @Router /api/posts/{postId}/images [post]
func (c *PostsController) AddImageToPost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctxUser := ctx.Value(consts.CtxUser)
	user, ok := ctxUser.(*models.User)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	log.Println(user.UserId)
}

// EditPost godoc
// @Summary Редактировать пост
// @Tags Управление постами
// @Accept json
// @Produce json
// @Param postId path string true "ID поста"
// @Param request body models.EditPostRequest true "Новое название поста"
// @Param Authorization header string true "Токен авторизации"
// @Success 200 {object} models.EditPostResponse
// @Router /api/posts/{postId} [put]
func (c *PostsController) EditPost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctxUser := ctx.Value(consts.CtxUser)
	user, ok := ctxUser.(*models.User)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if user.Role != consts.AuthorRole {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	var raws models.EditPostRequest
	err := json.NewDecoder(r.Body).Decode(&raws)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println(err)
		return
	}
	raws.PostId = r.PathValue("postId")
	raws.AuthorId = user.UserId

	post, err := c.srv.EditPost(&raws)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println(err)
		err = json.NewEncoder(w).Encode(models.ErrorResponse{Error: err.Error()})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
		}
		return
	}
	err = json.NewEncoder(w).Encode(post)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
}

// DeleteImageFromPost godoc
// @Summary Удалить картинку из поста
// @Tags Управление постами
// @Accept json
// @Produce json
// @Param postId path string true "ID поста"
// @Param imageId path string true "ID картинки"
// @Param Authorization header string true "Токен авторизации"
// @Success 200 {object} models.DeleteImageFromPostResponse
// @Router /api/posts/{postId}/images/{imageId} [delete]
func (c *PostsController) DeleteImageFromPost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctxUser := ctx.Value(consts.CtxUser)
	user, ok := ctxUser.(*models.User)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if user.Role != consts.AuthorRole {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
}

//TODO: Обернуть получение юзера из контекста в функцию

// PublishPost godoc
// @Summary Опубликовать пост
// @Tags Управление постами
// @Accept json
// @Produce json
// @Param postId path string true "ID поста"
// @Param status body models.PublishPostRequest true "Статус поста"
// @Param Authorization header string true "Токен авторизации"
// @Success 200 {object} models.PublishPostResponse
// @Router /api/posts/{postId}/status [patch]
func (c *PostsController) PublishPost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctxUser := ctx.Value(consts.CtxUser)
	user, ok := ctxUser.(*models.User)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if user.Role != consts.AuthorRole {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var raws models.PublishPostRequest
	err := json.NewDecoder(r.Body).Decode(&raws)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println(err)
		return
	}
	raws.PostId = r.PathValue("postId")
	raws.AuthorId = user.UserId

	post, err := c.srv.PublishPost(&raws)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println(err)
		err = json.NewEncoder(w).Encode(models.ErrorResponse{Error: err.Error()})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
		}
		return
	}
	err = json.NewEncoder(w).Encode(post)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
}
