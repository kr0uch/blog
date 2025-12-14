package repository

import (
	"blog/internal/models/entities"
	"blog/pkg/consts/errors"
	"database/sql"
	stderr "errors"
	"time"

	"github.com/lib/pq"
)

type BlogRepository struct {
	DB *sql.DB
}

func NewBlogRepository(db *sql.DB) *BlogRepository {
	return &BlogRepository{
		DB: db,
	}
}

func (r *BlogRepository) CreateUser(email, passwordHash, role, refreshToken string, refreshTokenExpiryTime time.Time) (*entities.User, error) {
	var user entities.User

	query := `INSERT INTO users (email, password_hash, role, refresh_token, refresh_token_expiry_time) VALUES ($1, $2, $3, $4, $5) RETURNING *`

	err := r.DB.QueryRow(query, email, passwordHash, role, refreshToken, refreshTokenExpiryTime).
		Scan(&user.UserId, &user.Email, &user.PasswordHash, &user.Role, &user.RefreshToken, &user.RefreshTokenExpiryTime)
	if err != nil {
		pgErr, ok := err.(*pq.Error)
		if ok && pgErr.Code == "23505" {
			return nil, errors.ErrUserAlreadyExists
		}
		return nil, err
	}
	return &user, nil
}

func (r *BlogRepository) GetUserByEmail(email string) (*entities.User, error) {
	var user entities.User

	query := `SELECT * FROM users WHERE email = $1`
	err := r.DB.QueryRow(query, email).Scan(&user.UserId, &user.Email, &user.PasswordHash, &user.Role, &user.RefreshToken, &user.RefreshTokenExpiryTime)
	if err != nil {
		if stderr.Is(err, sql.ErrNoRows) {
			return nil, errors.ErrInvalidEmailOrPassword
		}
		return nil, err
	}

	return &user, nil
}

func (r *BlogRepository) GetUserByRefreshToken(refreshToken string) (*entities.User, error) {
	var user entities.User

	query := `SELECT * FROM users WHERE refresh_token = $1`
	err := r.DB.QueryRow(query, refreshToken).
		Scan(&user.UserId, &user.Email, &user.PasswordHash, &user.Role, &user.RefreshToken, &user.RefreshTokenExpiryTime)
	if err != nil {
		if stderr.Is(err, sql.ErrNoRows) {
			return nil, errors.ErrInvalidEmailOrPassword
		}
		return nil, err
	}

	return &user, nil
}

func (r *BlogRepository) GetUserById(userId string) (*entities.User, error) {
	var user entities.User

	query := `SELECT * FROM users WHERE user_id = $1`
	err := r.DB.QueryRow(query, userId).
		Scan(&user.UserId, &user.Email, &user.PasswordHash, &user.Role, &user.RefreshToken, &user.RefreshTokenExpiryTime)
	if err != nil {
		if stderr.Is(err, sql.ErrNoRows) {
			return nil, errors.ErrInvalidAccessToken
		}
		return nil, err
	}
	return &user, nil
}

func (r *BlogRepository) UpdateRefreshToken(userId, refreshToken string) error {
	query := `UPDATE users SET refresh_token = $1 WHERE user_id = $2`
	_, err := r.DB.Exec(query, refreshToken, userId)
	if err != nil {
		return err
	}
	return nil
}

func (r *BlogRepository) CreatePost(authorId, idempotencyKey, title, content, status string, createdAt, updatedAt time.Time) (*entities.Post, error) {
	var post entities.Post

	query := `INSERT INTO posts (author_id, idempotency_key, title, content, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *`
	err := r.DB.QueryRow(query, authorId, idempotencyKey, title, content, status, createdAt, updatedAt).
		Scan(&post.PostId, &post.AuthorId, &post.IdempotencyKey, &post.Title, &post.Content, &post.Status, &post.CreatedAt, &post.UpdatedAt)
	if err != nil {
		pgErr, ok := err.(*pq.Error)
		if ok && pgErr.Code == "23505" {
			return nil, errors.ErrInvalidIdempotencyKey
		}
		return nil, err
	}
	return &post, nil
}

func (r *BlogRepository) GetPostById(postId string) (*entities.Post, error) {
	var post entities.Post

	query := `SELECT * FROM posts WHERE post_id = $1`
	err := r.DB.QueryRow(query, postId).Scan(&post.PostId, &post.AuthorId, &post.IdempotencyKey, &post.Title, &post.Content, &post.Status, &post.CreatedAt, &post.UpdatedAt)
	if err != nil {
		if stderr.Is(err, sql.ErrNoRows) {
			return nil, errors.ErrInvalidPostId
		}
		return nil, err
	}

	return &post, nil
}

func (r *BlogRepository) EditPost(postId, authorId, idempotencyKey, title, content, status string, createdAt, updatedAt time.Time) (*entities.Post, error) {
	var post entities.Post

	query := `UPDATE posts SET author_id = $1, idempotency_key = $2, title = $3, content = $4, status = $5, created_at = $6, updated_at = $7 WHERE post_id = $8 RETURNING *`
	err := r.DB.QueryRow(query, authorId, idempotencyKey, title, content, status, createdAt, updatedAt, postId).
		Scan(&post.PostId, &post.AuthorId, &post.IdempotencyKey, &post.Title, &post.Content, &post.Status, &post.CreatedAt, &post.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *BlogRepository) GetPostsById(userId string) ([]*entities.Post, error) {
	var posts []*entities.Post
	query := `SELECT * FROM posts WHERE author_id = $1`
	rows, err := r.DB.Query(query, userId)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var post entities.Post
		err = rows.Scan(&post.PostId, &post.AuthorId, &post.IdempotencyKey, &post.Title, &post.Content, &post.Status, &post.CreatedAt, &post.UpdatedAt)
		if err != nil {
			return nil, err
		}
		posts = append(posts, &post)
	}
	return posts, nil
}

func (r *BlogRepository) GetAllPosts() ([]*entities.Post, error) {
	var posts []*entities.Post
	query := `SELECT * FROM posts WHERE status = $1`
	rows, err := r.DB.Query(query, "Published")
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var post entities.Post
		err = rows.Scan(&post.PostId, &post.AuthorId, &post.IdempotencyKey, &post.Title, &post.Content, &post.Status, &post.CreatedAt, &post.UpdatedAt)
		if err != nil {
			return nil, err
		}
		posts = append(posts, &post)
	}
	return posts, nil
}

func (r *BlogRepository) AddImage(postId, imageURL string, createdAt time.Time) (*entities.Image, error) {
	var image entities.Image
	query := `INSERT INTO images (post_id, image_url, created_at) VALUES ($1, $2, $3) RETURNING *`
	err := r.DB.QueryRow(query, postId, imageURL, createdAt).Scan(&image.ImageId, &image.PostId, &image.ImageURL, &image.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &image, nil
}

func (r *BlogRepository) SetImageURLById(imageId, URL string) error {
	query := `UPDATE images SET image_url = $1 WHERE image_id = $2`
	_, err := r.DB.Exec(query, URL, imageId)
	if err != nil {
		return err
	}
	return nil
}

func (r *BlogRepository) GetImageById(imageId string) (*entities.Image, error) {
	var image entities.Image
	query := `SELECT * FROM images WHERE image_id = $1`
	row := r.DB.QueryRow(query, imageId)
	err := row.Scan(&image.ImageId, &image.PostId, &image.ImageURL, &image.CreatedAt)
	if err != nil {
		if stderr.Is(err, sql.ErrNoRows) {
			return nil, errors.ErrInvalidImageId
		}
		return nil, err
	}
	return &image, nil
}

func (r *BlogRepository) DeleteImageById(imageId string) error {
	query := `DELETE FROM images WHERE image_id = $1`
	_, err := r.DB.Exec(query, imageId)
	if err != nil {
		return err
	}
	return nil
}
