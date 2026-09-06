package leads

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/freel/backend/internal/leads/spec"
	"github.com/freel/backend/internal/svcerror"
	"github.com/go-chi/chi/v5"
	kitHttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

func AddLeadsHandlers(
	router chi.Router,
	endpoints Endpoints,
	authMiddleware func(http.Handler) http.Handler,
) {
	options := []kitHttp.ServerOption{
		kitHttp.ServerErrorEncoder(encodeErrorResponse),
	}

	router.With(authMiddleware).Get("/", kitHttp.NewServer(
		endpoints.ListLeadsEP,
		decodeListLeadsRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Post("/", kitHttp.NewServer(
		endpoints.CreateLeadEP,
		decodeCreateLeadRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Post("/import", kitHttp.NewServer(
		endpoints.ImportLeadsEP,
		decodeImportLeadsRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Get("/{id:[0-9]+}", kitHttp.NewServer(
		endpoints.GetLeadEP,
		decodeGetLeadRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Put("/{id:[0-9]+}", kitHttp.NewServer(
		endpoints.UpdateLeadEP,
		decodeUpdateLeadRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Delete("/{id:[0-9]+}", kitHttp.NewServer(
		endpoints.DeleteLeadEP,
		decodeDeleteLeadRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Post("/bulk", kitHttp.NewServer(
		endpoints.BulkUpdateLeadsEP,
		decodeBulkUpdateLeadsRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)

	router.With(authMiddleware).Get("/{id:[0-9]+}/timeline", kitHttp.NewServer(
		endpoints.GetLeadTimelineEP,
		decodeGetLeadTimelineRequest,
		encodeAPIResponse,
		options...,
	).ServeHTTP)
}

func getIDFromVars(r *http.Request) (int32, error) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		vars := mux.Vars(r)
		idStr = vars["id"]
	}
	if idStr == "" {
		return 0, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return 0, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	return int32(id), nil
}

func decodeListLeadsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	statusStr := r.URL.Query().Get("status")
	searchStr := r.URL.Query().Get("search")
	sourceStr := r.URL.Query().Get("source")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	var status *string
	if statusStr != "" {
		status = &statusStr
	}

	var search *string
	if searchStr != "" {
		search = &searchStr
	}

	var source *string
	if sourceStr != "" {
		source = &sourceStr
	}

	return &spec.ListLeadsRequest{
		Limit:  limit,
		Offset: offset,
		Status: status,
		Search: search,
		Source: source,
	}, nil
}

func decodeCreateLeadRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req spec.CreateLeadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	return &req, nil
}

func decodeImportLeadsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	// Parse multipart form (max 10MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	// Read headers
	headers, err := reader.Read()
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	companyIdx, contactIdx, emailIdx, phoneIdx, sourceIdx, notesIdx, locationIdx := -1, -1, -1, -1, -1, -1, -1
	for i, h := range headers {
		hClean := strings.ToLower(strings.TrimSpace(h))
		switch hClean {
		case "company name", "company":
			companyIdx = i
		case "contact name", "contact":
			contactIdx = i
		case "email":
			emailIdx = i
		case "phone":
			phoneIdx = i
		case "source":
			sourceIdx = i
		case "notes":
			notesIdx = i
		case "location":
			locationIdx = i
		}
	}

	if companyIdx == -1 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	var req []*spec.CreateLeadRequest
	records, err := reader.ReadAll()
	if err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	for _, record := range records {
		if companyIdx >= len(record) || strings.TrimSpace(record[companyIdx]) == "" {
			continue
		}

		companyName := strings.TrimSpace(record[companyIdx])
		var contactName, email, phone, source, notes, location *string

		if contactIdx >= 0 && contactIdx < len(record) && strings.TrimSpace(record[contactIdx]) != "" {
			val := strings.TrimSpace(record[contactIdx])
			contactName = &val
		}
		if emailIdx >= 0 && emailIdx < len(record) && strings.TrimSpace(record[emailIdx]) != "" {
			val := strings.TrimSpace(record[emailIdx])
			email = &val
		}
		if phoneIdx >= 0 && phoneIdx < len(record) && strings.TrimSpace(record[phoneIdx]) != "" {
			val := strings.TrimSpace(record[phoneIdx])
			phone = &val
		}
		if sourceIdx >= 0 && sourceIdx < len(record) && strings.TrimSpace(record[sourceIdx]) != "" {
			val := strings.TrimSpace(record[sourceIdx])
			source = &val
		}
		if notesIdx >= 0 && notesIdx < len(record) && strings.TrimSpace(record[notesIdx]) != "" {
			val := strings.TrimSpace(record[notesIdx])
			notes = &val
		}
		if locationIdx >= 0 && locationIdx < len(record) && strings.TrimSpace(record[locationIdx]) != "" {
			val := strings.TrimSpace(record[locationIdx])
			location = &val
		}

		req = append(req, &spec.CreateLeadRequest{
			CompanyName: companyName,
			ContactName: contactName,
			Email:       email,
			Phone:       phone,
			Source:      source,
			Notes:       notes,
			Location:    location,
		})
	}

	return &spec.ImportLeadsRequest{Leads: req}, nil
}

func decodeGetLeadRequest(_ context.Context, r *http.Request) (interface{}, error) {
	id, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	return &spec.GetLeadRequest{ID: id}, nil
}

func decodeUpdateLeadRequest(_ context.Context, r *http.Request) (interface{}, error) {
	id, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	var req spec.UpdateLeadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	req.ID = id
	return &req, nil
}

func decodeDeleteLeadRequest(_ context.Context, r *http.Request) (interface{}, error) {
	id, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	return &spec.DeleteLeadRequest{ID: id}, nil
}

func decodeBulkUpdateLeadsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req spec.BulkUpdateLeadsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	return &req, nil
}

func decodeGetLeadTimelineRequest(_ context.Context, r *http.Request) (interface{}, error) {
	id, err := getIDFromVars(r)
	if err != nil {
		return nil, err
	}
	return &spec.GetLeadRequest{ID: id}, nil
}

func encodeAPIResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(response)
}

func encodeErrorResponse(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if svcErr, ok := err.(*svcerror.ServiceError); ok {
		switch svcErr.Code {
		case svcerror.ErrInvalidArgument:
			w.WriteHeader(http.StatusBadRequest)
		case svcerror.ErrInsufficientResourceAccess:
			w.WriteHeader(http.StatusUnauthorized)
		case svcerror.ErrResourceNotFound:
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"code":    svcErr.Code,
				"message": svcErr.Message,
			},
		})
		return
	}

	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]interface{}{
			"code":    svcerror.ErrInternal,
			"message": err.Error(),
		},
	})
}
