package ai

import (
	"context"
	"strings"
	"time"
)

// mockProvider is a fake AI that returns hardcoded responses.
type mockProvider struct{}

func NewMockProvider() Provider {
	return &mockProvider{}
}

func (m *mockProvider) GenerateCompletion(ctx context.Context, prompt string) (string, error) {
	time.Sleep(100 * time.Millisecond)

	lower := strings.ToLower(prompt)

	if strings.Contains(lower, "15,000 kg") || strings.Contains(lower, "15000") || strings.Contains(lower, "actually, the weight is") {
		return `{
			"intent": "RFQ_REQUEST",
			"sentiment": "NEUTRAL",
			"summary": "Customer corrected the cargo weight to 15,000 kg.",
			"partial_rfq_context": {
				"cargo_weight": "15000"
			}
		}`, nil
	}

	if strings.Contains(lower, "weight is") || strings.Contains(lower, "volume is") || strings.Contains(lower, "12,000 kg") || strings.Contains(lower, "18 cbm") {
		return `{
			"intent": "RFQ_REQUEST",
			"sentiment": "NEUTRAL",
			"summary": "Customer provided missing weight (12,000 kg) and volume (18 CBM) for shipping request.",
			"partial_rfq_context": {
				"cargo_weight": "12000",
				"cargo_volume": "18"
			}
		}`, nil
	}

	if strings.Contains(lower, "machinery parts") || strings.Contains(lower, "don't have weight") {
		return `{
			"intent": "RFQ_REQUEST_INCOMPLETE",
			"sentiment": "NEUTRAL",
			"summary": "Incomplete Quote Request from NHAVA SHEVA (INNSA) to HAMBURG (DEHAM). Missing mandatory fields: Cargo Weight, Cargo Volume.",
			"drafted_reply": "Dear Varun,\n\nThank you for reaching out for a shipping quote from Mumbai to Hamburg. We would be happy to assist you with transporting your steel parts next month.\n\nTo provide you with an accurate rate, could you please provide missing details?\n\nBest regards,\nLogisticsHQ Sales Team",
			"partial_rfq_context": {
				"lead_name": "Varun Kanade",
				"cargo_description": "machinery parts",
				"origin_port": "NHAVA SHEVA (INNSA)",
				"destination_port": "HAMBURG (DEHAM)",
				"incoterms": "FOB",
				"target_date": "2026-11-20"
			}
		}`, nil
	}

	if strings.Contains(lower, "industrial valves") || strings.Contains(lower, "20 tons") {
		return `{
			"intent": "RFQ_REQUEST",
			"sentiment": "NEUTRAL",
			"summary": "The sender is requesting an ocean freight quote for shipping 20 tons of industrial valves from Nhava Sheva to Hamburg, with a cargo ready date of October 15, 2026.",
			"partial_rfq_context": {
				"lead_name": "Varun Kanade",
				"cargo_description": "industrial valves",
				"cargo_weight": "20000",
				"origin_port": "NHAVA SHEVA (INNSA)",
				"destination_port": "HAMBURG (DEHAM)",
				"incoterms": "FOB",
				"target_date": "2026-10-15"
			}
		}`, nil
	}

	if strings.Contains(lower, "brochure") || strings.Contains(lower, "catalog") || strings.Contains(lower, "general info") {
		return `{
			"intent": "QUESTION",
			"sentiment": "NEUTRAL",
			"summary": "A new potential client is requesting general company information, specifically a brochure and catalog.",
			"partial_rfq_context": {}
		}`, nil
	}

	fakeJSON := `{
		"score": 85,
		"research_report": "This company has high revenue and ships a significant volume of containers. They are a strong prospect for our logistics services."
	}`

	return fakeJSON, nil
}
