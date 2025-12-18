package controllers

import (
	"blog/internal/logger"
	"blog/internal/models/dto"
	"blog/internal/models/entities"
	"blog/pkg/consts"
	"blog/pkg/consts/errors"
	"context"
	"encoding/json"
	stderr "errors"
	"net/http"

	"go.uber.org/zap"
)

type PostsService interface {
	CreatePost(ctx context.Context, post *dto.CreatePostRequest) (*dto.CreatePostResponse, error)
	EditPost(ctx context.Context, rows *dto.EditPostRequest) (*dto.EditPostResponse, error)
	PublishPost(ctx context.Context, post *dto.PublishPostRequest) (*dto.PublishPostResponse, error)
	ViewPostsById(ctx context.Context, rows *dto.GetPostsByIdRequest) (*dto.GetPostsResponse, error)
	ViewAllPosts(ctx context.Context) (*dto.GetPostsResponse, error)
	AddImage(ctx context.Context, rows *dto.AddImageToPostRequest) (*dto.AddImageToPostResponse, error)
	DeleteImage(ctx context.Context, rows *dto.DeleteImageFromPostRequest) (*dto.DeleteImageFromPostResponse, error)
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
	reqLogger := logger.LoggerFromContext(r.Context()).WithFields(zap.String("controller", "CreatePost"))

	reqLogger.Info("Create Post")

	user, err := getUserFromCtx(r)
	if err != nil {
		reqLogger.Error("Failed to get user from context", zap.Error(err))
		http.Error(w, err.Error(), http.StatusForbidden)
	}

	if user.Role != consts.AuthorRole {
		reqLogger.Error("User have no permission", zap.Error(err))
		http.Error(w, errors.ErrNoPermission.Error(), http.StatusForbidden)
		return
	}

	var post dto.CreatePostRequest

	err = json.NewDecoder(r.Body).Decode(&post)
	if err != nil {
		reqLogger.Error("Failed to decode request", zap.Error(err))
		http.Error(w, errors.ErrIncorrectData.Error(), http.StatusBadRequest)
		return
	}

	post.AuthorId = user.UserId

	response, err := c.srv.CreatePost(r.Context(), &post)
	if err != nil {
		reqLogger.Error("Failed to create post", zap.Error(err))
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
		reqLogger.Error("Failed to encode response", zap.Error(err))
		http.Error(w, errors.ErrInternalServerError.Error(), http.StatusInternalServerError)
		return
	}

	reqLogger.Info("CreatePost done")
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
	reqLogger := logger.LoggerFromContext(r.Context()).WithFields(zap.String("controller", "AddImageToPost"))

	reqLogger.Info("Add Image To Post")

	user, err := getUserFromCtx(r)
	if err != nil {
		reqLogger.Error("Failed to get user from context", zap.Error(err))
		http.Error(w, err.Error(), http.StatusForbidden)
	}

	if user.Role != consts.AuthorRole {
		reqLogger.Error("User have no permission", zap.Error(err))
		http.Error(w, errors.ErrNoPermission.Error(), http.StatusForbidden)
		return
	}

	var rows dto.AddImageToPostRequest
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		reqLogger.Error("Failed to parse multipart form", zap.Error(err))
		http.Error(w, errors.ErrIncorrectData.Error(), http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("image")
	if err != nil {
		reqLogger.Error("Failed to get image", zap.Error(err))
		http.Error(w, errors.ErrIncorrectData.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	rows.PostId = r.PathValue("postId")
	rows.File = file
	rows.Handler = handler

	response, err := c.srv.AddImage(r.Context(), &rows)
	if err != nil {
		reqLogger.Error("Failed to add image to post", zap.Error(err))
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
		reqLogger.Error("Failed to write response", zap.Error(err))
		http.Error(w, errors.ErrInternalServerError.Error(), http.StatusInternalServerError)
		return
	}

	reqLogger.Info("AddImageToPost done")
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
	reqLogger := logger.LoggerFromContext(r.Context()).WithFields(zap.String("controller", "EditPost"))

	reqLogger.Info("Edit Post")

	user, err := getUserFromCtx(r)
	if err != nil {
		reqLogger.Error("Failed to get user from context", zap.Error(err))
		http.Error(w, err.Error(), http.StatusForbidden)
	}

	if user.Role != consts.AuthorRole {
		reqLogger.Error("User have no permission", zap.Error(err))
		http.Error(w, errors.ErrNoPermission.Error(), http.StatusForbidden)
		return
	}

	var rows dto.EditPostRequest

	err = json.NewDecoder(r.Body).Decode(&rows)
	if err != nil {
		reqLogger.Error("Failed to decode request", zap.Error(err))
		http.Error(w, errors.ErrIncorrectData.Error(), http.StatusBadRequest)
		return
	}
	rows.PostId = r.PathValue("postId")
	rows.AuthorId = user.UserId

	response, err := c.srv.EditPost(r.Context(), &rows)
	if err != nil {
		reqLogger.Error("Failed to edit post", zap.Error(err))
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
		reqLogger.Error("Failed to encode response", zap.Error(err))
		http.Error(w, errors.ErrInternalServerError.Error(), http.StatusInternalServerError)
		return
	}

	reqLogger.Info("EditPost done")
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
	reqLogger := logger.LoggerFromContext(r.Context()).WithFields(zap.String("controller", "DeleteImageFromPost"))

	reqLogger.Info("Delete Image From Post")

	user, err := getUserFromCtx(r)
	if err != nil {
		reqLogger.Error("Failed to get user from context", zap.Error(err))
		http.Error(w, err.Error(), http.StatusForbidden)
	}

	if user.Role != consts.AuthorRole {
		reqLogger.Error("User have no permission", zap.Error(err))
		http.Error(w, errors.ErrNoPermission.Error(), http.StatusForbidden)
		return
	}

	var rows dto.DeleteImageFromPostRequest
	rows.PostId = r.PathValue("postId")
	rows.ImageId = r.PathValue("imageId")

	response, err := c.srv.DeleteImage(r.Context(), &rows)
	if err != nil {
		reqLogger.Error("Failed to delete image from post", zap.Error(err))
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
		reqLogger.Error("Failed to encode response", zap.Error(err))
		http.Error(w, errors.ErrInternalServerError.Error(), http.StatusInternalServerError)
		return
	}

	reqLogger.Info("DeleteImageFromPost done")
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
	reqLogger := logger.LoggerFromContext(r.Context()).WithFields(zap.String("controller", "PublishPost"))

	reqLogger.Info("Publish Post")

	user, err := getUserFromCtx(r)
	if err != nil {
		reqLogger.Error("Failed to get user from context", zap.Error(err))
		http.Error(w, err.Error(), http.StatusForbidden)
	}

	if user.Role != consts.AuthorRole {
		reqLogger.Error("User have no permission", zap.Error(err))
		http.Error(w, errors.ErrNoPermission.Error(), http.StatusForbidden)
		return
	}

	var rows dto.PublishPostRequest
	err = json.NewDecoder(r.Body).Decode(&rows)
	if err != nil {
		reqLogger.Error("Failed to decode request", zap.Error(err))
		http.Error(w, errors.ErrIncorrectData.Error(), http.StatusBadRequest)
		return
	}
	rows.PostId = r.PathValue("postId")
	rows.AuthorId = user.UserId

	response, err := c.srv.PublishPost(r.Context(), &rows)
	if err != nil {
		reqLogger.Error("Failed to publish post", zap.Error(err))
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
		reqLogger.Error("Failed to encode response", zap.Error(err))
		http.Error(w, errors.ErrInternalServerError.Error(), http.StatusInternalServerError)
		return
	}

	reqLogger.Info("PublishPost done")
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
	reqLogger := logger.LoggerFromContext(r.Context()).WithFields(zap.String("controller", "ViewPosts"))

	reqLogger.Info("View Posts")

	user, err := getUserFromCtx(r)
	if err != nil {
		reqLogger.Error("Failed to get user from context", zap.Error(err))
		http.Error(w, err.Error(), http.StatusForbidden)
	}

	switch user.Role {
	case consts.AuthorRole:
		c.AuthorView(w, r)
	case consts.ReaderRole:
		c.ReaderView(w, r)
	default:
		reqLogger.Error("User have no permission", zap.Error(err))
		http.Error(w, errors.ErrNoPermission.Error(), http.StatusForbidden)
		return
	}

	reqLogger.Info("ViewPosts done")
}

func (c *PostsController) AuthorView(w http.ResponseWriter, r *http.Request) {
	reqLogger := logger.LoggerFromContext(r.Context()).WithFields(zap.String("controller", "AuthorView"))

	reqLogger.Info("Author View")

	user, err := getUserFromCtx(r)
	if err != nil {
		reqLogger.Error("Failed to get user from context", zap.Error(err))
		http.Error(w, err.Error(), http.StatusForbidden)
	}

	var posts dto.GetPostsByIdRequest

	posts.UserId = user.UserId
	response, err := c.srv.ViewPostsById(r.Context(), &posts)
	if err != nil {
		reqLogger.Error("Failed to view posts", zap.Error(err))
		http.Error(w, errors.ErrInternalServerError.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		reqLogger.Error("Failed to encode response", zap.Error(err))
		http.Error(w, errors.ErrInternalServerError.Error(), http.StatusInternalServerError)
		return
	}
}

func (c *PostsController) ReaderView(w http.ResponseWriter, r *http.Request) {
	reqLogger := logger.LoggerFromContext(r.Context()).WithFields(zap.String("controller", "ReaderView"))

	reqLogger.Info("Reader View")

	response, err := c.srv.ViewAllPosts(r.Context())
	if err != nil {
		reqLogger.Error("Failed to view posts", zap.Error(err))
		http.Error(w, errors.ErrInternalServerError.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		reqLogger.Error("Failed to encode response", zap.Error(err))
		http.Error(w, errors.ErrInternalServerError.Error(), http.StatusInternalServerError)
		return
	}
}
