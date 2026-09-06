package approvals

import (
	"time"
)

type ApprovalRequest struct {
	ID                int64      `db:"id" json:"id"`
	OrgID             int64      `db:"org_id" json:"org_id"`
	RequestCode       string     `db:"request_code" json:"request_code"`
	Title             string     `db:"title" json:"title"`
	Category          string     `db:"category" json:"category"`   // DOCUMENTS, COMMERCIAL, OPERATIONS, FINANCE
	Type              string     `db:"type" json:"type"`           // Document Approval, Commercial Approval, Operations Approval, Finance Approval
	Status            string     `db:"status" json:"status"`       // Pending, In Review, Approved, Rejected, Overdue, Cancelled
	Priority          string     `db:"priority" json:"priority"`   // LOW, MEDIUM, HIGH, URGENT
	RelatedEntityType *string    `db:"related_entity_type" json:"related_entity_type,omitempty"`
	RelatedEntityID   *int64     `db:"related_entity_id" json:"related_entity_id,omitempty"`
	RelatedRef        *string    `db:"related_ref" json:"related_ref,omitempty"`
	CustomerName      *string    `db:"customer_name" json:"customer_name,omitempty"`
	CustomerID        *int64     `db:"customer_id" json:"customer_id,omitempty"`
	ShipmentID        *int64     `db:"shipment_id" json:"shipment_id,omitempty"`
	DocumentID        *int64     `db:"document_id" json:"document_id,omitempty"`
	BookingID         *int64     `db:"booking_id" json:"booking_id,omitempty"`
	RequestedByID     *int64     `db:"requested_by_id" json:"requested_by_id,omitempty"`
	RequestedByName   string     `db:"requested_by_name" json:"requested_by_name"`
	Department        *string    `db:"department" json:"department,omitempty"`
	Avatar            *string    `db:"avatar" json:"avatar,omitempty"`
	DueDate           *time.Time `db:"due_date" json:"due_date,omitempty"`
	DueText           *string    `db:"due_text" json:"due_text,omitempty"`
	AssignedTo        *string    `db:"assigned_to" json:"assigned_to,omitempty"`
	ApprovedBy        *string    `db:"approved_by" json:"approved_by,omitempty"`
	ApprovedAt        *time.Time `db:"approved_at" json:"approved_at,omitempty"`
	RejectedBy        *string    `db:"rejected_by" json:"rejected_by,omitempty"`
	RejectedAt        *time.Time `db:"rejected_at" json:"rejected_at,omitempty"`
	RejectionReason   *string    `db:"rejection_reason" json:"rejection_reason,omitempty"`
	Comments          *string    `db:"comments" json:"comments,omitempty"`
	Description       *string    `db:"description" json:"description,omitempty"`
	CreatedAt         time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time  `db:"updated_at" json:"updated_at"`
}

type ApprovalStats struct {
	Pending       int    `json:"pending"`
	PendingTrend  string `json:"pending_trend"`
	Approved      int    `json:"approved"`
	ApprovedTrend string `json:"approved_trend"`
	Rejected      int    `json:"rejected"`
	RejectedTrend string `json:"rejected_trend"`
	Overdue       int    `json:"overdue"`
}

type CreateApprovalInput struct {
	Title             string  `json:"title"`
	Category          string  `json:"category"`
	Type              string  `json:"type"`
	Priority          string  `json:"priority"`
	RelatedRef        string  `json:"related_ref"`
	RelatedEntityType string  `json:"related_entity_type"`
	RelatedEntityID   int64   `json:"related_entity_id"`
	CustomerName      string  `json:"customer_name"`
	CustomerID        int64   `json:"customer_id"`
	ShipmentID        int64   `json:"shipment_id"`
	DocumentID        int64   `json:"document_id"`
	BookingID         int64   `json:"booking_id"`
	RequestedByName   string  `json:"requested_by_name"`
	Department        string  `json:"department"`
	DueDate           string  `json:"due_date"`
	Description       string  `json:"description"`
}

type ActionApprovalInput struct {
	Action string `json:"action"` // APPROVE, REJECT
	Reason string `json:"reason"`
	Notes  string `json:"notes"`
}
