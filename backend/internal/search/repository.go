package search

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

type Repository interface {
	SearchAll(ctx context.Context, orgID int64, query string, entityType string, limit int) ([]SearchItem, error)
}

type mysqlRepository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &mysqlRepository{db: db}
}

func (r *mysqlRepository) SearchAll(ctx context.Context, orgID int64, query string, entityType string, limit int) ([]SearchItem, error) {
	if limit <= 0 {
		limit = 30
	}
	items := make([]SearchItem, 0)
	pattern := "%" + strings.TrimSpace(query) + "%"

	perCategoryLimit := 6
	filterCat := strings.ToUpper(strings.TrimSpace(entityType))

	// 1. SHIPMENTS
	if filterCat == "" || filterCat == "ALL" || filterCat == "SHIPMENT" || filterCat == "SHIPMENTS" {
		shipQuery := `
			SELECT id, COALESCE(shipment_number, '') AS num, COALESCE(booking_number, '') AS bnum,
			       COALESCE(carrier_scac, '') AS scac, COALESCE(origin_port, '') AS orig,
			       COALESCE(destination_port, '') AS dest, COALESCE(status, 'BOOKED') AS status,
			       created_at
			FROM shipments
			WHERE org_id = ? AND (
				shipment_number LIKE ? OR
				booking_number LIKE ? OR
				mbl_number LIKE ? OR
				hbl_number LIKE ? OR
				carrier_scac LIKE ? OR
				origin_port LIKE ? OR
				destination_port LIKE ? OR
				vessel_name LIKE ?
			)
			ORDER BY id DESC LIMIT ?`
		rows, err := r.db.QueryContext(ctx, shipQuery, orgID, pattern, pattern, pattern, pattern, pattern, pattern, pattern, pattern, perCategoryLimit)
		if err == nil {
			for rows.Next() {
				var id int64
				var num, bnum, scac, orig, dest, status string
				var createdAt sql.NullTime
				if err := rows.Scan(&id, &num, &bnum, &scac, &orig, &dest, &status, &createdAt); err == nil {
					title := num
					if title == "" {
						title = fmt.Sprintf("Shipment #%d", id)
					}
					sub := fmt.Sprintf("%s ➔ %s", orig, dest)
					if scac != "" {
						sub = fmt.Sprintf("%s • %s", scac, sub)
					}
					if bnum != "" {
						sub = fmt.Sprintf("%s (Booking: %s)", sub, bnum)
					}
					var t *time.Time
					if createdAt.Valid {
						t = &createdAt.Time
					}
					badgeType := "info"
					if strings.Contains(strings.ToUpper(status), "DELIVER") {
						badgeType = "success"
					} else if strings.Contains(strings.ToUpper(status), "CANCEL") {
						badgeType = "danger"
					}
					items = append(items, SearchItem{
						ID:        fmt.Sprintf("shipment-%d", id),
						Category:  CategoryShipment,
						Title:     title,
						Subtitle:  sub,
						Badge:     status,
						BadgeType: badgeType,
						URL:       fmt.Sprintf("/dashboard/shipments/%d", id),
						CreatedAt: t,
					})
				}
			}
			rows.Close()
		}
	}

	// 2. BOOKINGS
	if filterCat == "" || filterCat == "ALL" || filterCat == "BOOKING" || filterCat == "BOOKINGS" {
		bkQuery := `
			SELECT id, COALESCE(booking_number, '') AS num, COALESCE(carrier_name, '') AS carrier,
			       COALESCE(carrier_scac, '') AS scac, COALESCE(origin_port, '') AS orig,
			       COALESCE(destination_port, '') AS dest, COALESCE(status, 'DRAFT') AS status,
			       created_at
			FROM bookings
			WHERE org_id = ? AND (
				booking_number LIKE ? OR
				carrier_name LIKE ? OR
				carrier_scac LIKE ? OR
				carrier_booking_reference LIKE ? OR
				origin_port LIKE ? OR
				destination_port LIKE ? OR
				vessel_name LIKE ?
			)
			ORDER BY id DESC LIMIT ?`
		rows, err := r.db.QueryContext(ctx, bkQuery, orgID, pattern, pattern, pattern, pattern, pattern, pattern, pattern, perCategoryLimit)
		if err == nil {
			for rows.Next() {
				var id int64
				var num, carrier, scac, orig, dest, status string
				var createdAt sql.NullTime
				if err := rows.Scan(&id, &num, &carrier, &scac, &orig, &dest, &status, &createdAt); err == nil {
					sub := fmt.Sprintf("%s ➔ %s", orig, dest)
					if carrier != "" {
						sub = fmt.Sprintf("%s • %s", carrier, sub)
					}
					var t *time.Time
					if createdAt.Valid {
						t = &createdAt.Time
					}
					badgeType := "info"
					if strings.ToUpper(status) == "CONFIRMED" || strings.ToUpper(status) == "ALLOCATED" {
						badgeType = "success"
					}
					items = append(items, SearchItem{
						ID:        fmt.Sprintf("booking-%d", id),
						Category:  CategoryBooking,
						Title:     num,
						Subtitle:  sub,
						Badge:     status,
						BadgeType: badgeType,
						URL:       fmt.Sprintf("/dashboard/bookings/%d", id),
						CreatedAt: t,
					})
				}
			}
			rows.Close()
		}
	}

	// 3. RFQS
	if filterCat == "" || filterCat == "ALL" || filterCat == "RFQ" || filterCat == "RFQS" {
		rfqQuery := `
			SELECT r.id, COALESCE(r.rfq_number, '') AS num, COALESCE(r.origin, '') AS orig,
			       COALESCE(r.destination, '') AS dest, COALESCE(r.stage, r.status, 'DRAFT') AS stage,
			       COALESCE(c.name, '') AS cust_name, r.created_at
			FROM rfqs r
			LEFT JOIN customers cust ON r.customer_id = cust.id
			LEFT JOIN companies c ON cust.company_id = c.id
			WHERE r.org_id = ? AND (
				r.rfq_number LIKE ? OR
				r.origin LIKE ? OR
				r.destination LIKE ? OR
				r.incoterms LIKE ? OR
				c.name LIKE ?
			)
			ORDER BY r.id DESC LIMIT ?`
		rows, err := r.db.QueryContext(ctx, rfqQuery, orgID, pattern, pattern, pattern, pattern, pattern, perCategoryLimit)
		if err == nil {
			for rows.Next() {
				var id int64
				var num, orig, dest, stage, custName string
				var createdAt sql.NullTime
				if err := rows.Scan(&id, &num, &orig, &dest, &stage, &custName, &createdAt); err == nil {
					sub := fmt.Sprintf("%s ➔ %s", orig, dest)
					if custName != "" {
						sub = fmt.Sprintf("%s • %s", custName, sub)
					}
					var t *time.Time
					if createdAt.Valid {
						t = &createdAt.Time
					}
					badgeType := "neutral"
					if stage == "WON" || stage == "QUOTED" {
						badgeType = "success"
					}
					items = append(items, SearchItem{
						ID:        fmt.Sprintf("rfq-%d", id),
						Category:  CategoryRFQ,
						Title:     num,
						Subtitle:  sub,
						Badge:     stage,
						BadgeType: badgeType,
						URL:       fmt.Sprintf("/dashboard/rfqs/%d", id),
						CreatedAt: t,
					})
				}
			}
			rows.Close()
		}
	}

	// 4. QUOTATIONS
	if filterCat == "" || filterCat == "ALL" || filterCat == "QUOTATION" || filterCat == "QUOTATIONS" {
		quoteQuery := `
			SELECT rq.id, rq.rfq_id, COALESCE(rq.quote_reference, '') AS ref,
			       COALESCE(rq.carrier_name, '') AS carrier, COALESCE(rq.status, 'DRAFT') AS status,
			       COALESCE(rq.sell_price, 0) AS sell_price, COALESCE(rq.currency, 'USD') AS curr,
			       rq.created_at
			FROM rfq_quotes rq
			JOIN rfqs r ON rq.rfq_id = r.id
			WHERE r.org_id = ? AND (
				rq.quote_reference LIKE ? OR
				rq.carrier_name LIKE ?
			)
			ORDER BY rq.id DESC LIMIT ?`
		rows, err := r.db.QueryContext(ctx, quoteQuery, orgID, pattern, pattern, perCategoryLimit)
		if err == nil {
			for rows.Next() {
				var id, rfqID int64
				var ref, carrier, status, curr string
				var sellPrice float64
				var createdAt sql.NullTime
				if err := rows.Scan(&id, &rfqID, &ref, &carrier, &status, &sellPrice, &curr, &createdAt); err == nil {
					title := ref
					if title == "" {
						title = fmt.Sprintf("Quote #%d", id)
					}
					sub := fmt.Sprintf("Carrier: %s • $%s %s", carrier, fmt.Sprintf("%.2f", sellPrice), curr)
					var t *time.Time
					if createdAt.Valid {
						t = &createdAt.Time
					}
					badgeType := "neutral"
					if status == "APPROVED" || status == "WON" {
						badgeType = "success"
					}
					items = append(items, SearchItem{
						ID:        fmt.Sprintf("quote-%d", id),
						Category:  CategoryQuotation,
						Title:     title,
						Subtitle:  sub,
						Badge:     status,
						BadgeType: badgeType,
						URL:       fmt.Sprintf("/dashboard/rfqs/%d", rfqID),
						CreatedAt: t,
					})
				}
			}
			rows.Close()
		}
	}

	// 5. CUSTOMERS & COMPANIES
	if filterCat == "" || filterCat == "ALL" || filterCat == "CUSTOMER" || filterCat == "CUSTOMERS" {
		custQuery := `
			SELECT cust.id, COALESCE(comp.name, '') AS name, COALESCE(comp.domain, '') AS domain,
			       COALESCE(comp.industry, '') AS industry, COALESCE(cust.status, 'ACTIVE') AS status,
			       cust.created_at
			FROM customers cust
			JOIN companies comp ON cust.company_id = comp.id
			WHERE cust.org_id = ? AND (
				comp.name LIKE ? OR
				comp.domain LIKE ? OR
				comp.industry LIKE ?
			)
			ORDER BY cust.id DESC LIMIT ?`
		rows, err := r.db.QueryContext(ctx, custQuery, orgID, pattern, pattern, pattern, perCategoryLimit)
		if err == nil {
			for rows.Next() {
				var id int64
				var name, domain, industry, status string
				var createdAt sql.NullTime
				if err := rows.Scan(&id, &name, &domain, &industry, &status, &createdAt); err == nil {
					sub := industry
					if domain != "" {
						if sub != "" {
							sub = fmt.Sprintf("%s • %s", domain, sub)
						} else {
							sub = domain
						}
					}
					if sub == "" {
						sub = "Customer Account"
					}
					var t *time.Time
					if createdAt.Valid {
						t = &createdAt.Time
					}
					items = append(items, SearchItem{
						ID:        fmt.Sprintf("customer-%d", id),
						Category:  CategoryCustomer,
						Title:     name,
						Subtitle:  sub,
						Badge:     status,
						BadgeType: "info",
						URL:       fmt.Sprintf("/dashboard/customers/%d", id),
						CreatedAt: t,
					})
				}
			}
			rows.Close()
		}
	}

	// 6. INVOICES
	if filterCat == "" || filterCat == "ALL" || filterCat == "INVOICE" || filterCat == "INVOICES" {
		invQuery := `
			SELECT id, COALESCE(invoice_number, '') AS num, COALESCE(vendor_name, '') AS vendor,
			       COALESCE(total_amount, 0) AS total, COALESCE(currency, 'USD') AS curr,
			       COALESCE(status, 'ISSUED') AS status, created_at
			FROM shipment_invoices
			WHERE org_id = ? AND (
				invoice_number LIKE ? OR
				vendor_name LIKE ? OR
				vendor_ref LIKE ?
			)
			ORDER BY id DESC LIMIT ?`
		rows, err := r.db.QueryContext(ctx, invQuery, orgID, pattern, pattern, pattern, perCategoryLimit)
		if err == nil {
			for rows.Next() {
				var id int64
				var num, vendor, curr, status string
				var total float64
				var createdAt sql.NullTime
				if err := rows.Scan(&id, &num, &vendor, &total, &curr, &status, &createdAt); err == nil {
					sub := fmt.Sprintf("%s • $%s %s", vendor, fmt.Sprintf("%.2f", total), curr)
					var t *time.Time
					if createdAt.Valid {
						t = &createdAt.Time
					}
					badgeType := "warning"
					if strings.ToUpper(status) == "PAID" {
						badgeType = "success"
					}
					items = append(items, SearchItem{
						ID:        fmt.Sprintf("invoice-%d", id),
						Category:  CategoryInvoice,
						Title:     num,
						Subtitle:  sub,
						Badge:     status,
						BadgeType: badgeType,
						URL:       "/dashboard/finance/invoices",
						CreatedAt: t,
					})
				}
			}
			rows.Close()
		}
	}

	// 7. LEADS
	if filterCat == "" || filterCat == "ALL" || filterCat == "LEAD" || filterCat == "LEADS" {
		leadQuery := `
			SELECT id, COALESCE(company_name, '') AS company, COALESCE(contact_name, '') AS contact,
			       COALESCE(email, '') AS email, COALESCE(status, 'NEW') AS status, created_at
			FROM leads
			WHERE org_id = ? AND (
				company_name LIKE ? OR
				contact_name LIKE ? OR
				email LIKE ? OR
				phone LIKE ?
			)
			ORDER BY id DESC LIMIT ?`
		rows, err := r.db.QueryContext(ctx, leadQuery, orgID, pattern, pattern, pattern, pattern, perCategoryLimit)
		if err == nil {
			for rows.Next() {
				var id int64
				var company, contact, email, status string
				var createdAt sql.NullTime
				if err := rows.Scan(&id, &company, &contact, &email, &status, &createdAt); err == nil {
					sub := contact
					if email != "" {
						if sub != "" {
							sub = fmt.Sprintf("%s (%s)", sub, email)
						} else {
							sub = email
						}
					}
					var t *time.Time
					if createdAt.Valid {
						t = &createdAt.Time
					}
					items = append(items, SearchItem{
						ID:        fmt.Sprintf("lead-%d", id),
						Category:  CategoryLead,
						Title:     company,
						Subtitle:  sub,
						Badge:     status,
						BadgeType: "neutral",
						URL:       "/dashboard/leads",
						CreatedAt: t,
					})
				}
			}
			rows.Close()
		}
	}

	// 8. CONTRACTS
	if filterCat == "" || filterCat == "ALL" || filterCat == "CONTRACT" || filterCat == "CONTRACTS" {
		contractQuery := `
			SELECT id, COALESCE(file_name, '') AS file, COALESCE(carrier_name, '') AS carrier,
			       COALESCE(carrier_scac, '') AS scac, COALESCE(status, 'CONFIRMED') AS status,
			       created_at
			FROM contract_documents
			WHERE org_id = ? AND (
				file_name LIKE ? OR
				carrier_name LIKE ? OR
				carrier_scac LIKE ?
			)
			ORDER BY id DESC LIMIT ?`
		rows, err := r.db.QueryContext(ctx, contractQuery, orgID, pattern, pattern, pattern, perCategoryLimit)
		if err == nil {
			for rows.Next() {
				var id, file, carrier, scac, status string
				var createdAt sql.NullTime
				if err := rows.Scan(&id, &file, &carrier, &scac, &status, &createdAt); err == nil {
					sub := carrier
					if scac != "" {
						sub = fmt.Sprintf("%s (%s)", sub, scac)
					}
					var t *time.Time
					if createdAt.Valid {
						t = &createdAt.Time
					}
					items = append(items, SearchItem{
						ID:        fmt.Sprintf("contract-%s", id),
						Category:  CategoryContract,
						Title:     file,
						Subtitle:  sub,
						Badge:     status,
						BadgeType: "info",
						URL:       "/dashboard/contracts",
						CreatedAt: t,
					})
				}
			}
			rows.Close()
		}
	}

	// 9. TRACKING & CONTAINER TELEMETRY
	if filterCat == "" || filterCat == "ALL" || filterCat == "TRACKING" || filterCat == "CONTAINERS" {
		trackQuery := `
			SELECT id, COALESCE(container_number, '') AS container, COALESCE(carrier_scac, '') AS scac,
			       COALESCE(booking_number, '') AS bnum, COALESCE(milestone_code, '') AS milestone,
			       COALESCE(location, '') AS loc, created_at
			FROM carrier_tracking_events
			WHERE org_id = ? AND (
				container_number LIKE ? OR
				booking_number LIKE ? OR
				mbl_number LIKE ? OR
				event_id LIKE ?
			)
			ORDER BY id DESC LIMIT ?`
		rows, err := r.db.QueryContext(ctx, trackQuery, orgID, pattern, pattern, pattern, pattern, perCategoryLimit)
		if err == nil {
			for rows.Next() {
				var id int64
				var container, scac, bnum, milestone, loc string
				var createdAt sql.NullTime
				if err := rows.Scan(&id, &container, &scac, &bnum, &milestone, &loc, &createdAt); err == nil {
					title := container
					if title == "" {
						title = fmt.Sprintf("Tracking Event #%d", id)
					}
					sub := loc
					if scac != "" {
						sub = fmt.Sprintf("%s • %s", scac, sub)
					}
					if bnum != "" {
						sub = fmt.Sprintf("%s (Booking: %s)", sub, bnum)
					}
					var t *time.Time
					if createdAt.Valid {
						t = &createdAt.Time
					}
					items = append(items, SearchItem{
						ID:        fmt.Sprintf("tracking-%d", id),
						Category:  CategoryTracking,
						Title:     title,
						Subtitle:  sub,
						Badge:     milestone,
						BadgeType: "info",
						URL:       "/dashboard/tracking",
						CreatedAt: t,
					})
				}
			}
			rows.Close()
		}
	}

	return items, nil
}
