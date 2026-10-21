package quotations

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DocumentGenerator defines the interface for generating quotation documents.
type DocumentGenerator interface {
	GenerateQuotationPDF(ctx context.Context, quote *CustomerQuotationPreview, version int) (filePath string, fileName string, pdfBytes []byte, err error)
}

type documentGenerator struct {
	storageDir string
}

// NewDocumentGenerator creates a new instance of DocumentGenerator.
func NewDocumentGenerator(storageDir string) DocumentGenerator {
	if storageDir == "" {
		storageDir = "./storage/quotations"
	}
	_ = os.MkdirAll(storageDir, 0755)
	return &documentGenerator{
		storageDir: storageDir,
	}
}

// GenerateQuotationPDF renders a customer-safe quotation preview into a professional PDF document.
func (g *documentGenerator) GenerateQuotationPDF(ctx context.Context, quote *CustomerQuotationPreview, version int) (string, string, []byte, error) {
	if quote == nil {
		return "", "", nil, fmt.Errorf("quote cannot be nil")
	}

	fileName := fmt.Sprintf("Quote_%s_v%d.pdf", sanitizeFileName(quote.QuotationNumber), version)
	filePath := filepath.Join(g.storageDir, fileName)

	pdfBytes := buildQuotationPDF(quote, version)

	// Persist to local storage directory
	if err := os.WriteFile(filePath, pdfBytes, 0644); err != nil {
		// Log warning but return pdfBytes so caller still has content
		fmt.Printf("[DocumentGenerator] Warning: failed to write PDF to %s: %v\n", filePath, err)
	}

	return filePath, fileName, pdfBytes, nil
}

func sanitizeFileName(name string) string {
	r := strings.NewReplacer("/", "_", "\\", "_", " ", "_", ":", "_")
	return r.Replace(name)
}

// ─────────────────────────────────────────────────────────────────────────────
// Pure Go PDF 1.4 Generation Engine
// ─────────────────────────────────────────────────────────────────────────────

type pdfCanvas struct {
	buf bytes.Buffer
}

func (c *pdfCanvas) write(s string) {
	c.buf.WriteString(s)
}

func (c *pdfCanvas) rect(x, y, w, h float64, fillR, fillG, fillB float64, strokeR, strokeG, strokeB float64, lineWidth float64) {
	c.buf.WriteString(fmt.Sprintf("%.3f %.3f %.3f rg\n", fillR, fillG, fillB))
	if lineWidth > 0 {
		c.buf.WriteString(fmt.Sprintf("%.3f %.3f %.3f RG\n", strokeR, strokeG, strokeB))
		c.buf.WriteString(fmt.Sprintf("%.2f w\n", lineWidth))
		c.buf.WriteString(fmt.Sprintf("%.2f %.2f %.2f %.2f re B\n", x, y, w, h))
	} else {
		c.buf.WriteString(fmt.Sprintf("%.2f %.2f %.2f %.2f re f\n", x, y, w, h))
	}
}

func (c *pdfCanvas) line(x1, y1, x2, y2 float64, r, g, b float64, lineWidth float64) {
	c.buf.WriteString(fmt.Sprintf("%.3f %.3f %.3f RG\n", r, g, b))
	c.buf.WriteString(fmt.Sprintf("%.2f w\n", lineWidth))
	c.buf.WriteString(fmt.Sprintf("%.2f %.2f m %.2f %.2f l S\n", x1, y1, x2, y2))
}

func (c *pdfCanvas) text(font string, size float64, x, y float64, text string, r, g, b float64) {
	escaped := escapePDFText(text)
	c.buf.WriteString(fmt.Sprintf("BT /%s %.2f Tf %.3f %.3f %.3f rg 1 0 0 1 %.2f %.2f Tm (%s) Tj ET\n", font, size, r, g, b, x, y, escaped))
}

func escapePDFText(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "(", "\\(")
	s = strings.ReplaceAll(s, ")", "\\)")
	return s
}

func (c *pdfCanvas) bullet(font string, size float64, x, y float64, text string, r, g, b float64) {
	escaped := escapePDFText(text)
	c.buf.WriteString(fmt.Sprintf("BT /%s %.2f Tf %.3f %.3f %.3f rg 1 0 0 1 %.2f %.2f Tm (\\225  %s) Tj ET\n", font, size, r, g, b, x, y, escaped))
}

func formatMoney(amount float64) string {
	raw := fmt.Sprintf("%.2f", amount)
	parts := strings.Split(raw, ".")
	intPart := parts[0]
	decPart := parts[1]

	var res []string
	for len(intPart) > 3 {
		res = append([]string{intPart[len(intPart)-3:]}, res...)
		intPart = intPart[:len(intPart)-3]
	}
	if len(intPart) > 0 {
		res = append([]string{intPart}, res...)
	}
	return strings.Join(res, ",") + "." + decPart
}

func buildQuotationPDF(quote *CustomerQuotationPreview, version int) []byte {
	c := &pdfCanvas{}

	// Page dimensions: Standard A4 (595.28 x 841.89 points)
	pageWidth := 595.28
	pageHeight := 841.89
	margin := 45.0
	contentWidth := pageWidth - 2.0*margin // 505.28 pt

	// Fallback values for dynamic fields if empty
	companyName := quote.CompanyName
	if strings.TrimSpace(companyName) == "" {
		companyName = "Freel Global Logistics Pvt Ltd"
	}
	companyAddress := quote.CompanyAddress
	if strings.TrimSpace(companyAddress) == "" {
		companyAddress = "Times Square Tower, Andheri-Kurla Road, Mumbai, India, 400059"
	}
	companyContact := quote.CompanyContact
	if strings.TrimSpace(companyContact) == "" {
		companyContact = "operations@freel.io | +91 (022) 6789-5500"
	}
	// Clean any non-WinAnsi bullet characters from header contacts
	companyContact = strings.ReplaceAll(companyContact, "•", "|")
	companyContact = strings.ReplaceAll(companyContact, "\u2022", "|")

	customerName := quote.CustomerName
	if strings.TrimSpace(customerName) == "" {
		customerName = "ABC Imports Pvt. Ltd."
	}
	attention := quote.CustomerAttention
	if strings.TrimSpace(attention) == "" {
		attention = "Logistics / Supply Chain Team"
	}

	transportMode := strings.ToUpper(strings.TrimSpace(quote.TransportMode))
	if transportMode == "" {
		transportMode = "SEA FREIGHT"
	}
	serviceType := quote.ServiceType
	if strings.TrimSpace(serviceType) == "" {
		serviceType = "FCL | Port-to-Port"
	}

	origin := quote.Origin
	if strings.TrimSpace(origin) == "" {
		origin = "Nhava Sheva, India"
	}
	destination := quote.Destination
	if strings.TrimSpace(destination) == "" {
		destination = "Jebel Ali, UAE"
	}
	equipment := quote.Equipment
	if strings.TrimSpace(equipment) == "" {
		equipment = "1 x 40' HC"
	}
	transitTime := quote.TransitTime
	if strings.TrimSpace(transitTime) == "" {
		transitTime = "Approx. 12-16 days"
	}
	commodity := quote.Commodity
	if strings.TrimSpace(commodity) == "" {
		commodity = "General cargo (sample)"
	}
	cargoWeight := quote.CargoWeight
	if strings.TrimSpace(cargoWeight) == "" {
		cargoWeight = "18,000 kg (estimated)"
	}
	incoterm := quote.Incoterm
	if strings.TrimSpace(incoterm) == "" {
		incoterm = "FOB (subject to confirmation)"
	}

	preparedBy := quote.PreparedByName
	if strings.TrimSpace(preparedBy) == "" {
		preparedBy = "Varun Kanade"
	}
	preparedEmail := quote.PreparedByEmail
	if strings.TrimSpace(preparedEmail) == "" {
		preparedEmail = "varunkanade3456@gmail.com"
	}

	currency := strings.ToUpper(strings.TrimSpace(quote.Currency))
	if currency == "" {
		currency = "USD"
	}

	quoteNo := quote.QuotationNumber
	if strings.TrimSpace(quoteNo) == "" {
		quoteNo = fmt.Sprintf("QT-2026-%04d", quote.QuotationID)
	}

	issuedDate := time.Now()
	if quote.SentAt != nil {
		issuedDate = *quote.SentAt
	} else if quote.ValidFrom != nil {
		issuedDate = *quote.ValidFrom
	}
	validUntilDate := issuedDate.Add(7 * 24 * time.Hour)
	if quote.ValidUntil != nil {
		validUntilDate = *quote.ValidUntil
	}
	dateStr := fmt.Sprintf("ISSUED %s | VALID UNTIL %s", issuedDate.Format("02 Jan 2006"), validUntilDate.Format("02 Jan 2006"))

	// ── 1. TOP HEADER (White background, Brand on Left, Freight Quotation on Right) ──
	logoY := pageHeight - 36.0 - 58.0
	logoW := 72.0
	logoH := 58.0

	// Logo container box
	c.rect(margin, logoY, logoW, logoH, 0.933, 0.961, 0.988, 0.098, 0.463, 0.824, 0.8)
	c.text("F2", 9.0, margin+20.0, logoY+33.0, "YOUR", 0.114, 0.165, 0.227)
	c.text("F2", 9.0, margin+20.0, logoY+20.0, "LOGO", 0.114, 0.165, 0.227)
	if quote.CompanyLogoURL != "" {
		c.text("F2", 6.0, margin+11.0, logoY+9.0, "BRAND LOGO", 0.15, 0.46, 0.82)
	}

	// Company Name & Info
	compX := margin + 84.0
	c.text("F2", 15.0, compX, logoY+40.0, truncateStr(companyName, 36), 0.078, 0.204, 0.357)
	c.text("F1", 8.2, compX, logoY+23.0, truncateStr(companyAddress, 54), 0.388, 0.459, 0.545)
	c.text("F1", 8.2, compX, logoY+9.0, truncateStr(companyContact, 56), 0.388, 0.459, 0.545)

	// Right Header: FREIGHT QUOTATION
	rightX := margin + 305.0
	c.text("F2", 15.0, rightX, logoY+40.0, "FREIGHT QUOTATION", 0.078, 0.204, 0.357)
	c.text("F2", 8.2, rightX, logoY+23.0, "QUOTE NO. ", 0.149, 0.204, 0.271)
	c.text("F2", 8.2, rightX+55.0, logoY+23.0, quoteNo, 0.098, 0.380, 0.675)
	c.text("F1", 7.8, rightX, logoY+9.0, dateStr, 0.388, 0.459, 0.545)

	// ── 2. SUB-BAR (Prepared for & Transport / Service Banner) ──────────────────
	subBarY := logoY - 14.0 - 36.0
	subBarH := 36.0
	c.rect(margin, subBarY, contentWidth, subBarH, 0.961, 0.973, 0.984, 0.835, 0.878, 0.918, 0.6)
	c.line(margin+305.0, subBarY, margin+305.0, subBarY+subBarH, 0.835, 0.878, 0.918, 0.6)

	c.text("F2", 8.8, margin+12.0, subBarY+20.0, "Prepared for "+truncateStr(customerName, 42), 0.078, 0.204, 0.357)
	c.text("F1", 7.8, margin+12.0, subBarY+8.0, "Attention: "+truncateStr(attention, 48), 0.388, 0.459, 0.545)

	c.text("F2", 8.8, margin+318.0, subBarY+20.0, transportMode, 0.078, 0.204, 0.357)
	c.text("F1", 7.8, margin+318.0, subBarY+8.0, serviceType, 0.388, 0.459, 0.545)

	// ── 3. SHIPMENT OVERVIEW (2 rows x 4 columns grid) ──────────────────────────
	curY := subBarY - 16.0
	c.text("F2", 10.0, margin, curY, "SHIPMENT OVERVIEW", 0.078, 0.176, 0.306)

	gridY := curY - 6.0 - 66.0
	gridH := 66.0
	colW := contentWidth / 4.0 // 126.32 pt per column

	c.rect(margin, gridY, contentWidth, gridH, 0.976, 0.984, 0.992, 0.835, 0.878, 0.918, 0.6)
	// Horizontal divider
	c.line(margin, gridY+33.0, margin+contentWidth, gridY+33.0, 0.835, 0.878, 0.918, 0.5)
	// Vertical dividers
	c.line(margin+colW, gridY, margin+colW, gridY+gridH, 0.835, 0.878, 0.918, 0.5)
	c.line(margin+2*colW, gridY, margin+2*colW, gridY+gridH, 0.835, 0.878, 0.918, 0.5)
	c.line(margin+3*colW, gridY, margin+3*colW, gridY+gridH, 0.835, 0.878, 0.918, 0.5)

	// Row 1 (Top)
	r1YLabel := gridY + 48.0
	r1YVal := gridY + 36.0
	c.text("F2", 7.0, margin+8.0, r1YLabel, "ORIGIN", 0.388, 0.459, 0.545)
	c.text("F2", 8.2, margin+8.0, r1YVal, truncateStr(origin, 26), 0.078, 0.204, 0.357)

	c.text("F2", 7.0, margin+colW+8.0, r1YLabel, "DESTINATION", 0.388, 0.459, 0.545)
	c.text("F2", 8.2, margin+colW+8.0, r1YVal, truncateStr(destination, 26), 0.078, 0.204, 0.357)

	c.text("F2", 7.0, margin+2*colW+8.0, r1YLabel, "EQUIPMENT", 0.388, 0.459, 0.545)
	c.text("F2", 8.2, margin+2*colW+8.0, r1YVal, truncateStr(equipment, 26), 0.078, 0.204, 0.357)

	c.text("F2", 7.0, margin+3*colW+8.0, r1YLabel, "TRANSIT TIME", 0.388, 0.459, 0.545)
	c.text("F2", 8.2, margin+3*colW+8.0, r1YVal, truncateStr(transitTime, 26), 0.078, 0.204, 0.357)

	// Row 2 (Bottom)
	r2YLabel := gridY + 20.0
	r2YVal := gridY + 8.0
	c.text("F2", 7.0, margin+8.0, r2YLabel, "COMMODITY", 0.388, 0.459, 0.545)
	c.text("F1", 7.8, margin+8.0, r2YVal, truncateStr(commodity, 28), 0.078, 0.204, 0.357)

	c.text("F2", 7.0, margin+colW+8.0, r2YLabel, "CARGO WEIGHT", 0.388, 0.459, 0.545)
	c.text("F1", 7.8, margin+colW+8.0, r2YVal, truncateStr(cargoWeight, 28), 0.078, 0.204, 0.357)

	c.text("F2", 7.0, margin+2*colW+8.0, r2YLabel, "SERVICE", 0.388, 0.459, 0.545)
	c.text("F1", 7.8, margin+2*colW+8.0, r2YVal, truncateStr(serviceType, 28), 0.078, 0.204, 0.357)

	c.text("F2", 7.0, margin+3*colW+8.0, r2YLabel, "INCOTERM", 0.388, 0.459, 0.545)
	c.text("F1", 7.5, margin+3*colW+8.0, r2YVal, truncateStr(incoterm, 32), 0.078, 0.204, 0.357)

	// ── 4. QUOTED CHARGES TABLE ─────────────────────────────────────────────────
	curY = gridY - 18.0
	c.text("F2", 10.0, margin, curY, "QUOTED CHARGES", 0.078, 0.176, 0.306)

	thY := curY - 6.0 - 24.0
	thH := 24.0
	// Navy Header banner
	c.rect(margin, thY, contentWidth, thH, 0.078, 0.204, 0.357, 0, 0, 0, 0)

	col1X := margin + 10.0  // CHARGE DESCRIPTION
	col2X := margin + 195.0 // BASIS
	col3X := margin + 300.0 // QTY
	col4X := margin + 355.0 // RATE
	col5X := margin + 430.0 // AMOUNT

	c.text("F2", 8.0, col1X, thY+8.0, "CHARGE DESCRIPTION", 1, 1, 1)
	c.text("F2", 8.0, col2X, thY+8.0, "BASIS", 1, 1, 1)
	c.text("F2", 8.0, col3X, thY+8.0, "QTY", 1, 1, 1)
	c.text("F2", 8.0, col4X, thY+8.0, fmt.Sprintf("RATE (%s)", currency), 1, 1, 1)
	c.text("F2", 8.0, col5X, thY+8.0, fmt.Sprintf("AMOUNT (%s)", currency), 1, 1, 1)

	// Items to render
	type chargeRow struct {
		Name   string
		Basis  string
		Qty    string
		Rate   string
		Amount string
	}
	var rows []chargeRow

	if len(quote.Charges) > 0 {
		for _, ch := range quote.Charges {
			basis := ch.CalculationBasis
			if strings.TrimSpace(basis) == "" {
				basis = "Per container"
			}
			qtyStr := fmt.Sprintf("%.0f", ch.Quantity)
			if ch.Quantity != float64(int(ch.Quantity)) {
				qtyStr = fmt.Sprintf("%.2f", ch.Quantity)
			}
			rows = append(rows, chargeRow{
				Name:   ch.ChargeName,
				Basis:  basis,
				Qty:    qtyStr,
				Rate:   formatMoney(ch.UnitPrice),
				Amount: formatMoney(ch.FinalAmount),
			})
		}
	} else if quote.TotalAmount > 0 {
		freightRate := quote.Subtotal
		if freightRate <= 0 {
			freightRate = quote.TotalAmount * 0.85
		}
		modeLabel := quote.TransportMode
		if strings.TrimSpace(modeLabel) == "" {
			modeLabel = "Ocean freight"
		}
		rows = append(rows, chargeRow{
			Name:   fmt.Sprintf("%s (%s to %s)", modeLabel, quote.OriginCode, quote.DestinationCode),
			Basis:  "Per container",
			Qty:    "1",
			Rate:   formatMoney(freightRate),
			Amount: formatMoney(freightRate),
		})
		if quote.TotalAmount > freightRate {
			surcharges := quote.TotalAmount - freightRate
			rows = append(rows, chargeRow{
				Name:   "Origin & Destination Port Surcharges",
				Basis:  "Per shipment",
				Qty:    "1",
				Rate:   formatMoney(surcharges),
				Amount: formatMoney(surcharges),
			})
		}
	} else {
		// Default fallback for brand-new empty quotes
		rows = []chargeRow{
			{Name: "Ocean freight", Basis: "Per container", Qty: "1", Rate: "1,250.00", Amount: "1,250.00"},
			{Name: "Origin terminal handling", Basis: "Per container", Qty: "1", Rate: "180.00", Amount: "180.00"},
			{Name: "Documentation fee", Basis: "Per shipment", Qty: "1", Rate: "45.00", Amount: "45.00"},
			{Name: "Export documentation / filing", Basis: "Per shipment", Qty: "1", Rate: "35.00", Amount: "35.00"},
		}
	}

	rowY := thY
	rowH := 24.0
	for i, r := range rows {
		if i >= 6 {
			// Limit to avoid overflow
			break
		}
		rowY -= rowH
		bgR, bgG, bgB := 1.0, 1.0, 1.0
		if i%2 == 1 {
			bgR, bgG, bgB = 0.980, 0.984, 0.990
		}
		c.rect(margin, rowY, contentWidth, rowH, bgR, bgG, bgB, 0.886, 0.910, 0.941, 0.4)

		c.text("F2", 8.2, col1X, rowY+7.5, truncateStr(r.Name, 36), 0.078, 0.204, 0.357)
		c.text("F1", 8.0, col2X, rowY+7.5, truncateStr(r.Basis, 22), 0.388, 0.459, 0.545)
		c.text("F1", 8.2, col3X, rowY+7.5, r.Qty, 0.078, 0.204, 0.357)
		c.text("F1", 8.2, col4X, rowY+7.5, r.Rate, 0.078, 0.204, 0.357)
		c.text("F2", 8.5, col5X, rowY+7.5, r.Amount, 0.078, 0.204, 0.357)
	}

	// Subtotal & Taxes lines
	subtotalVal := quote.Subtotal
	totalVal := quote.TotalAmount
	if totalVal <= 0 && len(rows) > 0 {
		subtotalVal = 1510.00
		totalVal = 1510.00
	} else if subtotalVal <= 0 && totalVal > 0 {
		subtotalVal = totalVal
	}

	curY = rowY - 14.0
	c.text("F1", 8.2, margin+285.0, curY, "Subtotal", 0.388, 0.459, 0.545)
	c.text("F2", 8.8, col5X, curY, fmt.Sprintf("%s %s", currency, formatMoney(subtotalVal)), 0.078, 0.204, 0.357)

	curY -= 14.0
	c.text("F1", 8.2, margin+285.0, curY, "Taxes / other charges", 0.388, 0.459, 0.545)
	taxesStr := "Not included unless stated"
	if quote.TaxTotal > 0 {
		taxesStr = fmt.Sprintf("%s %s", currency, formatMoney(quote.TaxTotal))
	}
	c.text("F1", 8.2, col5X, curY, taxesStr, 0.388, 0.459, 0.545)

	// TOTAL QUOTED AMOUNT Blue Bar
	totBarY := curY - 8.0 - 28.0
	totBarH := 28.0
	c.rect(margin, totBarY, contentWidth, totBarH, 0.098, 0.380, 0.675, 0, 0, 0, 0) // Vibrant Blue #1961AC
	c.text("F2", 11.0, margin+14.0, totBarY+9.0, "TOTAL QUOTED AMOUNT", 1, 1, 1)
	c.text("F2", 13.5, margin+contentWidth-150.0, totBarY+8.0, fmt.Sprintf("%s %s", currency, formatMoney(totalVal)), 1, 1, 1)

	// ── 5. INCLUSIONS, EXCLUSIONS & COMMERCIAL NOTES (from Option B / 4th diagram) ──
	curY = totBarY - 18.0
	c.text("F2", 10.0, margin, curY, "INCLUSIONS, EXCLUSIONS & COMMERCIAL NOTES", 0.078, 0.176, 0.306)

	// Anchored budget: Validity sits at 134-188, Prepared By sits at 66-122
	availBottom := 195.0
	boxesTop := curY - 6.0
	boxesH := boxesTop - availBottom
	if boxesH > 135.0 {
		boxesH = 135.0
	} else if boxesH < 90.0 {
		boxesH = 90.0
	}
	boxesY := boxesTop - boxesH
	halfW := (contentWidth - 10.0) / 2.0 // ~247.6 pt

	// Left Box: INCLUDED IN THIS OFFER
	incBoxX := margin
	// Header banner (Forest Green #217A55)
	c.rect(incBoxX, boxesY+boxesH-22.0, halfW, 22.0, 0.129, 0.478, 0.333, 0, 0, 0, 0)
	c.text("F2", 8.2, incBoxX+10.0, boxesY+boxesH-15.0, "INCLUDED IN THIS OFFER", 1, 1, 1)
	// Body
	c.rect(incBoxX, boxesY, halfW, boxesH-22.0, 0.960, 0.973, 0.984, 0.835, 0.878, 0.918, 0.5)

	inclusions := quote.Inclusions
	if len(inclusions) == 0 {
		inclusions = []string{
			"Ocean freight as listed",
			"Origin terminal handling as listed",
			"Export documentation as listed",
		}
	}
	incItemY := boxesY + boxesH - 38.0
	for _, inc := range inclusions {
		c.bullet("F1", 8.2, incBoxX+12.0, incItemY, inc, 0.149, 0.204, 0.271)
		incItemY -= 17.0
	}

	// Right Box: EXCLUDED / SUBJECT TO ADDITIONAL COST
	excBoxX := margin + halfW + 10.0
	// Header banner (Warm Amber/Brown #B66A15)
	c.rect(excBoxX, boxesY+boxesH-22.0, halfW, 22.0, 0.714, 0.416, 0.082, 0, 0, 0, 0)
	c.text("F2", 8.2, excBoxX+10.0, boxesY+boxesH-15.0, "EXCLUDED / SUBJECT TO ADDITIONAL COST", 1, 1, 1)
	// Body
	c.rect(excBoxX, boxesY, halfW, boxesH-22.0, 0.996, 0.984, 0.949, 0.980, 0.886, 0.749, 0.5)

	exclusions := quote.Exclusions
	if len(exclusions) == 0 {
		exclusions = []string{
			"Destination terminal/port charges",
			"Customs clearance, duties and taxes",
			"Delivery / inland haulage unless listed",
			"Insurance, inspections, demurrage and detention",
			"Storage, exams, penalties and exceptional charges",
		}
	}
	excItemY := boxesY + boxesH - 36.0
	for _, exc := range exclusions {
		c.bullet("F1", 7.8, excBoxX+12.0, excItemY, exc, 0.149, 0.204, 0.271)
		excItemY -= 14.5
	}

	// ── 6. VALIDITY & CONDITIONS PARAGRAPH (Balanced lower section) ─────────────
	curY = boxesY - 12.0
	c.text("F2", 7.5, margin, curY, "Validity and conditions:", 0.400, 0.467, 0.537)
	c.text("F1", 7.2, margin+94.0, curY, "Rates are illustrative and subject to space/equipment availability, final cargo details, carrier confirmation, surcharges and", 0.400, 0.467, 0.537)
	curY -= 10.5
	c.text("F1", 7.2, margin, curY, "changes imposed by carriers or authorities. Transit times are estimates, not guarantees. Any service or cost not expressly shown as included above", 0.400, 0.467, 0.537)
	curY -= 10.5
	c.text("F1", 7.2, margin, curY, "is excluded. Taxes, duties and destination charges are payable separately where applicable. Final booking is subject to written acceptance and", 0.400, 0.467, 0.537)
	curY -= 10.5
	c.text("F1", 7.2, margin, curY, "applicable commercial service terms.", 0.400, 0.467, 0.537)

	// ── 7. PREPARED BY & CONTACT CARD (Executive Layout, Bottom Anchored) ───────
	prepBoxH := 54.0
	prepBoxY := 66.0 // Anchored elegantly above the footer line
	c.rect(margin, prepBoxY, contentWidth, prepBoxH, 0.976, 0.984, 0.992, 0.835, 0.878, 0.918, 0.6)
	// Vertical divider
	c.line(margin+290.0, prepBoxY, margin+290.0, prepBoxY+prepBoxH, 0.835, 0.878, 0.918, 0.5)

	// Left: Prepared By info
	c.text("F2", 7.0, margin+12.0, prepBoxY+38.0, "AUTHORIZED COMMERCIAL REPRESENTATIVE", 0.118, 0.380, 0.675)
	c.text("F2", 8.5, margin+12.0, prepBoxY+23.0, "Prepared by:", 0.149, 0.204, 0.271)
	c.text("F2", 9.0, margin+78.0, prepBoxY+23.0, truncateStr(preparedBy, 34), 0.078, 0.204, 0.357)

	c.text("F2", 8.5, margin+12.0, prepBoxY+9.0, "Email:", 0.149, 0.204, 0.271)
	c.text("F1", 9.0, margin+78.0, prepBoxY+9.0, truncateStr(preparedEmail, 36), 0.098, 0.380, 0.675)

	// Right: Office & Contact info
	c.text("F2", 7.0, margin+302.0, prepBoxY+38.0, "COMMERCIAL DESK & SUPPORT", 0.118, 0.380, 0.675)
	c.text("F2", 8.2, margin+302.0, prepBoxY+23.0, "Headquarters:", 0.388, 0.459, 0.545)
	c.text("F1", 8.2, margin+372.0, prepBoxY+23.0, "Mumbai, India", 0.078, 0.204, 0.357)

	c.text("F2", 8.2, margin+302.0, prepBoxY+9.0, "Direct Phone:", 0.388, 0.459, 0.545)
	c.text("F1", 8.2, margin+372.0, prepBoxY+9.0, "+91 (022) 6789-5500", 0.078, 0.204, 0.357)

	// ── 8. PAGE FOOTER (Thin line + Legal Notice + Page 1) ──────────────────────
	c.line(margin, 46.0, margin+contentWidth, 46.0, 0.863, 0.898, 0.937, 0.6)
	c.text("F1", 7.2, margin, 32.0, "FREEL GLOBAL LOGISTICS | Official Commercial Quotation Confirmation | Confidential", 0.388, 0.459, 0.545)
	c.text("F2", 7.2, margin+contentWidth-36.0, 32.0, "Page 1 of 1", 0.388, 0.459, 0.545)

	return compilePDF(c.buf.Bytes(), pageWidth, pageHeight)
}

func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// compilePDF assembles standard PDF 1.4 objects into a complete PDF binary.
func compilePDF(streamBytes []byte, width, height float64) []byte {
	var out bytes.Buffer
	var offsets []int

	writeObj := func(data string) {
		offsets = append(offsets, out.Len())
		out.WriteString(data)
	}

	// 1. Header
	out.WriteString("%PDF-1.4\n%\xe2\xe3\xcf\xd3\n")

	// 2. Object 1: Catalog
	writeObj("1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n")

	// 3. Object 2: Pages
	writeObj(fmt.Sprintf("2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 /MediaBox [0 0 %.2f %.2f] >>\nendobj\n", width, height))

	// 4. Object 3: Page
	writeObj("3 0 obj\n<< /Type /Page /Parent 2 0 R /Resources 4 0 R /Contents 8 0 R >>\nendobj\n")

	// 5. Object 4: Resources
	writeObj("4 0 obj\n<< /Font << /F1 5 0 R /F2 6 0 R /F3 7 0 R >> >>\nendobj\n")

	// 6. Object 5: Font Helvetica
	writeObj("5 0 obj\n<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica /Encoding /WinAnsiEncoding >>\nendobj\n")

	// 7. Object 6: Font Helvetica-Bold
	writeObj("6 0 obj\n<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica-Bold /Encoding /WinAnsiEncoding >>\nendobj\n")

	// 8. Object 7: Font Helvetica-Oblique
	writeObj("7 0 obj\n<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica-Oblique /Encoding /WinAnsiEncoding >>\nendobj\n")

	// 9. Object 8: Content Stream
	streamLen := len(streamBytes)
	writeObj(fmt.Sprintf("8 0 obj\n<< /Length %d >>\nstream\n%sendstream\nendobj\n", streamLen, string(streamBytes)))

	// 10. XRef Table
	xrefOffset := out.Len()
	numObjects := len(offsets) + 1
	out.WriteString(fmt.Sprintf("xref\n0 %d\n0000000000 65535 f \n", numObjects))
	for _, off := range offsets {
		out.WriteString(fmt.Sprintf("%010d 00000 n \n", off))
	}

	// 11. Trailer
	out.WriteString(fmt.Sprintf("trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", numObjects, xrefOffset))

	return out.Bytes()
}
