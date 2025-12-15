package controllers

import (
	"blog/internal/models/dto"
	"blog/internal/models/entities"
	"blog/pkg/consts"
	"blog/pkg/consts/errors"
	"encoding/json"
	stderr "errors"
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

func getUserFromCtx(r *http.Request) (*entities.User, error) {
	ctx := r.Context()
	ctxUser := ctx.Value(consts.CtxUser)
	user, ok := ctxUser.(*entities.User)
	if !ok {
		return nil, errors.ErrNoPermission
	}
	return user, nil
}

// CreatePost godoc
// @Summary Создать пост
// @Tags Управление постами
// @Accept json
// @Produce json
// @Param request body dto.CreatePostRequest true "Данные поста"
// @Param Authorization header string true "Токен авторизации"
// @Success 200 {object} dto.CreatePostResponse
// @Failure 409 {string} errors.ErrInvalidIdempotencyKey "invalid idempotency key"
// @Failure 403 {string} errors.ErrNoPermission "no permission"
// @Router /api/posts [post]
func (c *PostsController) CreatePost(w http.ResponseWriter, r *http.Request) {
	user, err := getUserFromCtx(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
	}

	if user.Role != consts.AuthorRole {
		http.Error(w, errors.ErrNoPermission.Error(), http.StatusForbidden)
		return
	}

	var post dto.CreatePostRequest

	err = json.NewDecoder(r.Body).Decode(&post)
	if err != nil {
		log.Println(err)
		http.Error(w, errors.ErrIncorrectData.Error(), http.StatusBadRequest)
		return
	}

	post.AuthorId = user.UserId

	response, err := c.srv.CreatePost(&post)
	if err != nil {
		switch {
		case stderr.Is(err, errors.ErrInvalidIdempotencyKey):
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			http.Error(w, err.Error(), http.StatusForbidden)
		}
		return
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Println(err)
		http.Error(w, errors.ErrInternalServerError.Error(), http.StatusInternalServerError)
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
// @Failure 404 {string} errors.ErrPostNotFound "post not found"
// @Failure 403 {string} errors.ErrNoPermission "no permission"
// @Router /api/posts/{postId}/images [post]
func (c *PostsController) AddImageToPost(w http.ResponseWriter, r *http.Request) {
	user, err := getUserFromCtx(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
	}

	if user.Role != consts.AuthorRole {
		http.Error(w, errors.ErrNoPermission.Error(), http.StatusForbidden)
		return
	}

	var rows dto.AddImageToPostRequest
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		log.Println(err)
		http.Error(w, errors.ErrIncorrectData.Error(), http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("image")
	if err != nil {
		log.Println(err)
		http.Error(w, errors.ErrIncorrectData.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	rows.PostId = r.PathValue("postId")
	rows.File = file
	rows.Handler = handler

	response, err := c.srv.AddImage(&rows)
	if err != nil {
		switch {
		case stderr.Is(err, errors.ErrPostNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusForbidden)
		}
		return
	}
	err = json.NewEncoder(w).Encode(dto.AddImageToPostResponse{
		Message: response.Message,
	})
	if err != nil {
		log.Println(err)
		http.Error(w, errors.ErrInternalServerError.Error(), http.StatusInternalServerError)
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
// @Failure 404 {string} errors.ErrPostNotFound "post not found"
// @Failure 403 {string} errors.ErrNoPermission "no permission"
// @Router /api/posts/{postId} [put]
func (c *PostsController) EditPost(w http.ResponseWriter, r *http.Request) {
	user, err := getUserFromCtx(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
	}

	if user.Role != consts.AuthorRole {
		http.Error(w, errors.ErrNoPermission.Error(), http.StatusForbidden)
		return
	}

	var rows dto.EditPostRequest

	err = json.NewDecoder(r.Body).Decode(&rows)
	if err != nil {
		http.Error(w, errors.ErrIncorrectData.Error(), http.StatusBadRequest)
		return
	}
	rows.PostId = r.PathValue("postId")
	rows.AuthorId = user.UserId

	response, err := c.srv.EditPost(&rows)
	if err != nil {
		switch {
		case stderr.Is(err, errors.ErrPostNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusForbidden)
		}
		return
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Println(err)
		http.Error(w, errors.ErrInternalServerError.Error(), http.StatusInternalServerError)
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
// @Failure 404 {string} errors.ErrPostOrImageNotFound "post or image not found"
// @Failure 403 {string} errors.ErrNoPermission "no permission"
// @Router /api/posts/{postId}/images/{imageId} [delete]
func (c *PostsController) DeleteImageFromPost(w http.ResponseWriter, r *http.Request) {
	user, err := getUserFromCtx(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
	}

	if user.Role != consts.AuthorRole {
		http.Error(w, errors.ErrNoPermission.Error(), http.StatusForbidden)
		return
	}

	var rows dto.DeleteImageFromPostRequest
	rows.PostId = r.PathValue("postId")
	rows.ImageId = r.PathValue("imageId")

	response, err := c.srv.DeleteImage(&rows)
	if err != nil {
		switch {
		case stderr.Is(err, errors.ErrPostOrImageNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusForbidden)
		}
		return
	}
	err = json.NewEncoder(w).Encode(&response)
	if err != nil {
		log.Println(err)
		http.Error(w, errors.ErrInternalServerError.Error(), http.StatusInternalServerError)
		return
	}
}

// PublishPost godoc
// @Summary Опубликовать пост
// @Tags Управление постами
// @Accept json
// @Produce json
// @Param postId path string true "ID поста"
// @Param status body dto.PublishPostRequest true "Статус поста"
// @Param Authorization header string true "Токен авторизации"
// @Success 200 {object} dto.PublishPostResponse
// @Failure 400 {string} errors.ErrInvalidPostStatus "invalid post status"
// @Failure 404 {string} errors.ErrPostNotFound "post not found"
// @Failure 403 {string} errors.ErrNoPermission "no permission"
// @Router /api/posts/{postId}/status [patch]
func (c *PostsController) PublishPost(w http.ResponseWriter, r *http.Request) {
	user, err := getUserFromCtx(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
	}

	if user.Role != consts.AuthorRole {
		http.Error(w, errors.ErrNoPermission.Error(), http.StatusForbidden)
		return
	}

	var rows dto.PublishPostRequest
	err = json.NewDecoder(r.Body).Decode(&rows)
	if err != nil {
		log.Println(err)
		http.Error(w, errors.ErrIncorrectData.Error(), http.StatusBadRequest)
		return
	}
	rows.PostId = r.PathValue("postId")
	rows.AuthorId = user.UserId

	response, err := c.srv.PublishPost(&rows)
	if err != nil {
		switch {
		case stderr.Is(err, errors.ErrInvalidPostStatus):
			http.Error(w, err.Error(), http.StatusBadRequest)
		case stderr.Is(err, errors.ErrPostNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusForbidden)
		}
		return
	}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Println(err)
		http.Error(w, errors.ErrInternalServerError.Error(), http.StatusInternalServerError)
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
	user, err := getUserFromCtx(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
	}

	switch user.Role {
	case consts.AuthorRole:
		c.AuthorView(w, r)
	case consts.ReaderRole:
		c.ReaderView(w, r)
	default:
		http.Error(w, errors.ErrNoPermission.Error(), http.StatusForbidden)
		return
	}
}

func (c *PostsController) AuthorView(w http.ResponseWriter, r *http.Request) {
	user, err := getUserFromCtx(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
	}

	var posts dto.GetPostsByIdRequest

	posts.UserId = user.UserId
	response, err := c.srv.ViewPostsById(&posts)
	if err != nil {
		log.Println(err)
		http.Error(w, errors.ErrInternalServerError.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Println(err)
		http.Error(w, errors.ErrInternalServerError.Error(), http.StatusInternalServerError)
		return
	}
}

func (c *PostsController) ReaderView(w http.ResponseWriter, r *http.Request) {
	response, err := c.srv.ViewAllPosts()
	if err != nil {
		log.Println(err)
		http.Error(w, errors.ErrInternalServerError.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Println(err)
		http.Error(w, errors.ErrInternalServerError.Error(), http.StatusInternalServerError)
		return
	}
}
