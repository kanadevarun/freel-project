package search

import (
	"context"
	"strings"
)

type Service interface {
	Search(ctx context.Context, orgID int64, query string, entityType string, limit int) (*SearchResponse, error)
}

type searchService struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &searchService{repo: repo}
}

func (s *searchService) Search(ctx context.Context, orgID int64, query string, entityType string, limit int) (*SearchResponse, error) {
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return &SearchResponse{
			Query:        "",
			TotalMatches: 0,
			Groups:       []SearchResultGroup{},
		}, nil
	}

	items, err := s.repo.SearchAll(ctx, orgID, trimmed, entityType, limit)
	if err != nil {
		return nil, err
	}

	// Group results by category
	groupMap := make(map[Category][]SearchItem)
	categoryOrder := []Category{
		CategoryShipment,
		CategoryBooking,
		CategoryRFQ,
		CategoryQuotation,
		CategoryCustomer,
		CategoryInvoice,
		CategoryLead,
		CategoryContract,
		CategoryDocument,
		CategoryTracking,
	}

	for _, item := range items {
		groupMap[item.Category] = append(groupMap[item.Category], item)
	}

	categoryLabels := map[Category]string{
		CategoryShipment:  "Shipments",
		CategoryBooking:   "Bookings",
		CategoryRFQ:       "RFQs",
		CategoryQuotation: "Quotations",
		CategoryCustomer:  "Customers & Companies",
		CategoryInvoice:   "Invoices & Billing",
		CategoryLead:      "CRM Leads",
		CategoryContract:  "Rate Contracts",
		CategoryDocument:  "Transport Documents",
		CategoryTracking:  "Live Telemetry & Containers",
	}

	categoryIcons := map[Category]string{
		CategoryShipment:  "Ship",
		CategoryBooking:   "Layers",
		CategoryRFQ:       "FileText",
		CategoryQuotation: "BadgePercent",
		CategoryCustomer:  "Users",
		CategoryInvoice:   "CreditCard",
		CategoryLead:      "Sparkles",
		CategoryContract:  "FileSpreadsheet",
		CategoryDocument:  "FolderOpen",
		CategoryTracking:  "Radio",
	}

	groups := make([]SearchResultGroup, 0)
	total := 0

	for _, cat := range categoryOrder {
		if catItems, ok := groupMap[cat]; ok && len(catItems) > 0 {
			groups = append(groups, SearchResultGroup{
				Category:      cat,
				CategoryLabel: categoryLabels[cat],
				Icon:          categoryIcons[cat],
				Count:         len(catItems),
				Items:         catItems,
			})
			total += len(catItems)
		}
	}

	return &SearchResponse{
		Query:        trimmed,
		TotalMatches: total,
		Groups:       groups,
	}, nil
}
