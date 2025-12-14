package service

import (
	"blog/internal/models"
	"blog/pkg/consts"
	"blog/pkg/consts/errors"
	"time"
)

type PostsBlogRepository interface {
	CreatePost(authorId, idempotencyKey, title, content, status string, createdAt, updatedAt time.Time) (*models.Post, error)
	GetUserById(userId string) (*models.User, error)
	GetPostById(postId string) (*models.Post, error)
	EditPost(postId, authorId, idempotencyKey, title, content, status string, createdAt, updatedAt time.Time) (*models.Post, error)
}

type PostsService struct {
	repo   PostsBlogRepository
	secret string
}

func NewPostsService(repo PostsBlogRepository, secret string) *PostsService {
	return &PostsService{
		repo:   repo,
		secret: secret,
	}
}

func (s *PostsService) CreatePost(post *models.CreatePostRequest) (*models.CreatePostResponse, error) {

	newPost, err := s.repo.CreatePost(post.AuthorId, post.IdempotencyKey, post.Title, post.Content, consts.DraftState, time.Now(), time.Now())
	if err != nil {
		return nil, err
	}
	response := &models.CreatePostResponse{
		PostId: newPost.PostId,
	}
	return response, nil
}

func (s *PostsService) EditPost(raws *models.EditPostRequest) (*models.EditPostResponse, error) {
	post, err := s.repo.GetPostById(raws.PostId)
	if err != nil {
		return nil, err
	}
	if post.AuthorId != raws.AuthorId {
		return nil, errors.ErrInvalidUser
	}
	newPost, err := s.repo.EditPost(post.PostId, post.AuthorId, post.IdempotencyKey, raws.Title, raws.Content, post.Status, post.CreatedAt, time.Now())
	if err != nil {
		return nil, err
	}
	response := &models.EditPostResponse{
		PostId: newPost.PostId,
	}
	return response, nil
}

func (s *PostsService) PublishPost(raws *models.PublishPostRequest) (*models.PublishPostResponse, error) {
	if raws.Status != consts.PublishedState {
		return nil, errors.ErrInvalidPostState
	}
	post, err := s.repo.GetPostById(raws.PostId)
	if err != nil {
		return nil, err
	}
	if post.AuthorId != raws.AuthorId {
		return nil, errors.ErrInvalidUser
	}

	newPost, err := s.repo.EditPost(post.PostId, post.AuthorId, post.IdempotencyKey, post.Title, post.Content, raws.Status, post.CreatedAt, time.Now())
	if err != nil {
		return nil, err
	}
	response := &models.PublishPostResponse{
		PostId: newPost.PostId,
	}
	return response, nil
}
