package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/freel/backend/internal/audit"
	"github.com/freel/backend/internal/audit/domain"
	"github.com/freel/backend/internal/middleware"
	"github.com/freel/backend/internal/utils"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func getClientIP(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.Header.Get("X-Real-IP")
	}
	if ip == "" {
		ip = r.RemoteAddr
	}
	parts := strings.Split(ip, ",")
	return strings.TrimSpace(parts[0])
}

func (h *Handler) Signup(w http.ResponseWriter, r *http.Request) {
	var req SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "INVALID_PAYLOAD")
		return
	}

	err := h.service.Signup(r.Context(), req)
	if err != nil {
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

	clientIP := getClientIP(r)
	data, err := h.service.Login(r.Context(), req)
	if err != nil {
		// Log failed login attempt
		_, _ = audit.Record(r.Context(), domain.CreateAuditLogParams{
			OrgID:        1,
			ActorType:    domain.ActorTypeUser,
			ActorName:    req.Email,
			Action:       domain.ActionLoginFailed,
			Module:       domain.ModuleAuthentication,
			ResourceType: "USER",
			ResourceID:   req.Email,
			ResourceName: req.Email,
			Description:  fmt.Sprintf("Login attempt failed for %s", req.Email),
			Result:       domain.ResultFailed,
			ErrorMessage: err.Error(),
			IPAddress:    clientIP,
			UserAgent:    r.UserAgent(),
		})

		utils.Error(w, http.StatusUnauthorized, err.Error(), "LOGIN_FAILED")
		return
	}

	// Log successful login
	actorID := data.User.ID
	orgID := data.Org.ID
	if orgID <= 0 {
		orgID = 1
	}

	_, _ = audit.Record(r.Context(), domain.CreateAuditLogParams{
		OrgID:        orgID,
		ActorID:      &actorID,
		ActorType:    domain.ActorTypeUser,
		ActorName:    data.User.FullName,
		ActorRole:    data.Role.DisplayName,
		Action:       domain.ActionLogin,
		Module:       domain.ModuleAuthentication,
		ResourceType: "USER",
		ResourceID:   fmt.Sprintf("%d", data.User.ID),
		ResourceName: data.User.Email,
		Description:  fmt.Sprintf("%s logged in", data.User.FullName),
		Result:       domain.ResultSuccess,
		IPAddress:    clientIP,
		UserAgent:    r.UserAgent(),
	})

	utils.Success(w, http.StatusOK, "Login successful", data)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if userCtx, ok := middleware.GetUserContext(r.Context()); ok && userCtx.UserID > 0 {
		actorID := userCtx.UserID
		_, _ = audit.Record(r.Context(), domain.CreateAuditLogParams{
			OrgID:        userCtx.OrgID,
			ActorID:      &actorID,
			ActorType:    domain.ActorTypeUser,
			ActorRole:    userCtx.Role,
			Action:       domain.ActionLogout,
			Module:       domain.ModuleAuthentication,
			ResourceType: "USER",
			ResourceID:   fmt.Sprintf("%d", userCtx.UserID),
			Description:  "User logged out",
			Result:       domain.ResultSuccess,
			IPAddress:    getClientIP(r),
			UserAgent:    r.UserAgent(),
		})
	}
	utils.Success(w, http.StatusOK, "Logged out successfully", nil)
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request payload", "INVALID_PAYLOAD")
		return
	}

	data, err := h.service.Refresh(r.Context(), req)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, err.Error(), "REFRESH_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Token refreshed successfully", data)
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
	userCtx, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserContext)
	if !ok || userCtx.UserID == 0 {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	data, err := h.service.GetMe(r.Context(), userCtx.UserID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "FETCH_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "User details fetched successfully", data)
}

func (h *Handler) ValidateInvite(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		utils.Error(w, http.StatusBadRequest, "Token is required", "INVALID_PAYLOAD")
		return
	}

	data, err := h.service.ValidateInvite(r.Context(), token)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "VALIDATE_INVITE_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Invitation valid", data)
}

