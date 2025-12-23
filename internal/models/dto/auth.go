package dto

type RegistrateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type RegistrateUserResponse struct {
	Message      string `json:"message"`
	AccessToken  string `json:"-"`
	RefreshToken string `json:"-"`
}

type LoginUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginUserResponse struct {
	Message      string `json:"message"`
	AccessToken  string `json:"-"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshUserTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshUserTokenResponse struct {
	Message      string `json:"message"`
	AccessToken  string `json:"-"`
	RefreshToken string `json:"-"`
}
