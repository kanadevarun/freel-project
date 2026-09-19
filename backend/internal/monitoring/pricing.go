package monitoring

import (
	"strings"
	"unicode/utf8"
)

// DefaultPricingCatalog provides built-in reference pricing per 1,000 tokens (USD).
var DefaultPricingCatalog = map[string]struct {
	InputCostPer1k  float64
	OutputCostPer1k float64
	Currency        string
}{
	// Gemini models
	"gemini:gemini-1.5-flash": {InputCostPer1k: 0.000075, OutputCostPer1k: 0.000300, Currency: "USD"},
	"gemini:gemini-1.5-pro":   {InputCostPer1k: 0.001250, OutputCostPer1k: 0.005000, Currency: "USD"},
	"gemini:default":          {InputCostPer1k: 0.000075, OutputCostPer1k: 0.000300, Currency: "USD"},

	// OpenAI models
	"openai:gpt-4o-mini": {InputCostPer1k: 0.000150, OutputCostPer1k: 0.000600, Currency: "USD"},
	"openai:gpt-4o":      {InputCostPer1k: 0.002500, OutputCostPer1k: 0.010000, Currency: "USD"},
	"openai:default":     {InputCostPer1k: 0.000150, OutputCostPer1k: 0.000600, Currency: "USD"},

	// Anthropic models
	"anthropic:claude-3-5-sonnet": {InputCostPer1k: 0.003000, OutputCostPer1k: 0.015000, Currency: "USD"},
	"anthropic:default":           {InputCostPer1k: 0.003000, OutputCostPer1k: 0.015000, Currency: "USD"},

	// Local Mock & Fallback Provider (Free)
	"mock:mock-evaluator": {InputCostPer1k: 0.000000, OutputCostPer1k: 0.000000, Currency: "USD"},
	"mock:default":        {InputCostPer1k: 0.000000, OutputCostPer1k: 0.000000, Currency: "USD"},
}

// EstimateTokens calculates an approximate token count based on standard ~4 characters per token heuristic.
func EstimateTokens(text string) int {
	if text == "" {
		return 0
	}
	charCount := utf8.RuneCountInString(text)
	tokens := charCount / 4
	if tokens == 0 && charCount > 0 {
		return 1
	}
	return tokens
}

// CalculateCost computes estimated dollar cost for a given provider, model, and token counts.
// If customRates map is supplied (from DB), it checks customRates first, then DefaultPricingCatalog.
func CalculateCost(provider, model string, inputTokens, outputTokens int, customRates map[string]ModelPricingRecord) (float64, string) {
	providerKey := strings.ToLower(strings.TrimSpace(provider))
	modelKey := strings.ToLower(strings.TrimSpace(model))
	lookupKey := providerKey + ":" + modelKey
	defaultKey := providerKey + ":default"

	// 1. Check custom rates from DB if available
	if customRates != nil {
		if rate, exists := customRates[lookupKey]; exists && rate.IsActive {
			cost := (float64(inputTokens)/1000.0)*rate.InputCostPer1kTokens +
				(float64(outputTokens)/1000.0)*rate.OutputCostPer1kTokens
			return cost, rate.Currency
		}
		if rate, exists := customRates[defaultKey]; exists && rate.IsActive {
			cost := (float64(inputTokens)/1000.0)*rate.InputCostPer1kTokens +
				(float64(outputTokens)/1000.0)*rate.OutputCostPer1kTokens
			return cost, rate.Currency
		}
	}

	// 2. Check built-in reference catalog
	if rate, exists := DefaultPricingCatalog[lookupKey]; exists {
		cost := (float64(inputTokens)/1000.0)*rate.InputCostPer1k +
			(float64(outputTokens)/1000.0)*rate.OutputCostPer1k
		return cost, rate.Currency
	}

	if rate, exists := DefaultPricingCatalog[defaultKey]; exists {
		cost := (float64(inputTokens)/1000.0)*rate.InputCostPer1k +
			(float64(outputTokens)/1000.0)*rate.OutputCostPer1k
		return cost, rate.Currency
	}

	// 3. Fallback generic rate ($0.00015 / 1k in, $0.0006 / 1k out)
	cost := (float64(inputTokens)/1000.0)*0.000150 + (float64(outputTokens)/1000.0)*0.000600
	return cost, "USD"
}
