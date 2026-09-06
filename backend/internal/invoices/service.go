package invoices

import (
	"context"
	"errors"
	"fmt"
	"github.com/freel/backend/internal/audit"
	"github.com/freel/backend/internal/audit/domain"
	"math"
	"math/rand"
	"strings"
	"time"
)

var (
	ErrInvoiceNotFound = errors.New("invoice not found")
	ErrInvalidCustomer = errors.New("customer ID is required and must be valid")
	ErrInvalidAmount   = errors.New("invoice amounts must be non-negative")
)

type Service interface {
	GetInvoices(ctx context.Context, orgID int64, params ListInvoiceParams, currentUserID int64) ([]*Invoice, int, error)
	GetInvoiceByID(ctx context.Context, orgID int64, id int64) (*Invoice, error)
	GetKPIStats(ctx context.Context, orgID int64) (*InvoiceKPIStats, error)
	CreateInvoice(ctx context.Context, orgID int64, currentUserID int64, creatorName string, input CreateInvoiceInput) (*Invoice, error)
	UpdateDraftInvoice(ctx context.Context, orgID int64, id int64, input CreateInvoiceInput, userName string) (*Invoice, error)
	IssueInvoice(ctx context.Context, orgID int64, id int64, userName string) (*Invoice, error)
	SubmitForApproval(ctx context.Context, orgID int64, id int64, userName string) (*Invoice, error)
	UpdateInvoiceStatus(ctx context.Context, orgID int64, id int64, status string, userName string) error
	ToggleBookmark(ctx context.Context, orgID int64, id int64) (bool, error)
	CancelInvoice(ctx context.Context, orgID int64, id int64, reason string, userName string) error
	RecordPayment(ctx context.Context, orgID int64, invoiceID int64, input RecordPaymentInput, userName string) (*Invoice, error)
	GetAllPayments(ctx context.Context, orgID int64) ([]InvoicePayment, error)
	AddDocument(ctx context.Context, orgID int64, invoiceID int64, docName, fileSize, fileType, s3Key string) (*InvoiceDocument, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetInvoices(ctx context.Context, orgID int64, params ListInvoiceParams, currentUserID int64) ([]*Invoice, int, error) {
	return s.repo.GetInvoices(ctx, orgID, params, currentUserID)
}

func (s *service) GetInvoiceByID(ctx context.Context, orgID int64, id int64) (*Invoice, error) {
	inv, err := s.repo.GetInvoiceByID(ctx, orgID, id)
	if err != nil {
		return nil, ErrInvoiceNotFound
	}
	return inv, nil
}

func (s *service) GetKPIStats(ctx context.Context, orgID int64) (*InvoiceKPIStats, error) {
	return s.repo.GetKPIStats(ctx, orgID)
}

func (s *service) CreateInvoice(ctx context.Context, orgID int64, currentUserID int64, creatorName string, input CreateInvoiceInput) (*Invoice, error) {
	if input.CustomerID <= 0 && input.CustomerName == "" {
		return nil, ErrInvalidCustomer
	}

	invNum, err := s.repo.GenerateInvoiceNumber(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate invoice number: %w", err)
	}

	// Parse dates
	invDate := time.Now()
	if input.InvoiceDate != "" {
		if t, err := time.Parse("2006-01-02", input.InvoiceDate); err == nil {
			invDate = t
		}
	}

	dueDate := invDate.AddDate(0, 0, 15)
	if input.DueDate != "" {
		if t, err := time.Parse("2006-01-02", input.DueDate); err == nil {
			dueDate = t
		}
	}

	daysDiff := int(math.Ceil(dueDate.Sub(time.Now()).Hours() / 24))
	daysLeft := fmt.Sprintf("%d days left", daysDiff)
	if daysDiff < 0 {
		daysLeft = "Overdue"
	}

	// Calculate totals safely
	var subtotal float64
	var items []InvoiceItem
	for _, item := range input.LineItems {
		itemAmt := item.Quantity * item.UnitPrice
		subtotal += itemAmt
		items = append(items, InvoiceItem{
			Description:     item.Description,
			ServiceCategory: item.ServiceCategory,
			Quantity:        item.Quantity,
			UnitPrice:       item.UnitPrice,
			TotalAmount:     itemAmt,
		})
	}

	if input.Subtotal > 0 {
		subtotal = input.Subtotal
	}

	totalAmount := subtotal + input.TaxAmount - input.DiscountAmount
	if totalAmount < 0 {
		totalAmount = 0
	}
	balanceDue := totalAmount

	currency := input.Currency
	if currency == "" {
		currency = "USD"
	}

	creator := creatorName
	if creator == "" {
		creator = "Varun Sharma"
	}

	status := input.Status
	if status == "" {
		status = "Draft"
	}

	inv := &Invoice{
		OrgID:           orgID,
		InvoiceNumber:   invNum,
		CustomerID:      input.CustomerID,
		CustomerName:    input.CustomerName,
		CustomerCountry: input.CustomerCountry,
		ShipmentID:      input.ShipmentID,
		ShipmentNumber:  input.ShipmentNumber,
		BookingID:       input.BookingID,
		BookingNumber:   input.BookingNumber,
		QuotationID:     input.QuotationID,
		QuoteNumber:     input.QuoteNumber,
		Route:           input.Route,
		Origin:          input.Origin,
		Destination:     input.Destination,
		InvoiceDate:     invDate,
		DueDate:         dueDate,
		DaysLeft:        daysLeft,
		Currency:        currency,
		Subtotal:        subtotal,
		TaxAmount:       input.TaxAmount,
		DiscountAmount:  input.DiscountAmount,
		TotalAmount:     totalAmount,
		PaidAmount:      0.00,
		BalanceDue:      balanceDue,
		Status:          status,
		Type:            "CUSTOMER_AR",
		Bookmarked:      false,
		IsMyInvoice:     true,
		CreatorName:     "By " + creator,
		CreatedByID:     &currentUserID,
	}

	created, err := s.repo.CreateInvoice(ctx, inv, items)
	if err != nil {
		return nil, err
	}

	_, _ = audit.Record(ctx, domain.CreateAuditLogParams{
		OrgID:        orgID,
		ActorID:      &currentUserID,
		ActorName:    creatorName,
		Action:       domain.ActionCreate,
		Module:       domain.ModuleInvoices,
		ResourceType: "INVOICE",
		ResourceID:   fmt.Sprintf("%d", created.ID),
		ResourceName: created.InvoiceNumber,
		Description:  fmt.Sprintf("Created invoice %s for %s (%s %.2f)", created.InvoiceNumber, created.CustomerName, created.Currency, created.TotalAmount),
		Result:       domain.ResultSuccess,
	})

	return created, nil
}

func ValidateStatusTransition(current, next string) error {
	if current == next {
		return nil
	}

	switch current {
	case "Draft":
		if next == "Issued" || next == "Pending Approval" || next == "Cancelled" {
			return nil
		}
	case "Pending Approval":
		if next == "Issued" || next == "Draft" || next == "Cancelled" {
			return nil
		}
	case "Issued":
		if next == "Partially Paid" || next == "Paid" || next == "Overdue" || next == "Cancelled" {
			return nil
		}
	case "Partially Paid":
		if next == "Paid" || next == "Overdue" {
			return nil
		}
	case "Overdue":
		if next == "Partially Paid" || next == "Paid" || next == "Cancelled" || next == "Issued" {
			return nil
		}
	case "Paid":
		return errors.New("paid invoices are locked and cannot change status directly")
	case "Cancelled":
		return errors.New("cancelled invoices are read-only and cannot change status")
	}

	return fmt.Errorf("invalid invoice status transition from '%s' to '%s'", current, next)
}

func (s *service) UpdateDraftInvoice(ctx context.Context, orgID int64, id int64, input CreateInvoiceInput, userName string) (*Invoice, error) {
	existing, err := s.repo.GetInvoiceByID(ctx, orgID, id)
	if err != nil {
		return nil, ErrInvoiceNotFound
	}

	if existing.Status != "Draft" {
		return nil, fmt.Errorf("only draft invoices can be edited (current status: %s)", existing.Status)
	}

	if input.CustomerID <= 0 {
		input.CustomerID = existing.CustomerID
	}
	if input.CustomerName == "" {
		input.CustomerName = existing.CustomerName
		input.CustomerCountry = existing.CustomerCountry
	}

	invDate := existing.InvoiceDate
	if input.InvoiceDate != "" {
		if t, err := time.Parse("2006-01-02", input.InvoiceDate); err == nil {
			invDate = t
		}
	}

	dueDate := existing.DueDate
	if input.DueDate != "" {
		if t, err := time.Parse("2006-01-02", input.DueDate); err == nil {
			dueDate = t
		}
	}

	daysDiff := int(math.Ceil(dueDate.Sub(time.Now()).Hours() / 24))
	daysLeft := fmt.Sprintf("%d days left", daysDiff)
	if daysDiff < 0 {
		daysLeft = "Overdue"
	}

	var subtotal float64
	var items []InvoiceItem
	for _, item := range input.LineItems {
		itemAmt := item.Quantity * item.UnitPrice
		subtotal += itemAmt
		items = append(items, InvoiceItem{
			Description:     item.Description,
			ServiceCategory: item.ServiceCategory,
			Quantity:        item.Quantity,
			UnitPrice:       item.UnitPrice,
			TotalAmount:     itemAmt,
		})
	}

	if input.Subtotal > 0 {
		subtotal = input.Subtotal
	}

	totalAmount := subtotal + input.TaxAmount - input.DiscountAmount
	if totalAmount < 0 {
		totalAmount = 0
	}

	existing.CustomerID = input.CustomerID
	if input.CustomerName != "" {
		existing.CustomerName = input.CustomerName
	}
	if input.CustomerCountry != "" {
		existing.CustomerCountry = input.CustomerCountry
	}
	if input.Route != "" {
		existing.Route = input.Route
	}
	existing.InvoiceDate = invDate
	existing.DueDate = dueDate
	existing.DaysLeft = daysLeft
	if input.Currency != "" {
		existing.Currency = input.Currency
	}
	existing.Subtotal = subtotal
	existing.TaxAmount = input.TaxAmount
	existing.DiscountAmount = input.DiscountAmount
	existing.TotalAmount = totalAmount
	existing.BalanceDue = totalAmount - existing.PaidAmount

	updated, err := s.repo.UpdateInvoice(ctx, existing, items)
	if err != nil {
		return nil, fmt.Errorf("failed to update draft invoice: %w", err)
	}

	_ = s.repo.AddHistory(ctx, orgID, id, "Draft Updated", "Draft invoice details and line items were updated", userName)
	return updated, nil
}

func (s *service) IssueInvoice(ctx context.Context, orgID int64, id int64, userName string) (*Invoice, error) {
	inv, err := s.repo.GetInvoiceByID(ctx, orgID, id)
	if err != nil {
		return nil, ErrInvoiceNotFound
	}

	if err := ValidateStatusTransition(inv.Status, "Issued"); err != nil {
		return nil, err
	}

	if inv.TotalAmount <= 0 && len(inv.LineItems) == 0 {
		return nil, errors.New("cannot issue an invoice without valid line items or total amount")
	}

	if err := s.repo.UpdateInvoiceStatus(ctx, orgID, id, "Issued"); err != nil {
		return nil, fmt.Errorf("failed to issue invoice: %w", err)
	}

	_ = s.repo.AddHistory(ctx, orgID, id, "Invoice Issued", "Invoice transitioned from Draft to Issued", userName)

	_, _ = audit.Record(ctx, domain.CreateAuditLogParams{
		OrgID:        orgID,
		ActorName:    userName,
		Action:       domain.ActionSend,
		Module:       domain.ModuleInvoices,
		ResourceType: "INVOICE",
		ResourceID:   fmt.Sprintf("%d", id),
		ResourceName: inv.InvoiceNumber,
		Description:  fmt.Sprintf("Issued invoice %s for customer %s", inv.InvoiceNumber, inv.CustomerName),
		Result:       domain.ResultSuccess,
	})

	return s.repo.GetInvoiceByID(ctx, orgID, id)
}

func (s *service) SubmitForApproval(ctx context.Context, orgID int64, id int64, userName string) (*Invoice, error) {
	inv, err := s.repo.GetInvoiceByID(ctx, orgID, id)
	if err != nil {
		return nil, ErrInvoiceNotFound
	}

	if err := ValidateStatusTransition(inv.Status, "Pending Approval"); err != nil {
		return nil, err
	}

	if err := s.repo.UpdateInvoiceStatus(ctx, orgID, id, "Pending Approval"); err != nil {
		return nil, fmt.Errorf("failed to submit invoice for approval: %w", err)
	}

	_ = s.repo.CreateApprovalRequestForInvoice(ctx, inv, userName)
	_ = s.repo.AddHistory(ctx, orgID, id, "Submitted for Approval", "Invoice submitted for manager approval", userName)
	return s.repo.GetInvoiceByID(ctx, orgID, id)
}

func (s *service) UpdateInvoiceStatus(ctx context.Context, orgID int64, id int64, status string, userName string) error {
	inv, err := s.repo.GetInvoiceByID(ctx, orgID, id)
	if err != nil {
		return ErrInvoiceNotFound
	}

	if err := ValidateStatusTransition(inv.Status, status); err != nil {
		return err
	}

	if err := s.repo.UpdateInvoiceStatus(ctx, orgID, id, status); err != nil {
		return err
	}

	_ = s.repo.AddHistory(ctx, orgID, id, fmt.Sprintf("Status updated to %s", status), fmt.Sprintf("Invoice status changed from %s to %s", inv.Status, status), userName)
	return nil
}

func (s *service) ToggleBookmark(ctx context.Context, orgID int64, id int64) (bool, error) {
	return s.repo.ToggleBookmark(ctx, orgID, id)
}

func (s *service) CancelInvoice(ctx context.Context, orgID int64, id int64, reason string, userName string) error {
	inv, err := s.repo.GetInvoiceByID(ctx, orgID, id)
	if err != nil {
		return ErrInvoiceNotFound
	}

	if err := ValidateStatusTransition(inv.Status, "Cancelled"); err != nil {
		return err
	}

	if err := s.repo.UpdateInvoiceStatus(ctx, orgID, id, "Cancelled"); err != nil {
		return err
	}

	desc := "Invoice has been cancelled"
	if reason != "" {
		desc = fmt.Sprintf("Invoice cancelled. Reason: %s", reason)
	}

	_ = s.repo.AddHistory(ctx, orgID, id, "Invoice Cancelled", desc, userName)

	_, _ = audit.Record(ctx, domain.CreateAuditLogParams{
		OrgID:        orgID,
		ActorName:    userName,
		Action:       domain.ActionDelete,
		Module:       domain.ModuleInvoices,
		ResourceType: "INVOICE",
		ResourceID:   fmt.Sprintf("%d", id),
		ResourceName: inv.InvoiceNumber,
		Description:  desc,
		Result:       domain.ResultSuccess,
	})

	return nil
}

func (s *service) RecordPayment(ctx context.Context, orgID int64, invoiceID int64, input RecordPaymentInput, userName string) (*Invoice, error) {
	inv, err := s.repo.GetInvoiceByID(ctx, orgID, invoiceID)
	if err != nil {
		return nil, ErrInvoiceNotFound
	}

	if inv.Status == "Draft" {
		return nil, fmt.Errorf("cannot record payments on draft invoices (issue invoice first)")
	}
	if inv.Status == "Cancelled" {
		return nil, fmt.Errorf("cannot record payments on cancelled invoices")
	}
	if inv.BalanceDue <= 0 || inv.Status == "Paid" {
		return nil, fmt.Errorf("invoice is already fully paid")
	}
	if input.Amount <= 0 {
		return nil, fmt.Errorf("payment amount must be greater than zero")
	}
	if input.Amount > inv.BalanceDue+0.01 {
		return nil, fmt.Errorf("payment amount ($%.2f) exceeds remaining balance due ($%.2f)", input.Amount, inv.BalanceDue)
	}

	payDate := time.Now()
	if input.PaymentDate != "" {
		if parsed, err := time.Parse("2006-01-02", input.PaymentDate); err == nil {
			payDate = parsed
		} else if parsed, err := time.Parse(time.RFC3339, input.PaymentDate); err == nil {
			payDate = parsed
		}
	}

	payRef := strings.TrimSpace(input.PaymentRef)
	if payRef == "" {
		payRef = fmt.Sprintf("PAY-%d-%04d", time.Now().Year(), rand.Intn(9000)+1000)
	}

	payMethod := strings.TrimSpace(input.PaymentMethod)
	if payMethod == "" {
		payMethod = "Wire Transfer"
	}

	newPaidAmount := inv.PaidAmount + input.Amount
	newBalanceDue := inv.TotalAmount - newPaidAmount
	if newBalanceDue < 0 {
		newBalanceDue = 0
	}

	var newStatus string
	if newBalanceDue <= 0.001 {
		newStatus = "Paid"
	} else {
		newStatus = "Partially Paid"
	}

	paymentRecord := &InvoicePayment{
		OrgID:         orgID,
		InvoiceID:     invoiceID,
		PaymentRef:    payRef,
		Amount:        input.Amount,
		PaymentMethod: payMethod,
		Status:        "Completed",
		PaymentDate:   payDate,
	}
	if input.Notes != "" {
		notes := input.Notes
		paymentRecord.Notes = &notes
	}

	var histTitle, histDesc string
	if newStatus == "Paid" {
		histTitle = "Invoice Paid in Full"
		histDesc = fmt.Sprintf("Recorded final payment of $%.2f via %s (Ref: %s). Balance is now $0.00", input.Amount, payMethod, payRef)
	} else {
		histTitle = "Partial Payment Received"
		histDesc = fmt.Sprintf("Recorded partial payment of $%.2f via %s (Ref: %s). Remaining balance: $%.2f", input.Amount, payMethod, payRef, newBalanceDue)
	}

	res, err := s.repo.RecordPayment(ctx, orgID, invoiceID, paymentRecord, newPaidAmount, newBalanceDue, newStatus, histTitle, histDesc, userName)
	if err != nil {
		return nil, err
	}

	_, _ = audit.Record(ctx, domain.CreateAuditLogParams{
		OrgID:        orgID,
		ActorName:    userName,
		Action:       domain.ActionCreate,
		Module:       domain.ModulePayments,
		ResourceType: "PAYMENT",
		ResourceID:   fmt.Sprintf("%d", invoiceID),
		ResourceName: inv.InvoiceNumber,
		Description:  fmt.Sprintf("Recorded payment of %s %.2f for invoice %s", inv.Currency, input.Amount, inv.InvoiceNumber),
		After: map[string]interface{}{
			"amount":         input.Amount,
			"payment_method": payMethod,
			"payment_ref":    payRef,
		},
		Result: domain.ResultSuccess,
	})

	return res, nil
}

func (s *service) GetAllPayments(ctx context.Context, orgID int64) ([]InvoicePayment, error) {
	return s.repo.GetAllPayments(ctx, orgID)
}

func (s *service) AddDocument(ctx context.Context, orgID int64, invoiceID int64, docName, fileSize, fileType, s3Key string) (*InvoiceDocument, error) {
	inv, err := s.repo.GetInvoiceByID(ctx, orgID, invoiceID)
	if err != nil {
		return nil, ErrInvoiceNotFound
	}
	_ = inv

	if docName == "" {
		docName = "Payment Receipt.pdf"
	}
	if fileSize == "" {
		fileSize = "250 KB"
	}
	if fileType == "" {
		fileType = "application/pdf"
	}

	doc := &InvoiceDocument{
		OrgID:        orgID,
		InvoiceID:    invoiceID,
		DocumentName: docName,
		FileSize:     fileSize,
		FileType:     fileType,
		S3Key:        &s3Key,
	}

	if err := s.repo.AddDocument(ctx, orgID, invoiceID, doc); err != nil {
		return nil, err
	}

	_ = s.repo.AddHistory(ctx, orgID, invoiceID, "Document Attached", fmt.Sprintf("Attached supporting document '%s'", docName), "System")

	return doc, nil
}
