package predictions

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/freel/backend/internal/middleware"
	"github.com/freel/backend/internal/utils"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/", h.HandleListPredictions)
	r.Post("/generate", h.HandleGeneratePrediction)
	r.Get("/{id}", h.HandleGetPrediction)
	r.Post("/{id}/acknowledge", h.HandleAcknowledge)
	r.Post("/{id}/dismiss", h.HandleDismiss)
	r.Post("/{id}/request-action", h.HandleRequestAction)
	r.Post("/{id}/outcome", h.HandleRecordOutcome)
	r.Get("/{id}/audit", h.HandleGetAuditHistory)
	r.Get("/shipments/{id}/eta", h.HandleGetShipmentPredictedETA)
	r.Post("/shipments/{id}/eta/refresh", h.HandleRefreshShipmentPredictedETA)
}

func (h *Handler) getUser(r *http.Request) (middleware.UserContext, bool) {
	userCtx, ok := middleware.GetUserContext(r.Context())
	if !ok || userCtx.OrgID <= 0 {
		return middleware.UserContext{}, false
	}
	return userCtx, true
}

func (h *Handler) HandleListPredictions(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	params := FilterParams{
		Module:            q.Get("module"),
		Severity:          q.Get("severity"),
		ConfidenceBand:    q.Get("confidence_band"),
		Status:            q.Get("status"),
		RelatedRecordType: q.Get("related_record_type"),
		RelatedRecordID:   q.Get("related_record_id"),
		Limit:             limit,
		Offset:            offset,
	}

	items, total, err := h.svc.ListPredictions(r.Context(), int64(user.OrgID), params)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":     true,
		"predictions": items,
		"total":       total,
		"limit":       params.Limit,
		"offset":      params.Offset,
	})
}

func (h *Handler) HandleGeneratePrediction(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	var req GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid JSON payload", "BAD_REQUEST")
		return
	}

	var uid *int64
	if user.UserID > 0 {
		val := int64(user.UserID)
		uid = &val
	}

	pred, err := h.svc.GenerateAndPersist(r.Context(), int64(user.OrgID), uid, &req)
	if err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "PREDICTION_GENERATION_FAILED")
		return
	}

	utils.JSON(w, http.StatusCreated, map[string]interface{}{
		"success":    true,
		"prediction": pred,
	})
}

func (h *Handler) HandleGetPrediction(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	predID := chi.URLParam(r, "id")
	if predID == "" {
		utils.Error(w, http.StatusBadRequest, "Missing prediction ID", "BAD_REQUEST")
		return
	}

	pred, err := h.svc.GetPrediction(r.Context(), int64(user.OrgID), predID)
	if err != nil {
		if err == ErrPredictionNotFound {
			utils.Error(w, http.StatusNotFound, "Prediction not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success":    true,
		"prediction": pred,
	})
}

func (h *Handler) HandleAcknowledge(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	predID := chi.URLParam(r, "id")
	var uid *int64
	if user.UserID > 0 {
		val := int64(user.UserID)
		uid = &val
	}

	if err := h.svc.AcknowledgePrediction(r.Context(), int64(user.OrgID), predID, uid); err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "ACTION_FAILED")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Prediction acknowledged",
	})
}

func (h *Handler) HandleDismiss(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	predID := chi.URLParam(r, "id")
	var payload struct {
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&payload)

	var uid *int64
	if user.UserID > 0 {
		val := int64(user.UserID)
		uid = &val
	}

	if err := h.svc.DismissPrediction(r.Context(), int64(user.OrgID), predID, uid, payload.Reason); err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "ACTION_FAILED")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Prediction dismissed",
	})
}

func (h *Handler) HandleRequestAction(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	predID := chi.URLParam(r, "id")
	var payload struct {
		Notes string `json:"notes"`
	}
	_ = json.NewDecoder(r.Body).Decode(&payload)

	var uid *int64
	if user.UserID > 0 {
		val := int64(user.UserID)
		uid = &val
	}

	if err := h.svc.RequestAction(r.Context(), int64(user.OrgID), predID, uid, payload.Notes); err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "ACTION_FAILED")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Action requested and queued for approval",
	})
}

func (h *Handler) HandleRecordOutcome(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	predID := chi.URLParam(r, "id")
	var payload struct {
		OutcomeStatus OutcomeStatus `json:"outcome_status"`
		OutcomeValue  *string       `json:"outcome_value"`
		FeedbackNotes *string       `json:"feedback_notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid JSON payload", "BAD_REQUEST")
		return
	}

	if err := h.svc.RecordOutcome(r.Context(), int64(user.OrgID), predID, payload.OutcomeStatus, payload.OutcomeValue, payload.FeedbackNotes); err != nil {
		utils.Error(w, http.StatusBadRequest, err.Error(), "OUTCOME_RECORD_FAILED")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Actual outcome recorded successfully",
	})
}

func (h *Handler) HandleGetAuditHistory(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	predID := chi.URLParam(r, "id")
	history, err := h.svc.GetAuditHistory(r.Context(), int64(user.OrgID), predID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "INTERNAL_ERROR")
		return
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"history": history,
	})
}

func (h *Handler) HandleGetShipmentPredictedETA(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	shipmentID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || shipmentID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid shipment id", "INVALID_SHIPMENT_ID")
		return
	}

	forceRefresh := r.URL.Query().Get("refresh") == "true"
	pred, err := h.svc.GetOrPredictShipmentETA(r.Context(), int64(user.OrgID), &user.UserID, shipmentID, forceRefresh)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "PREDICTION_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Shipment ETA prediction retrieved", pred)
}

func (h *Handler) HandleRefreshShipmentPredictedETA(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	shipmentID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || shipmentID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid shipment id", "INVALID_SHIPMENT_ID")
		return
	}

	pred, err := h.svc.GetOrPredictShipmentETA(r.Context(), int64(user.OrgID), &user.UserID, shipmentID, true)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "PREDICTION_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Shipment ETA prediction refreshed", pred)
}

func (h *Handler) HandleGetShipmentPredictedExceptions(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	shipmentID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || shipmentID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid shipment id", "INVALID_SHIPMENT_ID")
		return
	}

	forceRefresh := r.URL.Query().Get("refresh") == "true"
	pred, err := h.svc.GetOrForecastShipmentExceptions(r.Context(), int64(user.OrgID), &user.UserID, shipmentID, forceRefresh)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "PREDICTION_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Shipment exception disruption forecast retrieved", pred)
}

func (h *Handler) HandleRefreshShipmentPredictedExceptions(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	shipmentID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || shipmentID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid shipment id", "INVALID_SHIPMENT_ID")
		return
	}

	pred, err := h.svc.GetOrForecastShipmentExceptions(r.Context(), int64(user.OrgID), &user.UserID, shipmentID, true)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "PREDICTION_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Shipment exception disruption forecast refreshed", pred)
}

func (h *Handler) HandleGetLeadPredictedIntelligence(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	leadID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || leadID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid lead id", "INVALID_LEAD_ID")
		return
	}

	forceRefresh := r.URL.Query().Get("refresh") == "true"
	pred, err := h.svc.GetOrPredictLeadIntelligence(r.Context(), int64(user.OrgID), &user.UserID, leadID, forceRefresh)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "PREDICTION_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Lead predictive intelligence retrieved", pred)
}

func (h *Handler) HandleRefreshLeadPredictedIntelligence(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	leadID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || leadID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid lead id", "INVALID_LEAD_ID")
		return
	}

	pred, err := h.svc.GetOrPredictLeadIntelligence(r.Context(), int64(user.OrgID), &user.UserID, leadID, true)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "PREDICTION_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Lead predictive intelligence refreshed", pred)
}

func (h *Handler) HandleGetCustomerPredictedIntelligence(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	customerID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || customerID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid customer id", "INVALID_CUSTOMER_ID")
		return
	}

	forceRefresh := r.URL.Query().Get("refresh") == "true"
	pred, err := h.svc.GetOrPredictCustomerIntelligence(r.Context(), int64(user.OrgID), &user.UserID, customerID, forceRefresh)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "PREDICTION_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Customer predictive intelligence retrieved", pred)
}

func (h *Handler) HandleRefreshCustomerPredictedIntelligence(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	customerID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || customerID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid customer id", "INVALID_CUSTOMER_ID")
		return
	}

	pred, err := h.svc.GetOrPredictCustomerIntelligence(r.Context(), int64(user.OrgID), &user.UserID, customerID, true)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "PREDICTION_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Customer predictive intelligence refreshed", pred)
}

func (h *Handler) HandleGetRFQPredictedMargin(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	rfqID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || rfqID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid rfq id", "INVALID_RFQ_ID")
		return
	}

	forceRefresh := r.URL.Query().Get("refresh") == "true"
	pred, err := h.svc.GetOrPredictRFQMarginIntelligence(r.Context(), int64(user.OrgID), &user.UserID, rfqID, forceRefresh)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "PREDICTION_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "RFQ predictive margin intelligence retrieved", pred)
}

func (h *Handler) HandleRefreshRFQPredictedMargin(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	rfqID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || rfqID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid rfq id", "INVALID_RFQ_ID")
		return
	}

	pred, err := h.svc.GetOrPredictRFQMarginIntelligence(r.Context(), int64(user.OrgID), &user.UserID, rfqID, true)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "PREDICTION_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "RFQ predictive margin intelligence refreshed", pred)
}

func (h *Handler) HandleGetContractPredictedRatePressure(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	contractID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || contractID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid contract id", "INVALID_CONTRACT_ID")
		return
	}

	forceRefresh := r.URL.Query().Get("refresh") == "true"
	pred, err := h.svc.GetOrPredictContractRatePressure(r.Context(), int64(user.OrgID), &user.UserID, contractID, forceRefresh)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "PREDICTION_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Contract predicted rate pressure retrieved", pred)
}

func (h *Handler) HandleRefreshContractPredictedRatePressure(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	contractID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || contractID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid contract id", "INVALID_CONTRACT_ID")
		return
	}

	pred, err := h.svc.GetOrPredictContractRatePressure(r.Context(), int64(user.OrgID), &user.UserID, contractID, true)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "PREDICTION_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Contract predicted rate pressure refreshed", pred)
}

func (h *Handler) HandleGetInvoicePredictedCollection(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	invoiceID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || invoiceID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid invoice id", "INVALID_INVOICE_ID")
		return
	}

	forceRefresh := r.URL.Query().Get("refresh") == "true"
	pred, err := h.svc.GetOrPredictInvoiceCollections(r.Context(), int64(user.OrgID), &user.UserID, invoiceID, forceRefresh)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "PREDICTION_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Invoice predicted collections intelligence retrieved", pred)
}

func (h *Handler) HandleRefreshInvoicePredictedCollection(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	invoiceID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || invoiceID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid invoice id", "INVALID_INVOICE_ID")
		return
	}

	pred, err := h.svc.GetOrPredictInvoiceCollections(r.Context(), int64(user.OrgID), &user.UserID, invoiceID, true)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "PREDICTION_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Invoice predicted collections intelligence refreshed", pred)
}

func (h *Handler) HandleGetContractPredictedComplianceRisk(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	contractID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || contractID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid contract id", "INVALID_CONTRACT_ID")
		return
	}

	forceRefresh := r.URL.Query().Get("refresh") == "true"
	pred, err := h.svc.GetOrPredictContractComplianceRisk(r.Context(), int64(user.OrgID), &user.UserID, contractID, forceRefresh)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			utils.Error(w, http.StatusNotFound, err.Error(), "CONTRACT_NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "PREDICTION_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Contract predicted compliance and documentation risk retrieved", pred)
}

func (h *Handler) HandleRefreshContractPredictedComplianceRisk(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	contractID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || contractID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid contract id", "INVALID_CONTRACT_ID")
		return
	}

	pred, err := h.svc.GetOrPredictContractComplianceRisk(r.Context(), int64(user.OrgID), &user.UserID, contractID, true)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			utils.Error(w, http.StatusNotFound, err.Error(), "CONTRACT_NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "PREDICTION_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Contract predicted compliance and documentation risk refreshed", pred)
}

func (h *Handler) HandleGetShipmentPredictedReadiness(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	shipmentID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || shipmentID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid shipment id", "INVALID_SHIPMENT_ID")
		return
	}

	forceRefresh := r.URL.Query().Get("refresh") == "true"
	pred, err := h.svc.GetOrPredictShipmentReadiness(r.Context(), int64(user.OrgID), &user.UserID, shipmentID, forceRefresh)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			utils.Error(w, http.StatusNotFound, err.Error(), "SHIPMENT_NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "PREDICTION_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Shipment predicted readiness and compliance intelligence retrieved", pred)
}

func (h *Handler) HandleRefreshShipmentPredictedReadiness(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	shipmentID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || shipmentID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid shipment id", "INVALID_SHIPMENT_ID")
		return
	}

	pred, err := h.svc.GetOrPredictShipmentReadiness(r.Context(), int64(user.OrgID), &user.UserID, shipmentID, true)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			utils.Error(w, http.StatusNotFound, err.Error(), "SHIPMENT_NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "PREDICTION_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Shipment predicted readiness and compliance intelligence refreshed", pred)
}

func (h *Handler) HandleGetCarrierPredictedPerformance(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	scac := chi.URLParam(r, "scac")
	if strings.TrimSpace(scac) == "" {
		utils.Error(w, http.StatusBadRequest, "Invalid carrier scac", "INVALID_CARRIER_SCAC")
		return
	}

	forceRefresh := r.URL.Query().Get("refresh") == "true"
	pred, err := h.svc.GetOrPredictCarrierPerformance(r.Context(), int64(user.OrgID), &user.UserID, scac, forceRefresh)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			utils.Error(w, http.StatusNotFound, err.Error(), "CARRIER_NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "PREDICTION_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Carrier predicted performance intelligence retrieved", pred)
}

func (h *Handler) HandleRefreshCarrierPredictedPerformance(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	scac := chi.URLParam(r, "scac")
	if strings.TrimSpace(scac) == "" {
		utils.Error(w, http.StatusBadRequest, "Invalid carrier scac", "INVALID_CARRIER_SCAC")
		return
	}

	pred, err := h.svc.GetOrPredictCarrierPerformance(r.Context(), int64(user.OrgID), &user.UserID, scac, true)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			utils.Error(w, http.StatusNotFound, err.Error(), "CARRIER_NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "PREDICTION_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Carrier predicted performance intelligence refreshed", pred)
}

func (h *Handler) HandleGetLanePredictedPerformance(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	laneCode := chi.URLParam(r, "laneCode")
	if strings.TrimSpace(laneCode) == "" {
		utils.Error(w, http.StatusBadRequest, "Invalid trade lane corridor code", "INVALID_LANE_CODE")
		return
	}

	forceRefresh := r.URL.Query().Get("refresh") == "true"
	pred, err := h.svc.GetOrPredictLanePerformance(r.Context(), int64(user.OrgID), &user.UserID, laneCode, forceRefresh)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			utils.Error(w, http.StatusNotFound, err.Error(), "LANE_NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "PREDICTION_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Trade lane predicted corridor performance retrieved", pred)
}

func (h *Handler) HandleRefreshLanePredictedPerformance(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	laneCode := chi.URLParam(r, "laneCode")
	if strings.TrimSpace(laneCode) == "" {
		utils.Error(w, http.StatusBadRequest, "Invalid trade lane corridor code", "INVALID_LANE_CODE")
		return
	}

	pred, err := h.svc.GetOrPredictLanePerformance(r.Context(), int64(user.OrgID), &user.UserID, laneCode, true)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			utils.Error(w, http.StatusNotFound, err.Error(), "LANE_NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "PREDICTION_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Trade lane predicted corridor performance refreshed", pred)
}

func (h *Handler) HandleGetCustomerPredictedServicePerformance(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	customerID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || customerID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid customer id", "INVALID_CUSTOMER_ID")
		return
	}

	forceRefresh := r.URL.Query().Get("refresh") == "true"
	pred, err := h.svc.GetOrPredictCustomerServicePerformance(r.Context(), int64(user.OrgID), &user.UserID, customerID, forceRefresh)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			utils.Error(w, http.StatusNotFound, err.Error(), "CUSTOMER_NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "PREDICTION_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Customer predicted service and relationship performance retrieved", pred)
}

func (h *Handler) HandleRefreshCustomerPredictedServicePerformance(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	customerID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || customerID <= 0 {
		utils.Error(w, http.StatusBadRequest, "Invalid customer id", "INVALID_CUSTOMER_ID")
		return
	}

	pred, err := h.svc.GetOrPredictCustomerServicePerformance(r.Context(), int64(user.OrgID), &user.UserID, customerID, true)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			utils.Error(w, http.StatusNotFound, err.Error(), "CUSTOMER_NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "PREDICTION_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Customer predicted service and relationship performance refreshed", pred)
}

func (h *Handler) HandleGetPredictedWorkload(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	area := r.URL.Query().Get("area")
	if strings.TrimSpace(area) == "" {
		area = "approvals"
	}

	forceRefresh := r.URL.Query().Get("refresh") == "true"
	pred, err := h.svc.GetOrPredictWorkloadPlanning(r.Context(), int64(user.OrgID), &user.UserID, area, forceRefresh)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "PREDICTION_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Predicted workload planning intelligence retrieved", pred)
}

func (h *Handler) HandleRefreshPredictedWorkload(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	area := r.URL.Query().Get("area")
	if strings.TrimSpace(area) == "" {
		area = "approvals"
	}

	pred, err := h.svc.GetOrPredictWorkloadPlanning(r.Context(), int64(user.OrgID), &user.UserID, area, true)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "PREDICTION_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Predicted workload planning intelligence refreshed", pred)
}

func (h *Handler) HandleGetPredictedCapacity(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	dimension := r.URL.Query().Get("dimension")
	if strings.TrimSpace(dimension) == "" {
		dimension = "corridors"
	}

	forceRefresh := r.URL.Query().Get("refresh") == "true"
	pred, err := h.svc.GetOrPredictCapacityPlanning(r.Context(), int64(user.OrgID), &user.UserID, dimension, forceRefresh)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "PREDICTION_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Predicted capacity planning intelligence retrieved", pred)
}

func (h *Handler) HandleRefreshPredictedCapacity(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	dimension := r.URL.Query().Get("dimension")
	if strings.TrimSpace(dimension) == "" {
		dimension = "corridors"
	}

	pred, err := h.svc.GetOrPredictCapacityPlanning(r.Context(), int64(user.OrgID), &user.UserID, dimension, true)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "PREDICTION_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Predicted capacity planning intelligence refreshed", pred)
}

func (h *Handler) HandleGetPredictedDemand(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	segment := r.URL.Query().Get("segment")
	if strings.TrimSpace(segment) == "" {
		segment = "commercial"
	}

	forceRefresh := r.URL.Query().Get("refresh") == "true"
	pred, err := h.svc.GetOrPredictDemandPlanning(r.Context(), int64(user.OrgID), &user.UserID, segment, forceRefresh)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "PREDICTION_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Predicted demand planning intelligence retrieved", pred)
}

func (h *Handler) HandleRefreshPredictedDemand(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	segment := r.URL.Query().Get("segment")
	if strings.TrimSpace(segment) == "" {
		segment = "commercial"
	}

	pred, err := h.svc.GetOrPredictDemandPlanning(r.Context(), int64(user.OrgID), &user.UserID, segment, true)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "PREDICTION_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Predicted demand planning intelligence refreshed", pred)
}

func (h *Handler) HandleGetWorkloadCapacitySummary(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	orgID := int64(user.OrgID)
	userID := &user.UserID

	apprWorkload, _ := h.svc.GetOrPredictWorkloadPlanning(r.Context(), orgID, userID, "approvals", false)
	docsWorkload, _ := h.svc.GetOrPredictWorkloadPlanning(r.Context(), orgID, userID, "documentation", false)
	corridorCap, _ := h.svc.GetOrPredictCapacityPlanning(r.Context(), orgID, userID, "corridors", false)
	quoteDemand, _ := h.svc.GetOrPredictDemandPlanning(r.Context(), orgID, userID, "commercial", false)

	summary := map[string]interface{}{
		"org_id":                 orgID,
		"approvals_workload":     apprWorkload,
		"documentation_workload": docsWorkload,
		"corridor_capacity":      corridorCap,
		"quote_demand":           quoteDemand,
		"generated_at":           time.Now().UTC().Format(time.RFC3339),
	}

	utils.Success(w, http.StatusOK, "Consolidated workload, capacity, and demand planning intelligence retrieved", summary)
}

func (h *Handler) HandleGetOperationalBottleneck(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	bottleneckType := r.URL.Query().Get("type")
	if strings.TrimSpace(bottleneckType) == "" {
		bottleneckType = "approvals"
	}

	forceRefresh := r.URL.Query().Get("refresh") == "true"
	pred, err := h.svc.GetOrPredictOperationalBottleneck(r.Context(), int64(user.OrgID), &user.UserID, bottleneckType, forceRefresh)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "PREDICTION_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Operational bottleneck intelligence retrieved", pred)
}

func (h *Handler) HandleRefreshOperationalBottleneck(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	bottleneckType := r.URL.Query().Get("type")
	if strings.TrimSpace(bottleneckType) == "" {
		bottleneckType = "approvals"
	}

	pred, err := h.svc.GetOrPredictOperationalBottleneck(r.Context(), int64(user.OrgID), &user.UserID, bottleneckType, true)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "PREDICTION_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Operational bottleneck intelligence refreshed", pred)
}

func (h *Handler) HandleGetResourceAllocation(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	dimension := r.URL.Query().Get("dimension")
	if strings.TrimSpace(dimension) == "" {
		dimension = "allocation"
	}

	forceRefresh := r.URL.Query().Get("refresh") == "true"
	pred, err := h.svc.GetOrPredictResourceAllocation(r.Context(), int64(user.OrgID), &user.UserID, dimension, forceRefresh)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "PREDICTION_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Resource allocation intelligence retrieved", pred)
}

func (h *Handler) HandleRefreshResourceAllocation(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	dimension := r.URL.Query().Get("dimension")
	if strings.TrimSpace(dimension) == "" {
		dimension = "allocation"
	}

	pred, err := h.svc.GetOrPredictResourceAllocation(r.Context(), int64(user.OrgID), &user.UserID, dimension, true)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "PREDICTION_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Resource allocation intelligence refreshed", pred)
}

func (h *Handler) HandleGetResourceBottleneckSummary(w http.ResponseWriter, r *http.Request) {
	user, ok := h.getUser(r)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	summary, err := h.svc.GetResourceBottleneckSummary(r.Context(), int64(user.OrgID), &user.UserID)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "PREDICTION_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Consolidated resource allocation and operational bottleneck intelligence retrieved", summary)
}






