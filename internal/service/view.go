package service

import "blog/internal/models"

type BlogRepository interface {
	GetPostsById(userId string) ([]*models.Post, error)
	GetAllPosts() ([]*models.Post, error)
}
type ViewService struct {
	repo BlogRepository
}

func NewViewService(repo BlogRepository) *ViewService {
	return &ViewService{
		repo: repo,
	}
}

func (s *ViewService) ViewPostsById(raws *models.GetPostsByIdRequest) (*models.GetPostsResponse, error) {
	posts, err := s.repo.GetPostsById(raws.UserId)
	if err != nil {
		return nil, err
	}

	response := &models.GetPostsResponse{}
	for _, post := range posts {
		response.Posts = append(response.Posts, *post)
	}
	return response, nil
}

func (s *ViewService) ViewAllPosts() (*models.GetPostsResponse, error) {
	posts, err := s.repo.GetAllPosts()
	if err != nil {
		return nil, err
	}

	response := &models.GetPostsResponse{}
	for _, post := range posts {
		response.Posts = append(response.Posts, *post)
	}
	return response, nil
}
