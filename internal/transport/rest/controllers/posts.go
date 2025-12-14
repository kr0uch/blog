package controllers

import (
	"blog/internal/models/dto"
	"blog/internal/models/entities"
	"blog/pkg/consts"
	"encoding/json"
	"log"
	"net/http"
)

type PostsService interface {
	CreatePost(post *dto.CreatePostRequest) (*dto.CreatePostResponse, error)
	EditPost(rows *dto.EditPostRequest) (*dto.EditPostResponse, error)
	PublishPost(post *dto.PublishPostRequest) (*dto.PublishPostResponse, error)
	ViewPostsById(rows *dto.GetPostsByIdRequest) (*dto.GetPostsResponse, error)
	ViewAllPosts() (*dto.GetPostsResponse, error)
	AddImage(rows *dto.AddImageToPostRequest) (*dto.AddImageToPostResponse, error)
	DeleteImage(rows *dto.DeleteImageFromPostRequest) (*dto.DeleteImageFromPostResponse, error)
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
// @Param request body dto.CreatePostRequest true "Данные поста"
// @Param Authorization header string true "Токен авторизации"
// @Success 200 {object} dto.CreatePostResponse
// @Router /api/posts [post]
func (c *PostsController) CreatePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctxUser := ctx.Value(consts.CtxUser)
	user, ok := ctxUser.(*entities.User)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if user.Role != consts.AuthorRole {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var post dto.CreatePostRequest

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
		err = json.NewEncoder(w).Encode(dto.ErrorResponse{Error: err.Error()})
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
// @Accept multipart/form-data
// @Produce json
// @Param postId path string true "ID поста"
// @Param file formData file true "Картинка"
// @Param Authorization header string true "Токен авторизации"
// @Success 200 {object} dto.AddImageToPostResponse
// @Router /api/posts/{postId}/images [post]
func (c *PostsController) AddImageToPost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctxUser := ctx.Value(consts.CtxUser)
	user, ok := ctxUser.(*entities.User)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if user.Role != consts.AuthorRole {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var rows dto.AddImageToPostRequest
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println(err)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("formFile getting", err)
		return
	}
	defer file.Close()

	rows.PostId = r.PathValue("postId")
	rows.File = file
	rows.Handler = handler

	image, err := c.srv.AddImage(&rows)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println(err)
		return
	}
	err = json.NewEncoder(w).Encode(dto.AddImageToPostResponse{
		ImageId: image.ImageId,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
}

// EditPost godoc
// @Summary Редактировать пост
// @Tags Управление постами
// @Accept json
// @Produce json
// @Param postId path string true "ID поста"
// @Param request body dto.EditPostRequest true "Новое название поста"
// @Param Authorization header string true "Токен авторизации"
// @Success 200 {object} dto.EditPostResponse
// @Router /api/posts/{postId} [put]
func (c *PostsController) EditPost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctxUser := ctx.Value(consts.CtxUser)
	user, ok := ctxUser.(*entities.User)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if user.Role != consts.AuthorRole {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	var rows dto.EditPostRequest
	err := json.NewDecoder(r.Body).Decode(&rows)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println(err)
		return
	}
	rows.PostId = r.PathValue("postId")
	rows.AuthorId = user.UserId

	post, err := c.srv.EditPost(&rows)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println(err)
		err = json.NewEncoder(w).Encode(dto.ErrorResponse{Error: err.Error()})
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
// @Success 200 {object} dto.DeleteImageFromPostResponse
// @Router /api/posts/{postId}/images/{imageId} [delete]
func (c *PostsController) DeleteImageFromPost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctxUser := ctx.Value(consts.CtxUser)
	user, ok := ctxUser.(*entities.User)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if user.Role != consts.AuthorRole {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var rows dto.DeleteImageFromPostRequest
	rows.PostId = r.PathValue("postId")
	rows.ImageId = r.PathValue("imageId")

	postId, err := c.srv.DeleteImage(&rows)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println(err)
		return
	}
	err = json.NewEncoder(w).Encode(&postId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
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
// @Param status body dto.PublishPostRequest true "Статус поста"
// @Param Authorization header string true "Токен авторизации"
// @Success 200 {object} dto.PublishPostResponse
// @Router /api/posts/{postId}/status [patch]
func (c *PostsController) PublishPost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctxUser := ctx.Value(consts.CtxUser)
	user, ok := ctxUser.(*entities.User)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if user.Role != consts.AuthorRole {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var rows dto.PublishPostRequest
	err := json.NewDecoder(r.Body).Decode(&rows)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println(err)
		return
	}
	rows.PostId = r.PathValue("postId")
	rows.AuthorId = user.UserId

	post, err := c.srv.PublishPost(&rows)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println(err)
		err = json.NewEncoder(w).Encode(dto.ErrorResponse{Error: err.Error()})
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

// ViewPosts godoc
// @Summary Просмотр постов
// @Tags Просмотр постов
// @Accept json
// @Produce json
// @Param Authorization header string true "Токен авторизации"
// @Success 200 {object} dto.GetPostsResponse
// @Router /api/posts [get]
func (c *PostsController) ViewPosts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctxUser := ctx.Value(consts.CtxUser)
	user, ok := ctxUser.(*entities.User)
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

func (c *PostsController) AuthorView(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctxUser := ctx.Value(consts.CtxUser)
	user, _ := ctxUser.(*entities.User)

	var posts dto.GetPostsByIdRequest
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

func (c *PostsController) ReaderView(w http.ResponseWriter, r *http.Request) {
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
