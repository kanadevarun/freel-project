package contracts

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"time"

	"github.com/freel/backend/internal/files"
	"github.com/freel/backend/internal/middleware"
	"github.com/freel/backend/internal/rates"
	"github.com/google/uuid"
)

// Service interface defines the contract management and human-in-the-loop review operations.
type Service interface {
	// UploadContract uploads a contract document, creates the database tracking record, and triggers sidecar processing.
	UploadContract(ctx context.Context, orgID int64, userID int64, filename string, r io.Reader, size int64, carrierSCAC *string) (*ContractDocument, error)
	// GetDocument fetches a single contract document with its extraction logs and summaries.
	GetDocument(ctx context.Context, orgID int64, docID string) (*ContractDocument, error)
	// ListDocuments lists all contract documents, optionally filtered by status.
	ListDocuments(ctx context.Context, orgID int64, status *ProcessingStatus) ([]ContractDocument, error)
	// TriggerReprocessing resets log counters and re-triggers the sidecar extraction for an existing file.
	TriggerReprocessing(ctx context.Context, orgID int64, docID string) error
	// HandleAICallback ingests confirmed rates and creates review items for flagged extractions.
	HandleAICallback(ctx context.Context, callback AIProcessingCallback) error

	// ListReviewItems gets review queue items filtered by status.
	ListReviewItems(ctx context.Context, orgID int64, status *ReviewStatus) ([]RateReviewItem, error)
	// ApproveReviewItem corrects/approves a flagged rate item, normalizes ports, and upserts it into the live rate list.
	ApproveReviewItem(ctx context.Context, orgID int64, id string, reviewerID int64, correctedData []byte, notes string) error
	// RejectReviewItem marks a flagged review rate item as rejected.
	RejectReviewItem(ctx context.Context, orgID int64, id string, reviewerID int64, notes string) error
}

// AIProcessingCallback represents the structured payload posted by the Python AI sidecar upon extraction completion.
type AIProcessingCallback struct {
	// DocumentID corresponds to the ID of the contract document.
	DocumentID string `json:"document_id"`
	// OrgID corresponds to the target tenant organization ID.
	OrgID int64 `json:"org_id"`
	// Status is either COMPLETED or FAILED.
	Status string `json:"status"`
	// ConfirmedRates contains high-confidence parsed rates to be auto-ingested.
	ConfirmedRates []rates.CanonicalRate `json:"confirmed_rates"`
	// FlaggedItems contains anomalous rates queued for human operators.
	FlaggedItems []ReviewItemDraft `json:"flagged_items"`
	// ProcessingLog lists step-by-step agent logs from the sidecar.
	ProcessingLog []LogEntry `json:"processing_log"`
	// AISummary is a brief textual abstract of the contract rules and terms.
	AISummary string `json:"ai_summary"`
	// CorrelationID is the trace identifier for logging.
	CorrelationID string `json:"correlation_id"`
}

// ReviewItemDraft represents the data shape for a flagged rate entry sent by the sidecar callback.
type ReviewItemDraft struct {
	// ExtractedData represents the proposed rate structure in JSON format.
	ExtractedData interface{} `json:"extracted_data"`
	// ConfidenceScore is the quality estimation of the extraction (0-100).
	ConfidenceScore int `json:"confidence_score"`
	// ReviewFlags lists reasons why the rate was flagged (e.g. PRICE_ANOMALY).
	ReviewFlags []string `json:"review_flags"`
	// AIReasoning explains the anomaly diagnostic.
	AIReasoning string `json:"ai_reasoning"`
	// SourcePage is the PDF page number index where the rate was located.
	SourcePage int `json:"source_page"`
	// SourceText is the surrounding note quote sentence.
	SourceText string `json:"source_text"`
	// SourceImageURL is the reference screenshot URL.
	SourceImageURL string `json:"source_image_url"`
}

// service implements the Service interface.
type service struct {
	repo        Repository
	fileSvc     files.Service
	aiBridge    AIBridge
	rateSvc     rates.Service
	callbackURL string
}

// NewService acts as a constructor.
// callbackURL must be the fully-qualified URL the Python sidecar should POST results to,
// e.g. http://backend:8080/internal/contracts/callback. It is supplied by main.go
// using the resolved GO_BACKEND_URL environment variable.
func NewService(repo Repository, fileSvc files.Service, aiBridge AIBridge, rateSvc rates.Service, callbackURL string) Service {
	if callbackURL == "" {
		// This should never happen in production: main.go always provides the URL.
		// Fail fast rather than silently routing callbacks to localhost.
		log.Fatal("[contracts] NewService: callbackURL is required but was not provided. Set GO_BACKEND_URL.")
	}
	return &service{
		repo:        repo,
		fileSvc:     fileSvc,
		aiBridge:    aiBridge,
		rateSvc:     rateSvc,
		callbackURL: callbackURL,
	}
}

// UploadContract uploads the document, writes the database record, and dispatches the pipeline processing task in a background goroutine.
func (s *service) UploadContract(ctx context.Context, orgID int64, userID int64, filename string, r io.Reader, size int64, carrierSCAC *string) (*ContractDocument, error) {
	// 1. Upload the physical document file using the file storage service provider.
	s3Key, err := s.fileSvc.UploadFile(ctx, filename, r)
	if err != nil {
		return nil, fmt.Errorf("upload file: %w", err)
	}

	docID := uuid.NewString()

	// Parse file extension to classify the document layout format.
	ext := stringsToLower(filepath.Ext(filename))
	fileType := "PDF"
	if ext == ".xlsx" || ext == ".xls" {
		fileType = "XLSX"
	}

	doc := &ContractDocument{
		ID:          docID,
		OrgID:       orgID,
		CarrierSCAC: carrierSCAC,
		CarrierName: nil, // Populated later by the Classifier Agent in the sidecar.
		FileName:    filename,
		S3Key:       s3Key,
		FileType:    fileType,
		FileSize:    size,
		Status:      StatusQueued,
		ProcessingLog: []LogEntry{
			{Step: "UPLOAD", Timestamp: time.Now().UTC(), Message: "Document uploaded successfully"},
		},
		CreatedBy: userID,
	}

	// 2. Persist the contract document entry in the DB.
	if err := s.repo.CreateDocument(ctx, doc); err != nil {
		return nil, fmt.Errorf("save document record: %w", err)
	}

	// 3. Queue processing task into the PostgreSQL database task table.
	//
	// Simple meaning:
	//   Instead of calling the sidecar API over HTTP, we write the request parameters
	//   into the 'ai_processing_tasks' table in PostgreSQL. The background worker running
	//   in Python will poll the table and run the LangGraph workflow.
	// Extract end-to-end trace correlation ID. Generate new one if absent in context.
	corrID := middleware.GetCorrelationID(ctx)
	if corrID == "" {
		corrID = uuid.NewString()
	}

	taskPayload := map[string]interface{}{
		"document_id":    doc.ID,
		"org_id":         doc.OrgID,
		"s3_key":         doc.S3Key,
		"file_type":      doc.FileType,
		"callback_url":   s.callbackURL,
		"correlation_id": corrID,
	}
	if err := s.repo.CreateAITask(ctx, doc.OrgID, doc.ID, "PROCESS", taskPayload); err != nil {
		fmt.Printf("[ContractsService] Failed to queue AI task for doc %s: %v\n", doc.ID, err)

		// If queueing fails, mark status as FAILED.
		failedLog := append(doc.ProcessingLog, LogEntry{
			Step:      "AI_QUEUE_FAIL",
			Timestamp: time.Now().UTC(),
			Message:   fmt.Sprintf("Failed to queue AI processing task: %v", err),
		})
		_ = s.repo.UpdateDocumentStatus(ctx, doc.OrgID, doc.ID, ProcessingStatus(StatusFailed), failedLog)
	} else {
		// Advance log indicating it is queued.
		queuedLog := append(doc.ProcessingLog, LogEntry{
			Step:      "AI_QUEUED",
			Timestamp: time.Now().UTC(),
			Message:   "Queued in database task processing pipeline",
		})
		_ = s.repo.UpdateDocumentStatus(ctx, doc.OrgID, doc.ID, StatusOCRProcessing, queuedLog)
	}

	return doc, nil
}

// GetDocument retrieves the document information from database repository.
func (s *service) GetDocument(ctx context.Context, orgID int64, docID string) (*ContractDocument, error) {
	return s.repo.GetDocumentByID(ctx, orgID, docID)
}

// ListDocuments lists documents for the organization.
func (s *service) ListDocuments(ctx context.Context, orgID int64, status *ProcessingStatus) ([]ContractDocument, error) {
	return s.repo.ListDocuments(ctx, orgID, status)
}

// TriggerReprocessing updates logs and starts extraction processing loop on the file.
func (s *service) TriggerReprocessing(ctx context.Context, orgID int64, docID string) error {
	doc, err := s.repo.GetDocumentByID(ctx, orgID, docID)
	if err != nil {
		return err
	}

	reprocessLog := append(doc.ProcessingLog, LogEntry{
		Step:      "REPROCESS",
		Timestamp: time.Now().UTC(),
		Message:   "Reprocessing triggered by user",
	})
	if err := s.repo.UpdateDocumentStatus(ctx, orgID, docID, StatusQueued, reprocessLog); err != nil {
		return err
	}

	// Queue processing task into the PostgreSQL database task table with trace propagation.
	corrID := middleware.GetCorrelationID(ctx)
	if corrID == "" {
		corrID = uuid.NewString()
	}

	taskPayload := map[string]interface{}{
		"document_id":    doc.ID,
		"org_id":         doc.OrgID,
		"s3_key":         doc.S3Key,
		"file_type":      doc.FileType,
		"callback_url":   s.callbackURL,
		"correlation_id": corrID,
	}
	if err := s.repo.CreateAITask(ctx, doc.OrgID, doc.ID, "PROCESS", taskPayload); err != nil {
		fmt.Printf("[ContractsService] Failed to queue reprocess AI task for doc %s: %v\n", doc.ID, err)
		return err
	}

	return nil
}

// HandleAICallback maps high-confidence rates and saves flagged rate items to the verification queue.
func (s *service) HandleAICallback(ctx context.Context, callback AIProcessingCallback) error {
	fmt.Printf("[ContractsService][Correlation ID: %s] Received AI sidecar callback for document %s, status=%s\n", callback.CorrelationID, callback.DocumentID, callback.Status)

	// 1. Load document record.
	doc, err := s.repo.GetDocumentByID(ctx, callback.OrgID, callback.DocumentID)
	if err != nil {
		return fmt.Errorf("get document %s: %w", callback.DocumentID, err)
	}

	// Combine incoming logs with current document history.
	updatedLog := append(doc.ProcessingLog, callback.ProcessingLog...)

	var finalStatus ProcessingStatus
	if callback.Status == "COMPLETED" {
		// Shift document status depending on whether there are pending review queue items.
		if len(callback.FlaggedItems) > 0 {
			finalStatus = StatusPendingReview
			updatedLog = append(updatedLog, LogEntry{
				Step:      "COMPLETED",
				Timestamp: time.Now().UTC(),
				Message:   fmt.Sprintf("AI extraction complete: %d rates confirmed, %d rates flagged for review", len(callback.ConfirmedRates), len(callback.FlaggedItems)),
			})
		} else {
			finalStatus = StatusConfirmed
			updatedLog = append(updatedLog, LogEntry{
				Step:      "COMPLETED",
				Timestamp: time.Now().UTC(),
				Message:   fmt.Sprintf("AI extraction complete: all %d rates successfully confirmed", len(callback.ConfirmedRates)),
			})
		}
	} else {
		finalStatus = ProcessingStatus(StatusFailed)
		updatedLog = append(updatedLog, LogEntry{
			Step:      "FAILED",
			Timestamp: time.Now().UTC(),
			Message:   "Extraction failed in AI sidecar pipeline",
		})
	}

	// Save status transitions to database.
	if err := s.repo.UpdateDocumentStatus(ctx, callback.OrgID, callback.DocumentID, finalStatus, updatedLog); err != nil {
		return fmt.Errorf("update status: %w", err)
	}

	// Record summary counts.
	extractedCount := len(callback.ConfirmedRates) + len(callback.FlaggedItems)
	if err := s.repo.UpdateDocumentSummaryAndCounts(ctx, callback.OrgID, callback.DocumentID,
		callback.AISummary, extractedCount, len(callback.ConfirmedRates), len(callback.FlaggedItems), 0,
	); err != nil {
		return fmt.Errorf("update summary and counts: %w", err)
	}

	// 2. Process high-confidence rate entries. Ingest them directly.
	if len(callback.ConfirmedRates) > 0 {
		for i := range callback.ConfirmedRates {
			// Tag rate attributes.
			callback.ConfirmedRates[i].OrgID = callback.OrgID
			callback.ConfirmedRates[i].Source = rates.RateSourceContractPDF
			callback.ConfirmedRates[i].ContractDocID = &callback.DocumentID
			callback.ConfirmedRates[i].ExtractionStatus = rates.ExtractionStatusConfirmed
			if callback.ConfirmedRates[i].ID == "" {
				callback.ConfirmedRates[i].ID = uuid.NewString()
			}
		}

		// Insert confirmed rates into live rate_entries list.
		if err := s.rateSvc.IngestRates(ctx, callback.ConfirmedRates); err != nil {
			return fmt.Errorf("ingest confirmed rates: %w", err)
		}
	}

	// 3. Process flagged anomalies. Append them to the human validation queue.
	for _, itemDraft := range callback.FlaggedItems {
		extractedDataBytes, err := json.Marshal(itemDraft.ExtractedData)
		if err != nil {
			return fmt.Errorf("marshal extracted draft: %w", err)
		}

		var pageNum *int
		if itemDraft.SourcePage > 0 {
			pageNum = &itemDraft.SourcePage
		}
		var textVal *string
		if itemDraft.SourceText != "" {
			textVal = &itemDraft.SourceText
		}
		var imgURL *string
		if itemDraft.SourceImageURL != "" {
			imgURL = &itemDraft.SourceImageURL
		}
		var reason *string
		if itemDraft.AIReasoning != "" {
			reason = &itemDraft.AIReasoning
		}

		reviewItem := &RateReviewItem{
			ID:             uuid.NewString(),
			OrgID:          callback.OrgID,
			ContractDocID:  callback.DocumentID,
			ExtractedData:  extractedDataBytes,
			Confidence:     itemDraft.ConfidenceScore,
			ReviewFlags:    itemDraft.ReviewFlags,
			AIReasoning:    reason,
			SourcePage:     pageNum,
			SourceText:     textVal,
			SourceImageURL: imgURL,
			Status:         ReviewStatusPending,
		}

		if err := s.repo.CreateReviewItem(ctx, reviewItem); err != nil {
			return fmt.Errorf("create review item: %w", err)
		}
	}

	return nil
}

// ListReviewItems gets review queue items.
func (s *service) ListReviewItems(ctx context.Context, orgID int64, status *ReviewStatus) ([]RateReviewItem, error) {
	return s.repo.ListReviewItems(ctx, orgID, status)
}

// ApproveReviewItem corrects, normalizes, and ingests a flagged rate item.
func (s *service) ApproveReviewItem(ctx context.Context, orgID int64, id string, reviewerID int64, correctedData []byte, notes string) error {
	item, err := s.repo.GetReviewItemByID(ctx, orgID, id)
	if err != nil {
		return fmt.Errorf("get review item: %w", err)
	}

	// Default to original AI data if no corrections were submitted.
	dataToIngest := item.ExtractedData
	reviewStatus := ReviewStatusApproved

	if len(correctedData) > 0 {
		dataToIngest = correctedData
		reviewStatus = ReviewStatusCorrected
	}

	// Parse JSON bytes to CanonicalRate structure.
	var rate rates.CanonicalRate
	if err := json.Unmarshal(dataToIngest, &rate); err != nil {
		return fmt.Errorf("parse rate json: %w", err)
	}

	// Enforce metadata.
	rate.OrgID = orgID
	rate.Source = rates.RateSourceContractPDF
	rate.ContractDocID = &item.ContractDocID
	rate.ExtractionStatus = rates.ExtractionStatusConfirmed
	rate.ConfidenceScore = 100 // Humans are always 100% confident.
	rate.ExtractedBy = fmt.Sprintf("human:%d", reviewerID)

	// Run port normalizer to resolve any manual typos to UN/LOCODEs.
	rate.OriginPort = rates.NormalizePort(rate.OriginPort)
	rate.DestinationPort = rates.NormalizePort(rate.DestinationPort)

	if rate.ID == "" {
		rate.ID = uuid.NewString()
	}

	// Save rate to live catalog.
	if err := s.rateSvc.IngestRates(ctx, []rates.CanonicalRate{rate}); err != nil {
		return fmt.Errorf("ingest rate: %w", err)
	}

	// Update queue item state in the repository.
	if err := s.repo.UpdateReviewItemStatus(ctx, orgID, id, reviewStatus, reviewerID, correctedData, notes); err != nil {
		return err
	}

	// Queue approved task resumption parameters into the PostgreSQL task table.
	var ratesList []interface{}
	if len(correctedData) > 0 {
		var parsed interface{}
		if err := json.Unmarshal(correctedData, &parsed); err == nil {
			ratesList = []interface{}{parsed}
		}
	} else {
		var parsed interface{}
		if err := json.Unmarshal(item.ExtractedData, &parsed); err == nil {
			ratesList = []interface{}{parsed}
		}
	}

	corrID := middleware.GetCorrelationID(ctx)
	if corrID == "" {
		corrID = uuid.NewString()
	}

	taskPayload := map[string]interface{}{
		"document_id":     item.ContractDocID,
		"org_id":          item.OrgID,
		"action":          "APPROVE",
		"corrected_rates": ratesList,
		"notes":           notes,
		"callback_url":    s.callbackURL,
		"correlation_id":  corrID,
	}
	if err := s.repo.CreateAITask(ctx, item.OrgID, item.ContractDocID, "RESUME", taskPayload); err != nil {
		fmt.Printf("[ContractsService] Failed to queue resume approve task for doc %s: %v\n", item.ContractDocID, err)
	}

	return nil
}

// RejectReviewItem marks rate item as rejected.
func (s *service) RejectReviewItem(ctx context.Context, orgID int64, id string, reviewerID int64, notes string) error {
	item, err := s.repo.GetReviewItemByID(ctx, orgID, id)
	if err != nil {
		return fmt.Errorf("get review item: %w", err)
	}

	// 1. Mark review status as REJECTED in SQL database.
	if err := s.repo.UpdateReviewItemStatus(ctx, orgID, id, ReviewStatusRejected, reviewerID, nil, notes); err != nil {
		return err
	}

	// 2. Queue rejected task resumption parameters into the PostgreSQL task table.
	corrID := middleware.GetCorrelationID(ctx)
	if corrID == "" {
		corrID = uuid.NewString()
	}

	taskPayload := map[string]interface{}{
		"document_id":    item.ContractDocID,
		"org_id":         item.OrgID,
		"action":         "REJECT",
		"notes":          notes,
		"callback_url":   s.callbackURL,
		"correlation_id": corrID,
	}
	if err := s.repo.CreateAITask(ctx, item.OrgID, item.ContractDocID, "RESUME", taskPayload); err != nil {
		fmt.Printf("[ContractsService] Failed to queue resume reject task for doc %s: %v\n", item.ContractDocID, err)
	}

	return nil
}

// stringsToLower converts a string to lowercase.
func stringsToLower(s string) string {
	var b []byte
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b = append(b, c)
	}
	return string(b)
}
