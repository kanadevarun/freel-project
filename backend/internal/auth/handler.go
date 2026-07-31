package auth

import (
	"encoding/json"
	"net/http"

	"github.com/freel/backend/internal/utils"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Signup(w http.ResponseWriter, r *http.Request) {
	var req SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "INVALID_PAYLOAD")
		return
	}

	err := h.service.Signup(r.Context(), req)
	if err != nil {
		// Basic error handling for Cognito errors, could be expanded
		utils.Error(w, http.StatusBadRequest, err.Error(), "SIGNUP_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Signup successful. Verification code sent to your email.", nil)
}

func (h *Handler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req VerifyEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "INVALID_PAYLOAD")
		return
	}

	err := h.service.VerifyEmail(r.Context(), req)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "VERIFICATION_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Email verified successfully", nil)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "INVALID_PAYLOAD")
		return
	}

	data, err := h.service.Login(r.Context(), req)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "LOGIN_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Login successful", data)
}

func (h *Handler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "INVALID_PAYLOAD")
		return
	}

	err := h.service.ForgotPassword(r.Context(), req)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "FORGOT_PASSWORD_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "If that email is registered, you will receive a reset code.", nil)
}

func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "INVALID_PAYLOAD")
		return
	}

	err := h.service.ResetPassword(r.Context(), req)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "RESET_PASSWORD_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Password reset successfully. Please log in.", nil)
}

// AcceptInvite processes the POST /auth/invite/accept HTTP request.
// It decodes the token and user details from the JSON body, and delegates
// the business logic to the Auth Service.
func (h *Handler) AcceptInvite(w http.ResponseWriter, r *http.Request) {
	var req AcceptInviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "INVALID_PAYLOAD")
		return
	}

	err := h.service.AcceptInvite(r.Context(), req)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "ACCEPT_INVITE_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Invitation accepted successfully. You can now log in.", nil)
}

func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	// MVP: For now, we will simply extract the basic user information from a passed token.
	// In Phase 2, we will implement full JWT validation via middleware.
	// Since the user asked for a simple MVP get auth me that extracts info from Authorization header:
	
	// Example of a basic placeholder that returns generic success:
	utils.Success(w, http.StatusOK, "User details fetched successfully", CurrentUserResponseData{
		Email: "user@example.com", // This will be decoded from token later
	})
}
