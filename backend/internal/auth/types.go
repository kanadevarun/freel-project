package auth

type SignupRequest struct {
	FullName    string `json:"full_name"`
	CompanyName string `json:"company_name"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	Role        string `json:"role"`
}

type VerifyEmailRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

type ResetPasswordRequest struct {
	Email       string `json:"email"`
	Code        string `json:"code"`
	NewPassword string `json:"new_password"`
}

type LoginResponseData struct {
	AccessToken  string `json:"access_token"`
	IDToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int32  `json:"expires_in"`
}

type CurrentUserResponseData struct {
	Email string `json:"email"`
}
