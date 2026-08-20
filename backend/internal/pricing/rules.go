package pricing

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// PricingRule represents a markup rate configuration.
type PricingRule struct {
	ID           int64           `db:"id" json:"id"`
	OrgID        int64           `db:"org_id" json:"org_id"`
	RuleName     string          `db:"rule_name" json:"rule_name"`
	RuleType     string          `db:"rule_type" json:"rule_type"` // DEFAULT | LANE | CUSTOMER_TIER
	Conditions   json.RawMessage `db:"conditions" json:"conditions"`
	MarkupPct    *float64        `db:"markup_pct" json:"markup_pct"`
	MarkupFlat   *float64        `db:"markup_flat" json:"markup_flat"`
	MinMarginPct float64         `db:"min_margin_pct" json:"min_margin_pct"`
	Priority     int             `db:"priority" json:"priority"`
	IsActive     bool            `db:"is_active" json:"is_active"`
}

// Service is the pricing rules service interface.
type Service interface {
	GetPricingRules(ctx context.Context, orgID int64) ([]PricingRule, error)
	GetApplicableRules(ctx context.Context, orgID int64, origin, destination, customerTier, equipment string) ([]PricingRule, error)
}

type service struct {
	db *sqlx.DB
}

// NewService creates a new pricing rules service.
func NewService(db *sqlx.DB) Service {
	return &service{db: db}
}

func (s *service) GetPricingRules(ctx context.Context, orgID int64) ([]PricingRule, error) {
	var rules []PricingRule
	query := `SELECT id, org_id, rule_name, rule_type, conditions, markup_pct, markup_flat, min_margin_pct, priority, is_active 
	          FROM pricing_rules WHERE org_id = ? AND is_active = 1 ORDER BY priority DESC`
	err := s.db.SelectContext(ctx, &rules, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("select pricing rules: %w", err)
	}
	return rules, nil
}

func (s *service) GetApplicableRules(ctx context.Context, orgID int64, origin, destination, customerTier, equipment string) ([]PricingRule, error) {
	allRules, err := s.GetPricingRules(ctx, orgID)
	if err != nil {
		return nil, err
	}

	var applicable []PricingRule
	for _, rule := range allRules {
		matches := true
		if len(rule.Conditions) > 2 { // Not empty "{}"
			var conds map[string]string
			if err := json.Unmarshal(rule.Conditions, &conds); err == nil {
				if val, ok := conds["origin"]; ok && val != origin {
					matches = false
				}
				if val, ok := conds["destination"]; ok && val != destination {
					matches = false
				}
				if val, ok := conds["tier"]; ok && val != customerTier {
					matches = false
				}
				if val, ok := conds["equipment"]; ok && val != equipment {
					matches = false
				}
			}
		}
		if matches {
			applicable = append(applicable, rule)
		}
	}
	return applicable, nil
}
