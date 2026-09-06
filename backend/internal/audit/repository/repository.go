package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/freel/backend/internal/audit/domain"
	"github.com/jmoiron/sqlx"
)

// Repository defines the data-access interface for universal audit logs.
type Repository interface {
	Create(ctx context.Context, log *domain.AuditLog) error
	List(ctx context.Context, filter domain.AuditLogFilter) ([]domain.AuditLog, int64, error)
	GetByID(ctx context.Context, orgID int64, id int64) (*domain.AuditLog, error)
}

type mysqlRepository struct {
	db *sqlx.DB
}

// NewMySQLRepository creates a new instance of Repository using MySQL.
func NewMySQLRepository(db *sqlx.DB) Repository {
	return &mysqlRepository{db: db}
}

// dbAuditLogRow represents the raw database row structure.
type dbAuditLogRow struct {
	ID           int64          `db:"id"`
	OrgID        int64          `db:"org_id"`
	UserID       sql.NullInt64  `db:"user_id"`
	ActorType    sql.NullString `db:"actor_type"`
	ActorName    sql.NullString `db:"actor_name"`
	ActorRole    sql.NullString `db:"actor_role"`
	Action       string         `db:"action"`
	Module       sql.NullString `db:"module"`
	ResourceType string         `db:"resource_type"`
	ResourceID   string         `db:"resource_id"`
	ResourceName sql.NullString `db:"resource_name"`
	Description  sql.NullString `db:"description"`
	Result       sql.NullString `db:"result"`
	ErrorMessage sql.NullString `db:"error_message"`
	BeforeData   sql.NullString `db:"before_data"`
	AfterData    sql.NullString `db:"after_data"`
	Changes      sql.NullString `db:"changes"`
	Metadata     sql.NullString `db:"metadata"`
	IPAddress    sql.NullString `db:"ip_address"`
	UserAgent    sql.NullString `db:"user_agent"`
	CreatedAt    time.Time      `db:"created_at"`
}

func (r *mysqlRepository) toDomain(row dbAuditLogRow) domain.AuditLog {
	actorType := domain.ActorTypeUser
	if row.ActorType.Valid && row.ActorType.String != "" {
		actorType = row.ActorType.String
	}

	module := domain.ModuleSettings
	if row.Module.Valid && row.Module.String != "" {
		module = row.Module.String
	}

	result := domain.ResultSuccess
	if row.Result.Valid && row.Result.String != "" {
		result = row.Result.String
	}

	description := fmt.Sprintf("%s %s", row.Action, row.ResourceType)
	if row.Description.Valid && row.Description.String != "" {
		description = row.Description.String
	}

	item := domain.AuditLog{
		ID:           row.ID,
		OrgID:        row.OrgID,
		ActorType:    actorType,
		Action:       row.Action,
		Module:       module,
		ResourceType: row.ResourceType,
		ResourceID:   row.ResourceID,
		Description:  description,
		Result:       result,
		CreatedAt:    row.CreatedAt,
	}

	if row.UserID.Valid {
		uid := row.UserID.Int64
		item.ActorID = &uid
	}
	if row.ActorName.Valid {
		item.ActorName = row.ActorName.String
	}
	if row.ActorRole.Valid {
		item.ActorRole = row.ActorRole.String
	}
	if row.ResourceName.Valid {
		item.ResourceName = row.ResourceName.String
	}
	if row.ErrorMessage.Valid {
		item.ErrorMessage = row.ErrorMessage.String
	}
	if row.IPAddress.Valid {
		item.IPAddress = row.IPAddress.String
	}
	if row.UserAgent.Valid {
		item.UserAgent = row.UserAgent.String
	}

	// JSON unmarshaling
	if row.BeforeData.Valid && row.BeforeData.String != "" {
		_ = json.Unmarshal([]byte(row.BeforeData.String), &item.BeforeData)
	}
	if row.AfterData.Valid && row.AfterData.String != "" {
		_ = json.Unmarshal([]byte(row.AfterData.String), &item.AfterData)
	}
	if row.Changes.Valid && row.Changes.String != "" {
		_ = json.Unmarshal([]byte(row.Changes.String), &item.Changes)
	}
	if row.Metadata.Valid && row.Metadata.String != "" {
		_ = json.Unmarshal([]byte(row.Metadata.String), &item.Metadata)
	}

	return item
}

func (r *mysqlRepository) Create(ctx context.Context, log *domain.AuditLog) error {
	if log.OrgID <= 0 {
		return errors.New("audit log must belong to a valid organization")
	}

	var beforeJSON, afterJSON, changesJSON, metadataJSON *string

	if log.BeforeData != nil {
		if b, err := json.Marshal(log.BeforeData); err == nil {
			s := string(b)
			beforeJSON = &s
		}
	}
	if log.AfterData != nil {
		if b, err := json.Marshal(log.AfterData); err == nil {
			s := string(b)
			afterJSON = &s
		}
	}
	if log.Changes != nil && len(log.Changes) > 0 {
		if b, err := json.Marshal(log.Changes); err == nil {
			s := string(b)
			changesJSON = &s
		}
	}
	if log.Metadata != nil {
		if b, err := json.Marshal(log.Metadata); err == nil {
			s := string(b)
			metadataJSON = &s
		}
	}

	query := `
		INSERT INTO audit_logs (
			org_id, user_id, actor_type, actor_name, actor_role,
			action, module, resource_type, resource_id, resource_name,
			description, result, error_message, before_data, after_data,
			changes, metadata, ip_address, user_agent, created_at
		) VALUES (
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?
		)
	`

	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now().UTC()
	}

	res, err := r.db.ExecContext(ctx, query,
		log.OrgID,
		log.ActorID,
		log.ActorType,
		log.ActorName,
		log.ActorRole,
		log.Action,
		log.Module,
		log.ResourceType,
		log.ResourceID,
		log.ResourceName,
		log.Description,
		log.Result,
		log.ErrorMessage,
		beforeJSON,
		afterJSON,
		changesJSON,
		metadataJSON,
		log.IPAddress,
		log.UserAgent,
		log.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert audit log: %w", err)
	}

	id, err := res.LastInsertId()
	if err == nil {
		log.ID = id
	}
	return nil
}

func (r *mysqlRepository) List(ctx context.Context, filter domain.AuditLogFilter) ([]domain.AuditLog, int64, error) {
	if filter.OrgID <= 0 {
		return nil, 0, errors.New("organization ID is required for audit logs list")
	}

	var conditions []string
	var args []interface{}

	// Organization isolation constraint
	conditions = append(conditions, "org_id = ?")
	args = append(args, filter.OrgID)

	if filter.ActorID != nil && *filter.ActorID > 0 {
		conditions = append(conditions, "user_id = ?")
		args = append(args, *filter.ActorID)
	}

	if filter.ActorType != "" && filter.ActorType != "ALL" {
		conditions = append(conditions, "actor_type = ?")
		args = append(args, filter.ActorType)
	}

	if filter.Module != "" && filter.Module != "ALL" {
		conditions = append(conditions, "module = ?")
		args = append(args, filter.Module)
	}

	if filter.Action != "" && filter.Action != "ALL" {
		conditions = append(conditions, "action = ?")
		args = append(args, filter.Action)
	}

	if filter.ResourceType != "" && filter.ResourceType != "ALL" {
		conditions = append(conditions, "resource_type = ?")
		args = append(args, filter.ResourceType)
	}

	if filter.ResourceID != "" {
		conditions = append(conditions, "resource_id = ?")
		args = append(args, filter.ResourceID)
	}

	if filter.Result != "" && filter.Result != "ALL" {
		conditions = append(conditions, "result = ?")
		args = append(args, filter.Result)
	}

	if filter.StartDate != nil && !filter.StartDate.IsZero() {
		conditions = append(conditions, "created_at >= ?")
		args = append(args, *filter.StartDate)
	}

	if filter.EndDate != nil && !filter.EndDate.IsZero() {
		conditions = append(conditions, "created_at <= ?")
		args = append(args, *filter.EndDate)
	}

	if strings.TrimSpace(filter.Search) != "" {
		searchTerm := "%" + strings.TrimSpace(filter.Search) + "%"
		conditions = append(conditions, "(description LIKE ? OR resource_name LIKE ? OR resource_id LIKE ? OR actor_name LIKE ? OR action LIKE ? OR module LIKE ?)")
		args = append(args, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm)
	}

	whereClause := strings.Join(conditions, " AND ")

	// Count Total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM audit_logs WHERE %s", whereClause)
	var total int64
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, fmt.Errorf("failed to count audit logs: %w", err)
	}

	// Pagination Defaults
	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	} else if limit > 100 {
		limit = 100
	}

	page := filter.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	// Select Query - Newest First (created_at DESC, id DESC)
	selectQuery := fmt.Sprintf(`
		SELECT id, org_id, user_id, actor_type, actor_name, actor_role,
		       action, module, resource_type, resource_id, resource_name,
		       description, result, error_message, before_data, after_data,
		       changes, metadata, ip_address, user_agent, created_at
		FROM audit_logs
		WHERE %s
		ORDER BY id DESC
		LIMIT ? OFFSET ?
	`, whereClause)

	queryArgs := append(args, limit, offset)

	var rows []dbAuditLogRow
	if err := r.db.SelectContext(ctx, &rows, selectQuery, queryArgs...); err != nil {
		return nil, 0, fmt.Errorf("failed to fetch audit logs: %w", err)
	}

	items := make([]domain.AuditLog, len(rows))
	for i, rRow := range rows {
		items[i] = r.toDomain(rRow)
	}

	return items, total, nil
}

func (r *mysqlRepository) GetByID(ctx context.Context, orgID int64, id int64) (*domain.AuditLog, error) {
	if orgID <= 0 || id <= 0 {
		return nil, errors.New("invalid organization ID or log ID")
	}

	query := `
		SELECT id, org_id, user_id, actor_type, actor_name, actor_role,
		       action, module, resource_type, resource_id, resource_name,
		       description, result, error_message, before_data, after_data,
		       changes, metadata, ip_address, user_agent, created_at
		FROM audit_logs
		WHERE org_id = ? AND id = ?
		LIMIT 1
	`

	var row dbAuditLogRow
	if err := r.db.GetContext(ctx, &row, query, orgID, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("audit log not found")
		}
		return nil, fmt.Errorf("failed to retrieve audit log: %w", err)
	}

	item := r.toDomain(row)
	return &item, nil
}
