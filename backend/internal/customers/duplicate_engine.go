package customers

import (
	"strings"
)

// EvaluateDuplicateScore calculates the confidence match score (0-100) between an incoming request and an existing customer record.
func EvaluateDuplicateScore(req CheckDuplicateReq, candidate Customer) (score int, reason string) {
	reasons := []string{}
	highestScore := 0

	// Rule 1: Exact Tax ID / GSTIN Match (100% confidence)
	if req.TaxID != nil && candidate.TaxID != nil && *req.TaxID != "" && *candidate.TaxID != "" {
		cleanReqTax := cleanString(*req.TaxID)
		cleanCandTax := cleanString(*candidate.TaxID)
		if cleanReqTax != "" && cleanReqTax == cleanCandTax {
			return 100, "Tax ID / GSTIN exact match"
		}
	}

	// Rule 2: Exact Domain Match (85% confidence)
	reqDomain := extractDomain(req.Domain, req.Email)
	candDomain := extractDomain(candidate.Domain, candidate.ContactEmail)
	domainMatched := false
	if reqDomain != "" && candDomain != "" && reqDomain == candDomain {
		domainMatched = true
		highestScore = maxInt(highestScore, 85)
		reasons = append(reasons, "Domain match ("+reqDomain+")")
	}

	// Rule 3: Cleaned Name Similarity (80% confidence for exact normalized match, lower for partial)
	reqNameClean := cleanString(req.Name)
	candNameClean := cleanString(candidate.Name)
	if reqNameClean != "" && candNameClean != "" {
		if reqNameClean == candNameClean {
			highestScore = maxInt(highestScore, 80)
			reasons = append(reasons, "Legal company name exact match")
		} else if strings.Contains(candNameClean, reqNameClean) || strings.Contains(reqNameClean, candNameClean) {
			if len(reqNameClean) > 4 && len(candNameClean) > 4 {
				highestScore = maxInt(highestScore, 65)
				reasons = append(reasons, "Company name partial match")
			}
		}
	}

	// Rule 4: Contact Email Match (75% confidence)
	if req.Email != nil && candidate.ContactEmail != nil && *req.Email != "" && *candidate.ContactEmail != "" {
		if strings.EqualFold(strings.TrimSpace(*req.Email), strings.TrimSpace(*candidate.ContactEmail)) {
			highestScore = maxInt(highestScore, 75)
			reasons = append(reasons, "Primary contact email match ("+*req.Email+")")
		}
	}

	// Rule 5: Contact Phone Match (70% confidence)
	if req.Phone != nil && candidate.ContactPhone != nil && *req.Phone != "" && *candidate.ContactPhone != "" {
		cleanReqPhone := cleanPhone(*req.Phone)
		cleanCandPhone := cleanPhone(*candidate.ContactPhone)
		if cleanReqPhone != "" && cleanReqPhone == cleanCandPhone {
			highestScore = maxInt(highestScore, 70)
			reasons = append(reasons, "Primary contact phone match")
		}
	}

	// If domain + name partial match both hit, boost score to 90
	if domainMatched && highestScore >= 65 {
		highestScore = maxInt(highestScore, 90)
	}

	if highestScore == 0 {
		return 0, ""
	}

	return highestScore, strings.Join(reasons, ", ")
}

func cleanString(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	// Strip common company suffixes for normalized matching
	suffixes := []string{"pvt", "ltd", "inc", "corp", "co", "llc", "gmbh", "private", "limited", "corporation", "exports", "imports", "logistics"}
	for _, suf := range suffixes {
		s = strings.ReplaceAll(s, suf, "")
	}
	var sb strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

func extractDomain(domainPtr, emailPtr *string) string {
	if domainPtr != nil && *domainPtr != "" {
		d := strings.ToLower(strings.TrimSpace(*domainPtr))
		d = strings.TrimPrefix(d, "http://")
		d = strings.TrimPrefix(d, "https://")
		d = strings.TrimPrefix(d, "www.")
		parts := strings.Split(d, "/")
		return parts[0]
	}
	if emailPtr != nil && *emailPtr != "" {
		parts := strings.Split(*emailPtr, "@")
		if len(parts) == 2 {
			d := strings.ToLower(strings.TrimSpace(parts[1]))
			// Ignore common public email providers for corporate domain matching
			publicProviders := map[string]bool{"gmail.com": true, "yahoo.com": true, "hotmail.com": true, "outlook.com": true, "icloud.com": true}
			if !publicProviders[d] {
				return d
			}
		}
	}
	return ""
}

func cleanPhone(phone string) string {
	var sb strings.Builder
	for _, r := range phone {
		if r >= '0' && r <= '9' {
			sb.WriteRune(r)
		}
	}
	res := sb.String()
	if len(res) > 10 {
		res = res[len(res)-10:] // compare last 10 digits
	}
	return res
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
