package service

import (
	"blog/internal/logger"
	"blog/internal/models/dto"
	"blog/internal/models/entities"
	"blog/pkg/consts"
	"blog/pkg/consts/errors"
	"context"
	"fmt"
	"io"
	"time"

	"go.uber.org/zap"
)

type PostsBlogRepository interface {
	CreatePost(authorId, idempotencyKey, title, content, status string, createdAt, updatedAt time.Time) (*entities.Post, error)
	GetUserById(userId string) (*entities.User, error)
	GetPostById(postId string) (*entities.Post, error)
	EditPost(postId, authorId, idempotencyKey, title, content, status string, createdAt, updatedAt time.Time) (*entities.Post, error)

	GetPostsByUserId(userId string) ([]*entities.Post, error)
	GetAllPosts() ([]*entities.Post, error)

	AddImage(postId, imageURL string, createdAt time.Time) (*entities.Image, error)
	SetImageURLById(imageId, URL string) error
	GetImageById(imageId string) (*entities.Image, error)
	DeleteImageById(imageId string) error
}

type MinioRepository interface {
	Upload(ctx context.Context, bucket, filename string, file io.Reader, size int64) (string, error)
	GenerateURL(ctx context.Context, bucket, filename string, expires time.Duration) (string, error)
	DeleteImage(ctx context.Context, bucket, filename string) error
}

type PostsService struct {
	repo   PostsBlogRepository
	minio  MinioRepository
	bucket string
}

func NewPostsService(repo PostsBlogRepository, minio MinioRepository, bucket string) *PostsService {
	return &PostsService{
		repo:   repo,
		minio:  minio,
		bucket: bucket,
	}
}

func (s *PostsService) CreatePost(ctx context.Context, post *dto.CreatePostRequest) (*dto.CreatePostResponse, error) {
	reqLogger := logger.LoggerFromContext(ctx).WithFields(zap.String("operation", "CreatePost"))

	reqLogger.Info("Create post")

	newPost, err := s.repo.CreatePost(post.AuthorId, post.IdempotencyKey, post.Title, post.Content, consts.DraftState, time.Now(), time.Now())
	if err != nil {
		reqLogger.Error("Failed to create post", zap.Error(err))
		return nil, err
	}

	var message string
	if newPost != nil {
		message = "Post created successfully"
	}

	response := &dto.CreatePostResponse{
		Message: message,
	}

	reqLogger.Info("Create post done")

	return response, nil
}

func (s *PostsService) EditPost(ctx context.Context, rows *dto.EditPostRequest) (*dto.EditPostResponse, error) {
	reqLogger := logger.LoggerFromContext(ctx).WithFields(zap.String("operation", "EditPost"))

	reqLogger.Info("Edit Post")

	post, err := s.repo.GetPostById(rows.PostId)
	if err != nil {
		reqLogger.Error("Failed to get post by id", zap.Error(err))
		return nil, errors.ErrPostNotFound
	}
	if post.AuthorId != rows.AuthorId {
		reqLogger.Error("Author id does not match", zap.String("AuthorId", rows.AuthorId))
		return nil, errors.ErrInvalidUser
	}

	newPost, err := s.repo.EditPost(post.PostId, post.AuthorId, post.IdempotencyKey, rows.Title, rows.Content, post.Status, post.CreatedAt, time.Now())
	if err != nil {
		reqLogger.Error("Failed to edit post", zap.Error(err))
		return nil, err
	}

	var message string
	if newPost != nil {
		message = "Post edited successfully"
	}

	response := &dto.EditPostResponse{
		Message: message,
	}

	reqLogger.Info("Edit post done")

	return response, nil
}

func (s *PostsService) PublishPost(ctx context.Context, rows *dto.PublishPostRequest) (*dto.PublishPostResponse, error) {
	reqLogger := logger.LoggerFromContext(ctx).WithFields(zap.String("operation", "PublishPost"))

	reqLogger.Info("Publish Post")

	if rows.Status != consts.PublishedState {
		reqLogger.Error("Post status is not published", zap.String("Status", rows.Status))
		return nil, errors.ErrInvalidPostStatus
	}

	post, err := s.repo.GetPostById(rows.PostId)
	if err != nil {
		reqLogger.Error("Failed to get post by id", zap.Error(err))
		return nil, errors.ErrPostNotFound
	}
	if post.AuthorId != rows.AuthorId {
		reqLogger.Error("Author id does not match", zap.String("AuthorId", rows.AuthorId))
		return nil, errors.ErrInvalidUser
	}

	newPost, err := s.repo.EditPost(post.PostId, post.AuthorId, post.IdempotencyKey, post.Title, post.Content, rows.Status, post.CreatedAt, time.Now())
	if err != nil {
		reqLogger.Error("Failed to edit post", zap.Error(err))
		return nil, err
	}

	var message string
	if newPost != nil {
		message = "Post published successfully"
	}

	response := &dto.PublishPostResponse{
		Message: message,
	}

	reqLogger.Info("Publish post done")

	return response, nil
}

func (s *PostsService) ViewPostsById(ctx context.Context, rows *dto.GetPostsByIdRequest) (*dto.GetPostsResponse, error) {
	reqLogger := logger.LoggerFromContext(ctx).WithFields(zap.String("operation", "ViewPostsById"))

	reqLogger.Info("View Posts By Id")

	posts, err := s.repo.GetPostsByUserId(rows.UserId)
	if err != nil {
		reqLogger.Error("Failed to get posts by id", zap.Error(err))
		return nil, err
	}

	response := &dto.GetPostsResponse{}
	for _, post := range posts {
		response.Posts = append(response.Posts, *post)
	}

	reqLogger.Info("View Posts By Id done")

	return response, nil
}

func (s *PostsService) ViewAllPosts(ctx context.Context) (*dto.GetPostsResponse, error) {
	reqLogger := logger.LoggerFromContext(ctx).WithFields(zap.String("operation", "ViewAllPosts"))

	reqLogger.Info("View All Posts")

	posts, err := s.repo.GetAllPosts()
	if err != nil {
		reqLogger.Error("Failed to get posts by id", zap.Error(err))
		return nil, err
	}

	response := &dto.GetPostsResponse{}
	for _, post := range posts {
		response.Posts = append(response.Posts, *post)
	}

	reqLogger.Info("View All Posts done")

	return response, nil
}

func (s *PostsService) AddImage(ctx context.Context, rows *dto.AddImageToPostRequest) (*dto.AddImageToPostResponse, error) {
	reqLogger := logger.LoggerFromContext(ctx).WithFields(zap.String("operation", "AddImage"))

	reqLogger.Info("Add Image")

	minioCtx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	_, err := s.repo.GetPostById(rows.PostId)
	if err != nil {
		reqLogger.Error("Failed to get post by id", zap.Error(err))
		return nil, errors.ErrPostNotFound
	}

	image, err := s.repo.AddImage(rows.PostId, "not-set", time.Now())
	if err != nil {
		reqLogger.Error("Failed to add image", zap.Error(err))
		return nil, err
	}

	filename := fmt.Sprintf("%s/%s.%s", rows.PostId, image.ImageId, "png")

	_, err = s.minio.Upload(minioCtx, s.bucket, filename, rows.File, rows.Handler.Size)
	if err != nil {
		reqLogger.Error("Failed to upload image", zap.Error(err))
		return nil, err
	}

	url, err := s.minio.GenerateURL(minioCtx, s.bucket, filename, time.Hour*24*7)
	if err != nil {
		reqLogger.Error("Failed to generate url", zap.Error(err))
		return nil, err
	}

	err = s.repo.SetImageURLById(image.ImageId, url)
	if err != nil {
		reqLogger.Error("Failed to set url", zap.Error(err))
		return nil, err
	}

	var message string
	if image != nil {
		message = "Image added successfully"
	}

	response := &dto.AddImageToPostResponse{
		Message: message,
	}

	reqLogger.Info("Add Image done")

	return response, nil
}

func (s *PostsService) DeleteImage(ctx context.Context, rows *dto.DeleteImageFromPostRequest) (*dto.DeleteImageFromPostResponse, error) {
	reqLogger := logger.LoggerFromContext(ctx).WithFields(zap.String("operation", "DeleteImage"))

	reqLogger.Info("Delete Image")

	minioCtx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	_, err := s.repo.GetPostById(rows.PostId)
	if err != nil {
		reqLogger.Error("Failed to get post by id", zap.Error(err))
		return nil, errors.ErrPostOrImageNotFound
	}

	image, err := s.repo.GetImageById(rows.ImageId)
	if err != nil {
		reqLogger.Error("Failed to get image by id", zap.Error(err))
		return nil, errors.ErrPostOrImageNotFound
	}

	filename := fmt.Sprintf("%s/%s.%s", rows.PostId, image.ImageId, "png")

	err = s.minio.DeleteImage(minioCtx, s.bucket, filename)
	if err != nil {
		reqLogger.Error("Failed to delete image", zap.Error(err))
		return nil, err
	}

	err = s.repo.DeleteImageById(image.ImageId)
	if err != nil {
		reqLogger.Error("Failed to delete image", zap.Error(err))
		return nil, err
	}

	var message string
	if image != nil {
		message = "Image deleted successfully"
	}

	response := &dto.DeleteImageFromPostResponse{
		Message: message,
	}

	reqLogger.Info("Delete Image done")

	return response, nil
}

//TODO: месаги сделать с маленькой буквы
