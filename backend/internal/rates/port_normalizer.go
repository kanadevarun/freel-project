package rates

import "strings"

// portAliasMap maps carrier-specific port names and common aliases
// to their standard UN/LOCODE. This is the authoritative normalization
// dictionary — add entries here when new carrier formats introduce new aliases.
//
// Format: "UPPERCASED ALIAS" → "LOCODE"
var portAliasMap = map[string]string{
	// ── India ────────────────────────────────────────────────────────────────
	"NHAVA SHEVA":       "INNSA",
	"NHAVASHEVA":        "INNSA",
	"JNPT":              "INNSA",
	"NAVI MUMBAI":       "INNSA",
	"JAWAHARLAL NEHRU":  "INNSA",
	"MUNDRA":            "INMUN",
	"MUNDRA PORT":       "INMUN",
	"ADANI MUNDRA":      "INMUN",
	"CHENNAI":           "INMAA",
	"MADRAS":            "INMAA",
	"KOLKATA":           "INCCU",
	"CALCUTTA":          "INCCU",
	"KOCHI":             "INCOK",
	"COCHIN":            "INCOK",
	"PIPAVAV":           "INVPZ",
	"KANDLA":            "INKLA",
	"HAZIRA":            "INHZA",
	"SURAT":             "INHZA",
	"VIZAG":             "INVTZ",
	"VISAKHAPATNAM":     "INVTZ",
	"TUTICORIN":         "INTUT",
	"THOOTHUKUDI":       "INTUT",

	// ── Germany ───────────────────────────────────────────────────────────────
	"HAMBURG":           "DEHAM",
	"HH":                "DEHAM",
	"BREMERHAVEN":       "DEBRV",
	"BREMEN":            "DEBRV",
	"DUISBURG":          "DEDUI",

	// ── Netherlands ───────────────────────────────────────────────────────────
	"ROTTERDAM":         "NLRTM",
	"RTM":               "NLRTM",

	// ── Belgium ───────────────────────────────────────────────────────────────
	"ANTWERP":           "BEANR",
	"ANR":               "BEANR",

	// ── United Kingdom ────────────────────────────────────────────────────────
	"FELIXSTOWE":        "GBFXT",
	"SOUTHAMPTON":       "GBSOU",
	"LONDON GATEWAY":    "GBLON",

	// ── Spain ─────────────────────────────────────────────────────────────────
	"BARCELONA":         "ESBCN",
	"VALENCIA":          "ESVLC",
	"ALGECIRAS":         "ESALG",

	// ── Italy ─────────────────────────────────────────────────────────────────
	"GENOA":             "ITGOA",
	"GENOVA":            "ITGOA",
	"LA SPEZIA":         "ITLSZ",

	// ── France ────────────────────────────────────────────────────────────────
	"LE HAVRE":          "FRLEH",
	"MARSEILLE":         "FRMRS",

	// ── Singapore ─────────────────────────────────────────────────────────────
	"SINGAPORE":         "SGSIN",
	"SGP":               "SGSIN",
	"PSA":               "SGSIN",

	// ── Malaysia ──────────────────────────────────────────────────────────────
	"PORT KLANG":        "MYPKG",
	"KLANG":             "MYPKG",
	"TANJUNG PELEPAS":   "MYMYY",
	"PTP":               "MYMYY",

	// ── China ─────────────────────────────────────────────────────────────────
	"SHANGHAI":          "CNSHA",
	"NINGBO":            "CNNGB",
	"NINGBO-ZHOUSHAN":   "CNNGB",
	"QINGDAO":           "CNTAO",
	"TIANJIN":           "CNTXG",
	"XINGANG":           "CNTXG",
	"GUANGZHOU":         "CNGZH",
	"NANSHA":            "CNNSZ",
	"YANTIAN":           "CNYTN",
	"SHEKOU":            "CNSKU",
	"SHENZHEN":          "CNSZN",
	"XIAMEN":            "CNXMN",
	"DALIAN":            "CNDLC",
	"BUSAN":             "KRPUS",
	"PUSAN":             "KRPUS",

	// ── UAE ───────────────────────────────────────────────────────────────────
	"JEBEL ALI":         "AEJEA",
	"DUBAI":             "AEDXB",
	"JEA":               "AEJEA",

	// ── USA ───────────────────────────────────────────────────────────────────
	"LOS ANGELES":       "USLAX",
	"LONG BEACH":        "USLGB",
	"NEW YORK":          "USNYC",
	"NEWARK":            "USNWK",
	"SAVANNAH":          "USSAV",
	"CHARLESTON":        "USCHS",
	"HOUSTON":           "USHOU",
	"SEATTLE":           "USSEA",
	"TACOMA":            "USTIW",

	// ── Australia ────────────────────────────────────────────────────────────
	"MELBOURNE":         "AUMEL",
	"SYDNEY":            "AUSYD",
	"BRISBANE":          "AUBNE",
	"FREMANTLE":         "AUFRE",
}

// NormalizePort converts a free-text port name into its UN/LOCODE.
//
// Resolution order:
//  1. If the input is already a valid 5-char UN/LOCODE, return as-is (uppercased).
//  2. Look up the uppercased, trimmed input in portAliasMap.
//  3. If not found, return the original input uppercased — the AI extraction
//     pipeline will flag it with review_flags=["PORT_UNKNOWN"].
//
// Example:
//
//	NormalizePort("Nhava Sheva") → "INNSA"
//	NormalizePort("INNSA")       → "INNSA"
//	NormalizePort("HAMBURG")     → "DEHAM"
func NormalizePort(input string) string {
	if input == "" {
		return ""
	}
	upper := strings.ToUpper(strings.TrimSpace(input))

	// Already a UN/LOCODE (country-code 2 chars + location-code 3 chars = 5 chars, all alpha)
	if isLOCODE(upper) {
		return upper
	}

	if locode, ok := portAliasMap[upper]; ok {
		return locode
	}

	// Unknown — return uppercased so at least casing is consistent.
	// The caller should add "PORT_UNKNOWN" to review_flags.
	return upper
}

// isLOCODE returns true if s looks like a standard UN/LOCODE:
// 5 uppercase ASCII letters (e.g., "DEHAM", "INNSA", "SGSIN").
func isLOCODE(s string) bool {
	if len(s) != 5 {
		return false
	}
	for _, c := range s {
		if c < 'A' || c > 'Z' {
			return false
		}
	}
	return true
}

// IsKnownPort returns true if the input maps to a known UN/LOCODE.
// Useful for the validator agent to decide whether to add PORT_UNKNOWN flag.
func IsKnownPort(input string) bool {
	upper := strings.ToUpper(strings.TrimSpace(input))
	if isLOCODE(upper) {
		return true
	}
	_, ok := portAliasMap[upper]
	return ok
}

// SearchPorts searches the portAliasMap for keys or values containing the query.
// It returns a map of all matching aliases and standard LOCODEs.
func SearchPorts(query string) map[string]string {
	query = strings.ToUpper(strings.TrimSpace(query))
	results := make(map[string]string)
	if query == "" {
		return results
	}
	for alias, locode := range portAliasMap {
		if strings.Contains(alias, query) || strings.Contains(locode, query) {
			results[alias] = locode
		}
	}
	return results
}

