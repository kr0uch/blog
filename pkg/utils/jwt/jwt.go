package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
)

func NewAccessToken(id string, secret string) string {
	token := jwt.New(jwt.SigningMethodHS512)
	token.Claims = jwt.MapClaims{
		"sub": id,
		"exp": time.Now().Add(time.Hour * 2).Unix(),
	}
	tokenString, _ := token.SignedString([]byte(secret))
	return tokenString
}

func NewRefreshToken(email string, secret string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS512)
	token.Claims = jwt.MapClaims{
		"sub": email,
		"exp": time.Now().Add(time.Hour * 24 * 7).Unix(),
	}
	return token.SignedString([]byte(secret))
}

func ValidateToken(tokenString string, secret string) (*jwt.MapClaims, error) { //TODO: переписать exp не как float а как timestamp
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	if token == nil {
		return nil, fmt.Errorf("token is nil")
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is invalid")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("token claims is invalid")
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return nil, fmt.Errorf("token exp is invalid")
	}
	if time.Now().After(time.Unix(int64(exp), 0)) {
		return nil, fmt.Errorf("token is expired")
	}

	return &claims, nil
}
