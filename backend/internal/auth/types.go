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

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
	Email        string `json:"email,omitempty"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

type ResetPasswordRequest struct {
	Email       string `json:"email"`
	Code        string `json:"code"`
	NewPassword string `json:"new_password"`
}

type RoleResponse struct {
	Name        string   `json:"name"`
	DisplayName string   `json:"display_name"`
	Permissions []string `json:"permissions"`
}

type UserResponse struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	FullName  string `json:"full_name"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type OrgResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type LoginResponseData struct {
	AccessToken  string       `json:"access_token"`
	IDToken      string       `json:"id_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int32        `json:"expires_in"`
	Role         RoleResponse `json:"role"`
	User         UserResponse `json:"user"`
	Org          OrgResponse  `json:"org"`
}

type RefreshResponseData struct {
	AccessToken  string `json:"access_token"`
	IDToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int32  `json:"expires_in"`
}

type CurrentUserResponseData struct {
	User UserResponse `json:"user"`
	Org  OrgResponse  `json:"org"`
	Role RoleResponse `json:"role"`
}
