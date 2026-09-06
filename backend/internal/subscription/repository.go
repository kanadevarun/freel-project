package subscription

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type Repository interface {
	GetAvailablePlans(ctx context.Context) ([]SubscriptionPlan, error)
	GetPlanByID(ctx context.Context, planID int64) (*SubscriptionPlan, error)
	GetSubscriptionByOrgID(ctx context.Context, orgID int64) (*OrganizationSubscription, error)
	GetCustomerByOrgID(ctx context.Context, orgID int64) (*BillingCustomer, error)
	GetPaymentMethodByOrgID(ctx context.Context, orgID int64) (*BillingPaymentMethod, error)
	GetUsageByOrgID(ctx context.Context, orgID int64) ([]SubscriptionUsage, error)
	GetUsageByMetric(ctx context.Context, orgID int64, metricName string) (*SubscriptionUsage, error)
	GetAddonsByOrgID(ctx context.Context, orgID int64) ([]SubscriptionAddon, error)
	GetInvoicesByOrgID(ctx context.Context, orgID int64) ([]Invoice, error)
	UpsertSubscription(ctx context.Context, sub *OrganizationSubscription) error
	CancelSubscription(ctx context.Context, orgID int64) error
	UpsertCustomer(ctx context.Context, customer *BillingCustomer) error
	IncrementUsage(ctx context.Context, orgID int64, metricName string, amount int) error
	UpsertUsage(ctx context.Context, usage *SubscriptionUsage) error
	UpsertInvoice(ctx context.Context, invoice *Invoice) error
	GetAddonConfigs(ctx context.Context) ([]AddonConfig, error)
	GetAddonConfigByID(ctx context.Context, id int64) (*AddonConfig, error)
	UpsertSubscriptionAddon(ctx context.Context, addon *SubscriptionAddon) error
	DeleteSubscriptionAddon(ctx context.Context, orgID int64, addonConfigID int64) error
	UpsertPaymentMethod(ctx context.Context, pm *BillingPaymentMethod) error
}

type sqlxRepository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &sqlxRepository{db: db}
}

func (r *sqlxRepository) GetAvailablePlans(ctx context.Context) ([]SubscriptionPlan, error) {
	var plans []SubscriptionPlan
	query := `SELECT id, name, description, price_monthly, price_annual, 
			features, limits, provider_product_id, 
			provider_price_id_monthly, provider_price_id_annual,
			is_active, created_at, updated_at
		FROM subscription_plans WHERE is_active = true ORDER BY price_monthly ASC`
	err := r.db.SelectContext(ctx, &plans, query)
	if err != nil {
		return nil, err
	}
	return plans, nil
}

func (r *sqlxRepository) GetPlanByID(ctx context.Context, planID int64) (*SubscriptionPlan, error) {
	var plan SubscriptionPlan
	query := `SELECT id, name, description, price_monthly, price_annual, 
			features, limits, provider_product_id, 
			provider_price_id_monthly, provider_price_id_annual,
			is_active, created_at, updated_at
		FROM subscription_plans WHERE id = ?`
	err := r.db.GetContext(ctx, &plan, query, planID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

func (r *sqlxRepository) GetSubscriptionByOrgID(ctx context.Context, orgID int64) (*OrganizationSubscription, error) {
	var sub OrganizationSubscription
	query := `SELECT * FROM organization_subscriptions WHERE org_id = ?`
	err := r.db.GetContext(ctx, &sub, query, orgID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

func (r *sqlxRepository) GetCustomerByOrgID(ctx context.Context, orgID int64) (*BillingCustomer, error) {
	var customer BillingCustomer
	query := `SELECT * FROM billing_customers WHERE org_id = ?`
	err := r.db.GetContext(ctx, &customer, query, orgID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

func (r *sqlxRepository) GetPaymentMethodByOrgID(ctx context.Context, orgID int64) (*BillingPaymentMethod, error) {
	var pm BillingPaymentMethod
	query := `SELECT * FROM billing_payment_methods WHERE org_id = ? AND is_default = true LIMIT 1`
	err := r.db.GetContext(ctx, &pm, query, orgID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &pm, nil
}

func (r *sqlxRepository) GetUsageByOrgID(ctx context.Context, orgID int64) ([]SubscriptionUsage, error) {
	var usages []SubscriptionUsage
	query := `SELECT * FROM subscription_usage WHERE org_id = ? ORDER BY metric_name ASC`
	err := r.db.SelectContext(ctx, &usages, query, orgID)
	if err != nil {
		return nil, err
	}
	return usages, nil
}

func (r *sqlxRepository) GetAddonsByOrgID(ctx context.Context, orgID int64) ([]SubscriptionAddon, error) {
	var addons []SubscriptionAddon
	query := `SELECT * FROM subscription_addons WHERE org_id = ? ORDER BY addon_name ASC`
	err := r.db.SelectContext(ctx, &addons, query, orgID)
	if err != nil {
		return nil, err
	}
	return addons, nil
}

func (r *sqlxRepository) GetInvoicesByOrgID(ctx context.Context, orgID int64) ([]Invoice, error) {
	var invoices []Invoice
	query := `SELECT * FROM invoices WHERE org_id = ? ORDER BY issued_at DESC`
	err := r.db.SelectContext(ctx, &invoices, query, orgID)
	if err != nil {
		return nil, err
	}
	return invoices, nil
}

func (r *sqlxRepository) UpsertSubscription(ctx context.Context, sub *OrganizationSubscription) error {
	query := `
		INSERT INTO organization_subscriptions (
			org_id, plan_id, status, billing_cycle, 
			current_period_start, current_period_end, 
			cancel_at_period_end, provider_subscription_id
		) VALUES (
			:org_id, :plan_id, :status, :billing_cycle, 
			:current_period_start, :current_period_end, 
			:cancel_at_period_end, :provider_subscription_id
		)
		ON DUPLICATE KEY UPDATE
			plan_id = :plan_id,
			status = :status,
			billing_cycle = :billing_cycle,
			current_period_start = :current_period_start,
			current_period_end = :current_period_end,
			cancel_at_period_end = :cancel_at_period_end,
			provider_subscription_id = :provider_subscription_id,
			updated_at = CURRENT_TIMESTAMP
	`
	_, err := r.db.NamedExecContext(ctx, query, sub)
	return err
}

func (r *sqlxRepository) CancelSubscription(ctx context.Context, orgID int64) error {
	query := `
		UPDATE organization_subscriptions 
		SET cancel_at_period_end = true, updated_at = CURRENT_TIMESTAMP 
		WHERE org_id = ?
	`
	_, err := r.db.ExecContext(ctx, query, orgID)
	return err
}

func (r *sqlxRepository) UpsertCustomer(ctx context.Context, c *BillingCustomer) error {
	query := `
		INSERT INTO billing_customers (org_id, provider_customer_id, created_at, updated_at)
		VALUES (:org_id, :provider_customer_id, :created_at, :updated_at)
		ON DUPLICATE KEY UPDATE
			provider_customer_id = VALUES(provider_customer_id),
			updated_at = VALUES(updated_at)
	`
	res, err := r.db.NamedExecContext(ctx, query, c)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err == nil && id > 0 {
		c.ID = id
	}
	return nil
}

func (r *sqlxRepository) GetUsageByMetric(ctx context.Context, orgID int64, metricName string) (*SubscriptionUsage, error) {
	var usage SubscriptionUsage
	query := `SELECT * FROM subscription_usage WHERE org_id = ? AND metric_name = ? LIMIT 1`
	err := r.db.GetContext(ctx, &usage, query, orgID, metricName)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &usage, nil
}

func (r *sqlxRepository) IncrementUsage(ctx context.Context, orgID int64, metricName string, amount int) error {
	query := `
		UPDATE subscription_usage 
		SET current_usage = current_usage + ?, updated_at = CURRENT_TIMESTAMP 
		WHERE org_id = ? AND metric_name = ?
	`
	_, err := r.db.ExecContext(ctx, query, amount, orgID, metricName)
	return err
}

func (r *sqlxRepository) UpsertUsage(ctx context.Context, u *SubscriptionUsage) error {
	query := `
		INSERT INTO subscription_usage (
			org_id, metric_name, current_usage, limit_amount, period_start, period_end
		) VALUES (
			:org_id, :metric_name, :current_usage, :limit_amount, :period_start, :period_end
		)
		ON DUPLICATE KEY UPDATE
			current_usage = :current_usage,
			limit_amount = :limit_amount,
			period_start = :period_start,
			period_end = :period_end,
			updated_at = CURRENT_TIMESTAMP
	`
	res, err := r.db.NamedExecContext(ctx, query, u)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err == nil && id > 0 {
		u.ID = id
	}
	return nil
}

func (r *sqlxRepository) UpsertInvoice(ctx context.Context, inv *Invoice) error {
	query := `
		INSERT INTO invoices 
			(org_id, provider_invoice_id, amount_due, status, invoice_pdf_url, issued_at, created_at, updated_at)
		VALUES 
			(?, ?, ?, ?, ?, ?, NOW(), NOW())
		ON DUPLICATE KEY UPDATE 
			status = VALUES(status),
			amount_due = VALUES(amount_due),
			invoice_pdf_url = VALUES(invoice_pdf_url),
			updated_at = NOW()
	`
	_, err := r.db.ExecContext(ctx, query,
		inv.OrgID,
		inv.ProviderInvoiceID,
		inv.AmountDue,
		inv.Status,
		inv.InvoicePdfUrl,
		inv.IssuedAt,
	)
	return err
}

func (r *sqlxRepository) GetAddonConfigs(ctx context.Context) ([]AddonConfig, error) {
	var configs []AddonConfig
	query := `SELECT * FROM addon_configs ORDER BY id ASC`
	err := r.db.SelectContext(ctx, &configs, query)
	if err != nil {
		return nil, err
	}
	return configs, nil
}

func (r *sqlxRepository) GetAddonConfigByID(ctx context.Context, id int64) (*AddonConfig, error) {
	var config AddonConfig
	query := `SELECT * FROM addon_configs WHERE id = ?`
	err := r.db.GetContext(ctx, &config, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *sqlxRepository) UpsertSubscriptionAddon(ctx context.Context, addon *SubscriptionAddon) error {
	query := `
		INSERT INTO subscription_addons (
			org_id, addon_config_id, addon_name, quantity, price_per_unit, provider_subscription_item_id
		) VALUES (
			:org_id, :addon_config_id, :addon_name, :quantity, :price_per_unit, :provider_subscription_item_id
		)
		ON DUPLICATE KEY UPDATE
			quantity = :quantity,
			price_per_unit = :price_per_unit,
			provider_subscription_item_id = :provider_subscription_item_id,
			updated_at = CURRENT_TIMESTAMP
	`
	res, err := r.db.NamedExecContext(ctx, query, addon)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err == nil && id > 0 {
		addon.ID = id
	}
	return nil
}

func (r *sqlxRepository) DeleteSubscriptionAddon(ctx context.Context, orgID int64, addonConfigID int64) error {
	query := `DELETE FROM subscription_addons WHERE org_id = ? AND addon_config_id = ?`
	_, err := r.db.ExecContext(ctx, query, orgID, addonConfigID)
	return err
}

func (r *sqlxRepository) UpsertPaymentMethod(ctx context.Context, pm *BillingPaymentMethod) error {
	query := `
		INSERT INTO billing_payment_methods (
			org_id, provider_payment_method_id, card_brand, card_last4, exp_month, exp_year, is_default, created_at, updated_at
		) VALUES (
			:org_id, :provider_payment_method_id, :card_brand, :card_last4, :exp_month, :exp_year, :is_default, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
		)
		ON DUPLICATE KEY UPDATE
			card_brand = :card_brand,
			card_last4 = :card_last4,
			exp_month = :exp_month,
			exp_year = :exp_year,
			is_default = :is_default,
			updated_at = CURRENT_TIMESTAMP
	`
	_, err := r.db.NamedExecContext(ctx, query, pm)
	return err
}


