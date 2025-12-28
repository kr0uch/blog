package controllers

import (
	"blog/internal/models/dto"
	"blog/internal/models/entities"
	"blog/pkg/consts"
	"blog/pkg/consts/errors"
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockPostsService struct {
	mock.Mock
}

func (m *MockPostsService) CreatePost(post *dto.CreatePostRequest) (*dto.CreatePostResponse, error) {
	args := m.Called(post)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.CreatePostResponse), args.Error(1)
}

func (m *MockPostsService) EditPost(rows *dto.EditPostRequest) (*dto.EditPostResponse, error) {
	args := m.Called(rows)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.EditPostResponse), args.Error(1)
}

func (m *MockPostsService) PublishPost(post *dto.PublishPostRequest) (*dto.PublishPostResponse, error) {
	args := m.Called(post)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.PublishPostResponse), args.Error(1)
}

func (m *MockPostsService) ViewPostsById(rows *dto.GetPostsByIdRequest) (*dto.GetPostsResponse, error) {
	args := m.Called(rows)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.GetPostsResponse), args.Error(1)
}

func (m *MockPostsService) ViewAllPosts() (*dto.GetPostsResponse, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.GetPostsResponse), args.Error(1)
}

func (m *MockPostsService) AddImage(rows *dto.AddImageToPostRequest) (*dto.AddImageToPostResponse, error) {
	args := m.Called(rows)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.AddImageToPostResponse), args.Error(1)
}

func (m *MockPostsService) DeleteImage(rows *dto.DeleteImageFromPostRequest) (*dto.DeleteImageFromPostResponse, error) {
	args := m.Called(rows)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.DeleteImageFromPostResponse), args.Error(1)
}

func TestPostsController_CreatePost(t *testing.T) {
	tests := []struct {
		name               string
		requestBody        interface{}
		role               string
		key                string
		mockFunc           func(m *MockPostsService)
		expectedStatusCode int
		checkResponseBody  func(t *testing.T, responseBody string)
	}{
		{
			name: "successful",
			requestBody: &dto.CreatePostRequest{
				AuthorId:       "authorId",
				IdempotencyKey: "idempotencyKey",
				Title:          "title",
				Content:        "content",
			},
			role: consts.AuthorRole,
			key:  consts.CtxUserKey,
			mockFunc: func(m *MockPostsService) {
				m.On("CreatePost", mock.AnythingOfType("*dto.CreatePostRequest")).
					Return(&dto.CreatePostResponse{
						Message: "message",
					}, nil)
			},
			expectedStatusCode: http.StatusOK,
			checkResponseBody: func(t *testing.T, responseBody string) {
				var response dto.CreatePostResponse
				err := json.Unmarshal([]byte(responseBody), &response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response.Message)
			},
		},
		{
			name: "no permission",
			requestBody: &dto.CreatePostRequest{
				AuthorId:       "authorId",
				IdempotencyKey: "idempotencyKey",
				Title:          "title",
				Content:        "content",
			},
			role:               consts.ReaderRole,
			key:                consts.CtxUserKey,
			expectedStatusCode: http.StatusForbidden,
		},
		{
			name: "failed to get user",
			requestBody: &dto.CreatePostRequest{
				AuthorId:       "authorId",
				IdempotencyKey: "idempotencyKey",
				Title:          "title",
				Content:        "content",
			},
			role:               consts.AuthorRole,
			key:                "testKey",
			expectedStatusCode: http.StatusForbidden,
		},
		{
			name:               "incorrect data",
			requestBody:        nil,
			role:               consts.AuthorRole,
			key:                consts.CtxUserKey,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "invalid idempotency key",
			requestBody: &dto.CreatePostRequest{
				AuthorId:       "authorId",
				IdempotencyKey: "idempotencyKey",
				Title:          "title",
				Content:        "content",
			},
			role: consts.AuthorRole,
			key:  consts.CtxUserKey,
			mockFunc: func(m *MockPostsService) {
				m.On("CreatePost", mock.AnythingOfType("*dto.CreatePostRequest")).
					Return(nil, errors.ErrInvalidIdempotencyKey)
			},
			expectedStatusCode: http.StatusConflict,
		},
		{
			name: "internal server error",
			requestBody: &dto.CreatePostRequest{
				AuthorId:       "authorId",
				IdempotencyKey: "idempotencyKey",
				Title:          "title",
				Content:        "content",
			},
			role: consts.AuthorRole,
			key:  consts.CtxUserKey,
			mockFunc: func(m *MockPostsService) {
				m.On("CreatePost", mock.AnythingOfType("*dto.CreatePostRequest")).
					Return(nil, errors.ErrInternalServerError)
			},
			expectedStatusCode: http.StatusForbidden,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockPostsService := &MockPostsService{}
			if test.mockFunc != nil {
				test.mockFunc(mockPostsService)
			}

			controller := NewPostsController(mockPostsService)

			req := &http.Request{}

			if test.requestBody != nil {
				body, _ := json.Marshal(test.requestBody)
				req = httptest.NewRequest(http.MethodPost, "/api/posts", bytes.NewBuffer(body))
			} else {
				req = httptest.NewRequest(http.MethodPost, "/api/posts", nil)
			}
			req.Header.Set("Content-Type", "application/json")

			ctx := context.WithValue(req.Context(), test.key, &entities.User{
				Role: test.role,
			})

			rr := httptest.NewRecorder()
			controller.CreatePost(rr, req.WithContext(ctx))

			assert.Equal(t, test.expectedStatusCode, rr.Code)

			if test.checkResponseBody != nil {
				test.checkResponseBody(t, rr.Body.String())
			}

			mockPostsService.AssertExpectations(t)
		})
	}
}

func TestPostsController_EditPost(t *testing.T) {
	postId := uuid.New().String()

	tests := []struct {
		name               string
		requestBody        interface{}
		role               string
		key                string
		mockFunc           func(m *MockPostsService)
		expectedStatusCode int
		checkResponseBody  func(t *testing.T, responseBody string)
	}{
		{
			name: "successful",
			requestBody: &dto.EditPostRequest{
				AuthorId: "authorId",
				PostId:   postId,
				Title:    "title",
				Content:  "content",
			},
			role: consts.AuthorRole,
			key:  consts.CtxUserKey,
			mockFunc: func(m *MockPostsService) {
				m.On("EditPost", mock.AnythingOfType("*dto.EditPostRequest")).
					Return(&dto.EditPostResponse{
						Message: "message",
					}, nil)
			},
			expectedStatusCode: http.StatusOK,
			checkResponseBody: func(t *testing.T, responseBody string) {
				var response dto.EditPostResponse
				err := json.Unmarshal([]byte(responseBody), &response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response.Message)
			},
		},
		{
			name: "no permission",
			requestBody: &dto.EditPostRequest{
				PostId:   postId,
				AuthorId: "authorId",
				Title:    "title",
				Content:  "content",
			},
			role: consts.AuthorRole,
			key:  consts.CtxUserKey,
			mockFunc: func(m *MockPostsService) {
				m.On("EditPost", mock.AnythingOfType("*dto.EditPostRequest")).
					Return(nil, errors.ErrNoPermission)
			},
			expectedStatusCode: http.StatusForbidden,
		},
		{
			name: "failed to get user",
			requestBody: &dto.EditPostRequest{
				PostId:   postId,
				AuthorId: "authorId",
				Title:    "title",
				Content:  "content",
			},
			role:               consts.AuthorRole,
			key:                "testKey",
			expectedStatusCode: http.StatusForbidden,
		},
		{
			name:               "incorrect data",
			requestBody:        nil,
			role:               consts.AuthorRole,
			key:                consts.CtxUserKey,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "post not found",
			requestBody: &dto.EditPostRequest{
				PostId:   "testUuid",
				AuthorId: "authorId",
				Title:    "title",
				Content:  "content",
			},
			role: consts.AuthorRole,
			key:  consts.CtxUserKey,
			mockFunc: func(m *MockPostsService) {
				m.On("EditPost", mock.AnythingOfType("*dto.EditPostRequest")).
					Return(nil, errors.ErrPostNotFound)
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name: "invalid user",
			requestBody: &dto.EditPostRequest{
				PostId:   postId,
				AuthorId: "authorId",
				Title:    "title",
				Content:  "content",
			},
			role: consts.AuthorRole,
			key:  consts.CtxUserKey,
			mockFunc: func(m *MockPostsService) {
				m.On("EditPost", mock.AnythingOfType("*dto.EditPostRequest")).
					Return(nil, errors.ErrInvalidUser)
			},
			expectedStatusCode: http.StatusForbidden,
		},
		{
			name: "invalid post postId",
			requestBody: &dto.EditPostRequest{
				PostId:   "",
				AuthorId: "authorId",
				Title:    "title",
				Content:  "content",
			},
			role: consts.AuthorRole,
			key:  consts.CtxUserKey,
			mockFunc: func(m *MockPostsService) {
				m.On("EditPost", mock.AnythingOfType("*dto.EditPostRequest")).
					Return(nil, errors.ErrInvalidPostId)
			},
			expectedStatusCode: http.StatusForbidden,
		},
		{
			name: "internal server error",
			requestBody: &dto.EditPostRequest{
				PostId:   postId,
				AuthorId: "authorId",
				Title:    "title",
				Content:  "content",
			},
			role: consts.AuthorRole,
			key:  consts.CtxUserKey,
			mockFunc: func(m *MockPostsService) {
				m.On("EditPost", mock.AnythingOfType("*dto.EditPostRequest")).
					Return(nil, errors.ErrInternalServerError)
			},
			expectedStatusCode: http.StatusForbidden,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockPostsService := &MockPostsService{}
			if test.mockFunc != nil {
				test.mockFunc(mockPostsService)
			}

			controller := NewPostsController(mockPostsService)

			req := &http.Request{}

			if test.requestBody != nil {
				body, _ := json.Marshal(test.requestBody)
				req = httptest.NewRequest(http.MethodPut, "/api/posts/"+postId, bytes.NewBuffer(body))
			} else {
				req = httptest.NewRequest(http.MethodPut, "/api/posts/"+postId, nil)
			}
			req.Header.Set("Content-Type", "application/json")

			ctx := context.WithValue(req.Context(), test.key, &entities.User{
				Role: test.role,
			})

			rr := httptest.NewRecorder()
			controller.EditPost(rr, req.WithContext(ctx))

			assert.Equal(t, test.expectedStatusCode, rr.Code)

			if test.checkResponseBody != nil {
				test.checkResponseBody(t, rr.Body.String())
			}

			mockPostsService.AssertExpectations(t)
		})
	}
}

func TestPostsController_PublishPost(t *testing.T) {
	postId := uuid.New().String()

	tests := []struct {
		name               string
		requestBody        interface{}
		role               string
		key                string
		mockFunc           func(m *MockPostsService)
		expectedStatusCode int
		checkResponseBody  func(t *testing.T, responseBody string)
	}{
		{
			name: "successful",
			requestBody: &dto.PublishPostRequest{
				PostId:   postId,
				AuthorId: "authorId",
				Status:   consts.PublishedState,
			},
			role: consts.AuthorRole,
			key:  consts.CtxUserKey,
			mockFunc: func(m *MockPostsService) {
				m.On("PublishPost", mock.AnythingOfType("*dto.PublishPostRequest")).
					Return(&dto.PublishPostResponse{
						Message: "message",
					}, nil)
			},
			expectedStatusCode: http.StatusOK,
			checkResponseBody: func(t *testing.T, responseBody string) {
				var response dto.PublishPostResponse
				err := json.Unmarshal([]byte(responseBody), &response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response.Message)
			},
		},
		{
			name: "failed to get user",
			requestBody: &dto.PublishPostRequest{
				PostId:   postId,
				AuthorId: "authorId",
				Status:   consts.PublishedState,
			},
			role:               consts.AuthorRole,
			key:                "testKey",
			expectedStatusCode: http.StatusForbidden,
		},
		{
			name: "no permission",
			requestBody: &dto.PublishPostRequest{
				PostId:   postId,
				AuthorId: "authorId",
				Status:   consts.PublishedState,
			},
			role:               consts.ReaderRole,
			key:                consts.CtxUserKey,
			expectedStatusCode: http.StatusForbidden,
		},
		{
			name:               "incorrect data",
			requestBody:        nil,
			role:               consts.AuthorRole,
			key:                consts.CtxUserKey,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "invalid post status",
			requestBody: &dto.PublishPostRequest{
				PostId:   postId,
				AuthorId: "authorId",
				Status:   "testStatus",
			},
			role: consts.AuthorRole,
			key:  consts.CtxUserKey,
			mockFunc: func(m *MockPostsService) {
				m.On("PublishPost", mock.AnythingOfType("*dto.PublishPostRequest")).
					Return(nil, errors.ErrInvalidPostStatus)
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "post not found",
			requestBody: &dto.PublishPostRequest{
				PostId:   postId,
				AuthorId: "authorId",
				Status:   consts.PublishedState,
			},
			role: consts.AuthorRole,
			key:  consts.CtxUserKey,
			mockFunc: func(m *MockPostsService) {
				m.On("PublishPost", mock.AnythingOfType("*dto.PublishPostRequest")).
					Return(nil, errors.ErrPostNotFound)
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name: "invalid user",
			requestBody: &dto.PublishPostRequest{
				PostId:   postId,
				AuthorId: "authorId",
				Status:   consts.PublishedState,
			},
			role: consts.AuthorRole,
			key:  consts.CtxUserKey,
			mockFunc: func(m *MockPostsService) {
				m.On("PublishPost", mock.AnythingOfType("*dto.PublishPostRequest")).
					Return(nil, errors.ErrInvalidUser)
			},
			expectedStatusCode: http.StatusForbidden,
		},
		{
			name: "invalid post postId",
			requestBody: &dto.PublishPostRequest{
				PostId:   "",
				AuthorId: "authorId",
				Status:   consts.PublishedState,
			},
			role: consts.AuthorRole,
			key:  consts.CtxUserKey,
			mockFunc: func(m *MockPostsService) {
				m.On("PublishPost", mock.AnythingOfType("*dto.PublishPostRequest")).
					Return(nil, errors.ErrInvalidPostId)
			},
			expectedStatusCode: http.StatusForbidden,
		},
		{
			name: "internal server error",
			requestBody: &dto.PublishPostRequest{
				PostId:   postId,
				AuthorId: "authorId",
				Status:   consts.PublishedState,
			},
			role: consts.AuthorRole,
			key:  consts.CtxUserKey,
			mockFunc: func(m *MockPostsService) {
				m.On("PublishPost", mock.AnythingOfType("*dto.PublishPostRequest")).
					Return(nil, errors.ErrInternalServerError)
			},
			expectedStatusCode: http.StatusForbidden,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockPostsService := &MockPostsService{}
			if test.mockFunc != nil {
				test.mockFunc(mockPostsService)
			}

			controller := NewPostsController(mockPostsService)

			req := &http.Request{}

			if test.requestBody != nil {
				body, _ := json.Marshal(test.requestBody)
				req = httptest.NewRequest(http.MethodPatch, "/api/posts/"+postId+"/status", bytes.NewBuffer(body))
			} else {
				req = httptest.NewRequest(http.MethodPatch, "/api/posts/"+postId+"/status", nil)
			}
			req.Header.Set("Content-Type", "application/json")

			ctx := context.WithValue(req.Context(), test.key, &entities.User{
				Role: test.role,
			})

			rr := httptest.NewRecorder()
			controller.PublishPost(rr, req.WithContext(ctx))

			assert.Equal(t, test.expectedStatusCode, rr.Code)

			if test.checkResponseBody != nil {
				test.checkResponseBody(t, rr.Body.String())
			}

			mockPostsService.AssertExpectations(t)
		})
	}
}

func TestPostsController_AddImageToPost(t *testing.T) {
	postId := uuid.New().String()

	tests := []struct {
		name               string
		requestBody        interface{}
		role               string
		key                string
		imageKey           string
		mockFunc           func(m *MockPostsService)
		expectedStatusCode int
		checkResponseBody  func(t *testing.T, responseBody string)
	}{
		{
			name: "successful",
			requestBody: &dto.AddImageToPostRequest{
				PostId:   postId,
				AuthorId: "authorId",
			},
			role:     consts.AuthorRole,
			key:      consts.CtxUserKey,
			imageKey: "image",
			mockFunc: func(m *MockPostsService) {
				m.On("AddImage", mock.AnythingOfType("*dto.AddImageToPostRequest")).
					Return(&dto.AddImageToPostResponse{
						Message: "message",
					}, nil)
			},
			expectedStatusCode: http.StatusOK,
			checkResponseBody: func(t *testing.T, responseBody string) {
				var response dto.AddImageToPostResponse
				err := json.Unmarshal([]byte(responseBody), &response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response.Message)
			},
		},
		{
			name: "incorrect data",
			requestBody: &dto.AddImageToPostRequest{
				PostId:   postId,
				AuthorId: "authorId",
			},
			role:               consts.AuthorRole,
			key:                consts.CtxUserKey,
			imageKey:           "testImageKey",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "no permission",
			requestBody: &dto.AddImageToPostRequest{
				PostId:   postId,
				AuthorId: "authorId",
			},
			role:               consts.ReaderRole,
			key:                consts.CtxUserKey,
			imageKey:           "image",
			expectedStatusCode: http.StatusForbidden,
		},
		{
			name: "failed to get user",
			requestBody: &dto.AddImageToPostRequest{
				PostId:   postId,
				AuthorId: "authorId",
			},
			role:               consts.AuthorRole,
			key:                "testKey",
			imageKey:           "image",
			expectedStatusCode: http.StatusForbidden,
		},
		{
			name: "post not found",
			requestBody: &dto.AddImageToPostRequest{
				PostId:   postId,
				AuthorId: "authorId",
			},
			role:     consts.AuthorRole,
			key:      consts.CtxUserKey,
			imageKey: "image",
			mockFunc: func(m *MockPostsService) {
				m.On("AddImage", mock.AnythingOfType("*dto.AddImageToPostRequest")).
					Return(nil, errors.ErrPostNotFound)
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name: "minio bucket not exists",
			requestBody: &dto.AddImageToPostRequest{
				PostId:   postId,
				AuthorId: "authorId",
			},
			role:     consts.AuthorRole,
			key:      consts.CtxUserKey,
			imageKey: "image",
			mockFunc: func(m *MockPostsService) {
				m.On("AddImage", mock.AnythingOfType("*dto.AddImageToPostRequest")).
					Return(nil, errors.ErrMinioBucketNotExists)
			},
			expectedStatusCode: http.StatusForbidden,
		},
		{
			name: "minio cant put object",
			requestBody: &dto.AddImageToPostRequest{
				PostId:   postId,
				AuthorId: "authorId",
			},
			role:     consts.AuthorRole,
			key:      consts.CtxUserKey,
			imageKey: "image",
			mockFunc: func(m *MockPostsService) {
				m.On("AddImage", mock.AnythingOfType("*dto.AddImageToPostRequest")).
					Return(nil, errors.ErrMinioPutObject)
			},
			expectedStatusCode: http.StatusForbidden,
		},
		{
			name: "minio cant presigned get object",
			requestBody: &dto.AddImageToPostRequest{
				PostId:   postId,
				AuthorId: "authorId",
			},
			role:     consts.AuthorRole,
			key:      consts.CtxUserKey,
			imageKey: "image",
			mockFunc: func(m *MockPostsService) {
				m.On("AddImage", mock.AnythingOfType("*dto.AddImageToPostRequest")).
					Return(nil, errors.ErrMinioPresignedGetObject)
			},
			expectedStatusCode: http.StatusForbidden,
		},
		{
			name: "invalid post postId",
			requestBody: &dto.AddImageToPostRequest{
				PostId:   "",
				AuthorId: "authorId",
			},
			role:     consts.AuthorRole,
			key:      consts.CtxUserKey,
			imageKey: "image",
			mockFunc: func(m *MockPostsService) {
				m.On("AddImage", mock.AnythingOfType("*dto.AddImageToPostRequest")).
					Return(nil, errors.ErrInvalidPostId)
			},
			expectedStatusCode: http.StatusForbidden,
		},
		{
			name: "internal server error",
			requestBody: &dto.AddImageToPostRequest{
				PostId:   postId,
				AuthorId: "authorId",
			},
			role:     consts.AuthorRole,
			key:      consts.CtxUserKey,
			imageKey: "image",
			mockFunc: func(m *MockPostsService) {
				m.On("AddImage", mock.AnythingOfType("*dto.AddImageToPostRequest")).
					Return(nil, errors.ErrInternalServerError)
			},
			expectedStatusCode: http.StatusForbidden,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockPostsService := &MockPostsService{}
			if test.mockFunc != nil {
				test.mockFunc(mockPostsService)
			}

			controller := NewPostsController(mockPostsService)

			req := &http.Request{}

			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)
			fileContent := []byte("test content")
			part, err := writer.CreateFormFile(test.imageKey, "test.jpg")
			if err != nil {
				t.Fatal(err)
			}
			part.Write(fileContent)

			writer.Close()

			req = httptest.NewRequest(http.MethodPost, "/api/posts/"+postId+"/images", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())

			ctx := context.WithValue(req.Context(), test.key, &entities.User{
				Role: test.role,
			})

			rr := httptest.NewRecorder()
			controller.AddImageToPost(rr, req.WithContext(ctx))

			assert.Equal(t, test.expectedStatusCode, rr.Code)

			if test.checkResponseBody != nil {
				test.checkResponseBody(t, rr.Body.String())
			}

			mockPostsService.AssertExpectations(t)
		})
	}
}

func TestPostsController_DeleteImageFromPost(t *testing.T) {
	postId := uuid.New().String()
	imageId := uuid.New().String()

	tests := []struct {
		name               string
		role               string
		key                string
		mockFunc           func(m *MockPostsService)
		expectedStatusCode int
		checkResponseBody  func(t *testing.T, responseBody string)
	}{
		{
			name: "successful",
			role: consts.AuthorRole,
			key:  consts.CtxUserKey,
			mockFunc: func(m *MockPostsService) {
				m.On("DeleteImage", mock.AnythingOfType("*dto.DeleteImageFromPostRequest")).
					Return(&dto.DeleteImageFromPostResponse{
						Message: "message",
					}, nil)
			},
			expectedStatusCode: http.StatusOK,
			checkResponseBody: func(t *testing.T, responseBody string) {
				var response dto.DeleteImageFromPostResponse
				err := json.Unmarshal([]byte(responseBody), &response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response.Message)
			},
		},
		{
			name:               "failed to get user",
			role:               consts.AuthorRole,
			key:                "testKey",
			expectedStatusCode: http.StatusForbidden,
		},
		{
			name:               "no permission",
			role:               consts.ReaderRole,
			key:                consts.CtxUserKey,
			expectedStatusCode: http.StatusForbidden,
		},
		{
			name: "post or image not found",
			role: consts.AuthorRole,
			key:  consts.CtxUserKey,
			mockFunc: func(m *MockPostsService) {
				m.On("DeleteImage", mock.AnythingOfType("*dto.DeleteImageFromPostRequest")).
					Return(nil, errors.ErrPostOrImageNotFound)
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name: "minio bucket not exists",
			role: consts.AuthorRole,
			key:  consts.CtxUserKey,
			mockFunc: func(m *MockPostsService) {
				m.On("DeleteImage", mock.AnythingOfType("*dto.DeleteImageFromPostRequest")).
					Return(nil, errors.ErrMinioBucketNotExists)
			},
			expectedStatusCode: http.StatusForbidden,
		},
		{
			name: "minio cant get object",
			role: consts.AuthorRole,
			key:  consts.CtxUserKey,
			mockFunc: func(m *MockPostsService) {
				m.On("DeleteImage", mock.AnythingOfType("*dto.DeleteImageFromPostRequest")).
					Return(nil, errors.ErrMinioGetObject)
			},
			expectedStatusCode: http.StatusForbidden,
		},
		{
			name: "invalid image id",
			role: consts.AuthorRole,
			key:  consts.CtxUserKey,
			mockFunc: func(m *MockPostsService) {
				m.On("DeleteImage", mock.AnythingOfType("*dto.DeleteImageFromPostRequest")).
					Return(nil, errors.ErrInvalidImageId)
			},
			expectedStatusCode: http.StatusForbidden,
		},
		{
			name: "minio cant remove object",
			role: consts.AuthorRole,
			key:  consts.CtxUserKey,
			mockFunc: func(m *MockPostsService) {
				m.On("DeleteImage", mock.AnythingOfType("*dto.DeleteImageFromPostRequest")).
					Return(nil, errors.ErrMinioRemoveObject)
			},
			expectedStatusCode: http.StatusForbidden,
		},
		{
			name: "internal server error",
			role: consts.AuthorRole,
			key:  consts.CtxUserKey,
			mockFunc: func(m *MockPostsService) {
				m.On("DeleteImage", mock.AnythingOfType("*dto.DeleteImageFromPostRequest")).
					Return(nil, errors.ErrInternalServerError)
			},
			expectedStatusCode: http.StatusForbidden,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockPostsService := &MockPostsService{}
			if test.mockFunc != nil {
				test.mockFunc(mockPostsService)
			}

			controller := NewPostsController(mockPostsService)

			req := httptest.NewRequest(http.MethodDelete, "/api/posts/"+postId+"/images/"+imageId, nil)

			req.Header.Set("Content-Type", "application/json")

			ctx := context.WithValue(req.Context(), test.key, &entities.User{
				Role: test.role,
			})

			rr := httptest.NewRecorder()
			controller.DeleteImageFromPost(rr, req.WithContext(ctx))

			assert.Equal(t, test.expectedStatusCode, rr.Code)

			if test.checkResponseBody != nil {
				test.checkResponseBody(t, rr.Body.String())
			}

			mockPostsService.AssertExpectations(t)
		})
	}
}

func TestPostsController_ViewPosts(t *testing.T) {
	tests := []struct {
		name               string
		role               string
		key                string
		mockFunc           func(m *MockPostsService)
		expectedStatusCode int
		checkResponseBody  func(t *testing.T, responseBody string)
	}{
		{
			name: "successful author view",
			role: consts.AuthorRole,
			key:  consts.CtxUserKey,
			mockFunc: func(m *MockPostsService) {
				m.On("ViewPostsById", mock.AnythingOfType("*dto.GetPostsByIdRequest")).
					Return(&dto.GetPostsResponse{
						Posts: make([]entities.Post, 1),
					}, nil)
			},
			expectedStatusCode: http.StatusOK,
			checkResponseBody: func(t *testing.T, responseBody string) {
				var response dto.GetPostsResponse
				err := json.Unmarshal([]byte(responseBody), &response)
				assert.NoError(t, err)
				assert.Equal(t, 1, len(response.Posts))
			},
		},
		{
			name: "successful reader view",
			role: consts.ReaderRole,
			key:  consts.CtxUserKey,
			mockFunc: func(m *MockPostsService) {
				m.On("ViewAllPosts").
					Return(&dto.GetPostsResponse{
						Posts: make([]entities.Post, 1),
					}, nil)
			},
			expectedStatusCode: http.StatusOK,
			checkResponseBody: func(t *testing.T, responseBody string) {
				var response dto.GetPostsResponse
				err := json.Unmarshal([]byte(responseBody), &response)
				assert.NoError(t, err)
				assert.Equal(t, 1, len(response.Posts))
			},
		},
		{
			name:               "failed to get user",
			role:               consts.AuthorRole,
			key:                "testKey",
			expectedStatusCode: http.StatusForbidden,
		},
		{
			name:               "no permission",
			role:               "testRole",
			key:                consts.CtxUserKey,
			expectedStatusCode: http.StatusForbidden,
		},
		{
			name: "invalid post id",
			role: consts.AuthorRole,
			key:  consts.CtxUserKey,
			mockFunc: func(m *MockPostsService) {
				m.On("ViewPostsById", mock.AnythingOfType("*dto.GetPostsByIdRequest")).
					Return(nil, errors.ErrInvalidPostId)
			},
			expectedStatusCode: http.StatusForbidden,
		},
		{
			name: "internal server error",
			role: consts.AuthorRole,
			key:  consts.CtxUserKey,
			mockFunc: func(m *MockPostsService) {
				m.On("ViewPostsById", mock.AnythingOfType("*dto.GetPostsByIdRequest")).
					Return(nil, errors.ErrInternalServerError)
			},
			expectedStatusCode: http.StatusForbidden,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockPostsService := &MockPostsService{}
			if test.mockFunc != nil {
				test.mockFunc(mockPostsService)
			}

			controller := NewPostsController(mockPostsService)

			req := httptest.NewRequest(http.MethodGet, "/api/posts", nil)

			req.Header.Set("Content-Type", "application/json")

			ctx := context.WithValue(req.Context(), test.key, &entities.User{
				Role: test.role,
			})

			rr := httptest.NewRecorder()
			controller.ViewPosts(rr, req.WithContext(ctx))

			assert.Equal(t, test.expectedStatusCode, rr.Code)

			if test.checkResponseBody != nil {
				test.checkResponseBody(t, rr.Body.String())
			}

			mockPostsService.AssertExpectations(t)
		})
	}
}
