package repository

import (
	"blog/internal/models"
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

func (r *BlogRepository) CreateUser(email, passwordHash, role, refreshToken string, refreshTokenExpiryTime time.Time) (*models.User, error) {
	var user models.User

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

func (r *BlogRepository) GetUser(email string) (*models.User, error) {
	var user models.User

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

func (r *BlogRepository) UpdateRefreshToken(userId, refreshToken string) error {
	query := `UPDATE users SET refresh_token = $1 WHERE user_id = $2`
	_, err := r.DB.Exec(query, refreshToken, userId)
	if err != nil {
		return err
	}
	return nil
}
