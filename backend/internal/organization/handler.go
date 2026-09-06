package organization

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/freel/backend/internal/middleware"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateOrgRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	org, err := h.svc.CreateOrganization(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(org)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid organization ID", http.StatusBadRequest)
		return
	}
	var req UpdateOrgRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	org, err := h.svc.UpdateOrganization(r.Context(), id, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(org)
}

func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	orgID := userCtx.OrgID
	org, err := h.svc.GetProfile(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(org)
}

func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	orgID := userCtx.OrgID
	var req CompanyProfileUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	org, err := h.svc.UpdateProfile(r.Context(), orgID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(org)
}

func (h *Handler) UploadLogo(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	orgID := userCtx.OrgID
	
	file, header, err := r.FormFile("logo")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()
	
	org, err := h.svc.UploadLogo(r.Context(), orgID, header.Filename, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(org)
}

func (h *Handler) GetNotificationPreferences(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	orgID := userCtx.OrgID
	prefs, err := h.svc.GetNotificationPreferences(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(prefs)
}

func (h *Handler) UpdateNotificationPreferences(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	orgID := userCtx.OrgID
	var req UpdateNotificationPreferencesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	prefs, err := h.svc.UpdateNotificationPreferences(r.Context(), orgID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(prefs)
}

// Mailboxes & Email Settings
func (h *Handler) HandleGetEmailSettings(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	orgID := userCtx.OrgID
	settings, err := h.svc.GetEmailSettings(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(settings)
}

func (h *Handler) HandleUpdateEmailSettings(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	orgID := userCtx.OrgID
	var req UpdateEmailSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err := h.svc.UpdateEmailSettings(r.Context(), orgID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) HandleGetConnectedMailboxes(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	orgID := userCtx.OrgID
	log.Printf("[Org Handler] GetConnectedMailboxes called for orgID: %d", orgID)
	mailboxes, err := h.svc.GetConnectedMailboxes(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mailboxes)
}

func (h *Handler) HandleGetConnectedMailboxByID(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	orgID := userCtx.OrgID
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	mailbox, err := h.svc.GetConnectedMailboxByID(r.Context(), id, orgID)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			http.Error(w, "Mailbox not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mailbox)
}

func (h *Handler) HandleConnectMailbox(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	orgID := userCtx.OrgID
	var req ConnectMailboxRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	mailbox, err := h.svc.ConnectMailbox(r.Context(), orgID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(mailbox)
}

func (h *Handler) HandleUpdateMailbox(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	orgID := userCtx.OrgID
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	var req UpdateMailboxRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = h.svc.UpdateMailbox(r.Context(), id, orgID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) HandleRemoveMailbox(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	orgID := userCtx.OrgID
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	err = h.svc.RemoveMailbox(r.Context(), id, orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) HandleDisconnectMailbox(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	orgID := userCtx.OrgID
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	err = h.svc.DisconnectMailbox(r.Context(), id, orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) HandleSyncMailbox(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	orgID := userCtx.OrgID
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	err = h.svc.SyncMailbox(r.Context(), id, orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) HandleToggleMailboxProcessing(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	orgID := userCtx.OrgID
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	var req struct { Pause bool `json:"pause"` }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = h.svc.ToggleMailboxProcessing(r.Context(), id, orgID, req.Pause)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// Carrier Integrations
func (h *Handler) HandleGetCarrierIntegrations(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	orgID := userCtx.OrgID
	integrations, err := h.svc.GetCarrierIntegrations(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(integrations)
}

func (h *Handler) HandleConnectCarrier(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	orgID := userCtx.OrgID
	var req ConnectCarrierRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ci, err := h.svc.ConnectCarrier(r.Context(), orgID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ci)
}

func (h *Handler) HandleUpdateCarrier(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	orgID := userCtx.OrgID
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	var req UpdateCarrierRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = h.svc.UpdateCarrier(r.Context(), id, orgID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

func (h *Handler) HandleRemoveCarrier(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	orgID := userCtx.OrgID
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	err = h.svc.RemoveCarrier(r.Context(), id, orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

func (h *Handler) HandleTestCarrierConnection(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	orgID := userCtx.OrgID
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	err = h.svc.TestCarrierConnection(r.Context(), id, orgID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "message": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Connection test successful"})
}

func (h *Handler) HandleSyncCarrier(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	orgID := userCtx.OrgID
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	err = h.svc.SyncCarrier(r.Context(), id, orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Sync completed"})
}

// Invite is handled via users service now, just a stub
func (h *Handler) Invite(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) HandleStartGmailOAuth(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	orgID := userCtx.OrgID

	authURL, err := h.svc.StartGmailOAuth(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"url": authURL})
}

func (h *Handler) HandleGmailOAuthCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	callbackErr := r.URL.Query().Get("error")

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}
	frontendURL = strings.TrimRight(frontendURL, "/")

	if callbackErr != "" {
		redirectURL := fmt.Sprintf("%s/dashboard/settings/email-settings?oauth=error&message=%s", frontendURL, url.QueryEscape(callbackErr))
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
		return
	}

	if code == "" || state == "" {
		redirectURL := fmt.Sprintf("%s/dashboard/settings/email-settings?oauth=error&message=%s", frontendURL, url.QueryEscape("missing authorization code or state"))
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
		return
	}

	_, err := h.svc.CompleteGmailOAuth(r.Context(), code, state)
	if err != nil {
		redirectURL := fmt.Sprintf("%s/dashboard/settings/email-settings?oauth=error&message=%s", frontendURL, url.QueryEscape(err.Error()))
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
		return
	}

	redirectURL := frontendURL + "/dashboard/settings/email-settings?oauth=success"
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

