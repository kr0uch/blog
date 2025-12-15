package service

import (
	"blog/internal/models/dto"
	"blog/internal/models/entities"
	"blog/pkg/consts"
	"blog/pkg/consts/errors"
	"context"
	"fmt"
	"io"
	"time"
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

func (s *PostsService) CreatePost(post *dto.CreatePostRequest) (*dto.CreatePostResponse, error) {
	newPost, err := s.repo.CreatePost(post.AuthorId, post.IdempotencyKey, post.Title, post.Content, consts.DraftState, time.Now(), time.Now())
	if err != nil {
		return nil, err
	}

	var message string
	if newPost != nil {
		message = "Post created successfully"
	}

	response := &dto.CreatePostResponse{
		Message: message,
	}
	return response, nil
}

func (s *PostsService) EditPost(rows *dto.EditPostRequest) (*dto.EditPostResponse, error) {
	post, err := s.repo.GetPostById(rows.PostId)
	if err != nil {
		return nil, errors.ErrPostNotFound
	}
	if post.AuthorId != rows.AuthorId {
		return nil, errors.ErrInvalidUser
	}

	newPost, err := s.repo.EditPost(post.PostId, post.AuthorId, post.IdempotencyKey, rows.Title, rows.Content, post.Status, post.CreatedAt, time.Now())
	if err != nil {
		return nil, err
	}

	var message string
	if newPost != nil {
		message = "Post edited successfully"
	}

	response := &dto.EditPostResponse{
		Message: message,
	}
	return response, nil
}

func (s *PostsService) PublishPost(rows *dto.PublishPostRequest) (*dto.PublishPostResponse, error) {
	if rows.Status != consts.PublishedState {
		return nil, errors.ErrInvalidPostStatus
	}

	post, err := s.repo.GetPostById(rows.PostId)
	if err != nil {
		return nil, errors.ErrPostNotFound
	}
	if post.AuthorId != rows.AuthorId {
		return nil, errors.ErrInvalidUser
	}

	newPost, err := s.repo.EditPost(post.PostId, post.AuthorId, post.IdempotencyKey, post.Title, post.Content, rows.Status, post.CreatedAt, time.Now())
	if err != nil {
		return nil, err
	}

	var message string
	if newPost != nil {
		message = "Post published successfully"
	}

	response := &dto.PublishPostResponse{
		Message: message,
	}
	return response, nil
}

func (s *PostsService) ViewPostsById(rows *dto.GetPostsByIdRequest) (*dto.GetPostsResponse, error) {
	posts, err := s.repo.GetPostsByUserId(rows.UserId)
	if err != nil {
		return nil, err
	}

	response := &dto.GetPostsResponse{}
	for _, post := range posts {
		response.Posts = append(response.Posts, *post)
	}
	return response, nil
}

func (s *PostsService) ViewAllPosts() (*dto.GetPostsResponse, error) {
	posts, err := s.repo.GetAllPosts()
	if err != nil {
		return nil, err
	}

	response := &dto.GetPostsResponse{}
	for _, post := range posts {
		response.Posts = append(response.Posts, *post)
	}
	return response, nil
}

func (s *PostsService) AddImage(rows *dto.AddImageToPostRequest) (*dto.AddImageToPostResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	_, err := s.repo.GetPostById(rows.PostId)
	if err != nil {
		return nil, errors.ErrPostNotFound
	}

	image, err := s.repo.AddImage(rows.PostId, "not-set", time.Now())
	if err != nil {
		return nil, err
	}

	filename := fmt.Sprintf("%s/%s.%s", rows.PostId, image.ImageId, "png")

	_, err = s.minio.Upload(ctx, s.bucket, filename, rows.File, rows.Handler.Size)
	if err != nil {
		return nil, err
	}

	url, err := s.minio.GenerateURL(ctx, s.bucket, filename, time.Hour*24*7)
	if err != nil {
		return nil, err
	}

	err = s.repo.SetImageURLById(image.ImageId, url)
	if err != nil {
		return nil, err
	}

	var message string
	if image != nil {
		message = "Image added successfully"
	}

	response := &dto.AddImageToPostResponse{
		Message: message,
	}
	return response, nil
}

func (s *PostsService) DeleteImage(rows *dto.DeleteImageFromPostRequest) (*dto.DeleteImageFromPostResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	_, err := s.repo.GetPostById(rows.PostId)
	if err != nil {
		return nil, errors.ErrPostOrImageNotFound
	}

	image, err := s.repo.GetImageById(rows.ImageId)
	if err != nil {
		return nil, errors.ErrPostOrImageNotFound
	}

	filename := fmt.Sprintf("%s/%s.%s", rows.PostId, image.ImageId, "png")

	err = s.minio.DeleteImage(ctx, s.bucket, filename)
	if err != nil {
		return nil, err
	}

	err = s.repo.DeleteImageById(image.ImageId)
	if err != nil {
		return nil, err
	}

	var message string
	if image != nil {
		message = "Image deleted successfully"
	}

	response := &dto.DeleteImageFromPostResponse{
		Message: message,
	}
	return response, nil
}
