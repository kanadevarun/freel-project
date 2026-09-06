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

func buildQuotationPDF(quote *CustomerQuotationPreview, version int) []byte {
	c := &pdfCanvas{}

	// Page dimensions: US Letter (612 x 792 points)
	pageWidth := 612.0
	pageHeight := 792.0
	margin := 36.0
	contentWidth := pageWidth - (2 * margin) // 540 pt

	// 1. Top Header Banner
	c.rect(0, pageHeight-80, pageWidth, 80, 0.059, 0.090, 0.165, 0, 0, 0, 0) // Navy Dark #0F172A

	companyName := quote.CompanyName
	if companyName == "" {
		companyName = "LogisticsHQ"
	}
	c.text("F2", 18, margin, pageHeight-42, companyName, 1, 1, 1)
	c.text("F1", 9, margin, pageHeight-58, "OPERATING SYSTEM FOR MODERN GLOBAL FREIGHT", 0.58, 0.64, 0.72)

	c.text("F2", 14, pageWidth-margin-210, pageHeight-38, "FREIGHT QUOTATION", 0.23, 0.51, 0.96) // Blue #3B82F6
	c.text("F1", 9, pageWidth-margin-210, pageHeight-54, fmt.Sprintf("Quote #: %s   (v%d)", quote.QuotationNumber, version), 0.9, 0.9, 0.9)
	c.text("F1", 8, pageWidth-margin-210, pageHeight-68, fmt.Sprintf("Status: %s", quote.Status), 0.7, 0.8, 0.9)

	curY := pageHeight - 100.0

	// 2. Metadata / Customer / Route Grid Box
	c.rect(margin, curY-75, contentWidth, 75, 0.973, 0.980, 0.988, 0.886, 0.910, 0.941, 1) // Light Slate Box

	// Left Column: Bill To / Customer
	c.text("F2", 8, margin+12, curY-16, "CUSTOMER / BILL TO", 0.39, 0.45, 0.55)
	c.text("F2", 11, margin+12, curY-30, quote.CustomerName, 0.06, 0.09, 0.16)
	if quote.CompanyAddress != "" {
		c.text("F1", 8, margin+12, curY-43, quote.CompanyAddress, 0.28, 0.33, 0.41)
	}
	if quote.CompanyContact != "" {
		c.text("F1", 8, margin+12, curY-55, quote.CompanyContact, 0.28, 0.33, 0.41)
	}

	// Middle Column: Route & Transport
	midX := margin + 200.0
	c.text("F2", 8, midX, curY-16, "SHIPMENT & ROUTE", 0.39, 0.45, 0.55)
	c.text("F2", 9, midX, curY-30, fmt.Sprintf("Origin: %s (%s)", quote.Origin, quote.OriginCode), 0.06, 0.09, 0.16)
	c.text("F2", 9, midX, curY-43, fmt.Sprintf("Destination: %s (%s)", quote.Destination, quote.DestinationCode), 0.06, 0.09, 0.16)
	c.text("F1", 8, midX, curY-57, fmt.Sprintf("Mode: %s | Service: %s", quote.TransportMode, quote.ServiceType), 0.28, 0.33, 0.41)

	// Right Column: Validity & Terms
	rightX := margin + 370.0
	c.text("F2", 8, rightX, curY-16, "COMMERCIAL TERMS", 0.39, 0.45, 0.55)
	validFromStr := "Immediate"
	if quote.ValidFrom != nil {
		validFromStr = quote.ValidFrom.Format("2006-01-02")
	}
	validUntilStr := "Open"
	if quote.ValidUntil != nil {
		validUntilStr = quote.ValidUntil.Format("2006-01-02")
	}
	c.text("F1", 8, rightX, curY-30, fmt.Sprintf("Valid: %s to %s", validFromStr, validUntilStr), 0.06, 0.09, 0.16)
	c.text("F1", 8, rightX, curY-43, fmt.Sprintf("Payment Terms: %s", quote.PaymentTerms), 0.06, 0.09, 0.16)
	c.text("F1", 8, rightX, curY-57, fmt.Sprintf("Currency: %s", quote.Currency), 0.06, 0.09, 0.16)

	curY -= 95.0

	// 3. Itemized Charges Table
	c.text("F2", 11, margin, curY, "COMMERCIAL CHARGES BREAKDOWN", 0.06, 0.09, 0.16)
	curY -= 14.0

	// Table Header
	tableHeaderH := 20.0
	c.rect(margin, curY-tableHeaderH, contentWidth, tableHeaderH, 0.941, 0.961, 0.984, 0.886, 0.910, 0.941, 1)

	col1 := margin + 8.0   // Category / Code
	col2 := margin + 90.0  // Description
	col3 := margin + 250.0 // Basis
	col4 := margin + 320.0 // Qty
	col5 := margin + 370.0 // Unit Price
	col6 := margin + 440.0 // Tax
	col7 := margin + 490.0 // Total

	c.text("F2", 8, col1, curY-13, "CATEGORY", 0.28, 0.33, 0.41)
	c.text("F2", 8, col2, curY-13, "CHARGE DESCRIPTION", 0.28, 0.33, 0.41)
	c.text("F2", 8, col3, curY-13, "BASIS", 0.28, 0.33, 0.41)
	c.text("F2", 8, col4, curY-13, "QTY", 0.28, 0.33, 0.41)
	c.text("F2", 8, col5, curY-13, "UNIT PRICE", 0.28, 0.33, 0.41)
	c.text("F2", 8, col6, curY-13, "TAX", 0.28, 0.33, 0.41)
	c.text("F2", 8, col7, curY-13, "AMOUNT", 0.28, 0.33, 0.41)

	curY -= tableHeaderH

	rowH := 20.0
	for i, ch := range quote.Charges {
		if curY < 180 {
			// Limit to avoid overflow on single page for now
			break
		}
		bgR, bgG, bgB := 1.0, 1.0, 1.0
		if i%2 == 1 {
			bgR, bgG, bgB = 0.984, 0.988, 0.992
		}
		c.rect(margin, curY-rowH, contentWidth, rowH, bgR, bgG, bgB, 0.925, 0.941, 0.961, 0.5)

		c.text("F1", 8, col1, curY-13, ch.ChargeCategory, 0.28, 0.33, 0.41)
		c.text("F2", 8, col2, curY-13, ch.ChargeName, 0.06, 0.09, 0.16)
		c.text("F1", 8, col3, curY-13, ch.CalculationBasis, 0.39, 0.45, 0.55)
		c.text("F1", 8, col4, curY-13, fmt.Sprintf("%.2f", ch.Quantity), 0.06, 0.09, 0.16)
		c.text("F1", 8, col5, curY-13, fmt.Sprintf("%.2f", ch.UnitPrice), 0.06, 0.09, 0.16)
		taxStr := "-"
		if ch.TaxRate > 0 {
			taxStr = fmt.Sprintf("%.1f%%", ch.TaxRate)
		}
		c.text("F1", 8, col6, curY-13, taxStr, 0.39, 0.45, 0.55)
		c.text("F2", 8, col7, curY-13, fmt.Sprintf("%.2f", ch.FinalAmount), 0.06, 0.09, 0.16)

		curY -= rowH
	}

	curY -= 12.0

	// 4. Totals & Notes Section
	// Left: Commercial Terms & Customer Notes
	termsWidth := 320.0
	c.rect(margin, curY-90, termsWidth, 90, 0.984, 0.988, 0.992, 0.886, 0.910, 0.941, 1)
	c.text("F2", 8, margin+10, curY-16, "COMMERCIAL TERMS & NOTES", 0.28, 0.33, 0.41)

	commTerms := quote.CommercialTerms
	if commTerms == "" {
		commTerms = "Subject to standard carrier terms & conditions and space availability."
	}
	c.text("F1", 7.5, margin+10, curY-30, truncateStr(commTerms, 75), 0.28, 0.33, 0.41)

	custNotes := quote.CustomerNotes
	if custNotes == "" {
		custNotes = "Rates valid for cargo described. Fuel & security surcharges included as listed."
	}
	c.text("F2", 8, margin+10, curY-48, "CUSTOMER INSTRUCTIONS:", 0.39, 0.45, 0.55)
	c.text("F1", 7.5, margin+10, curY-60, truncateStr(custNotes, 75), 0.28, 0.33, 0.41)

	c.text("F3", 7, margin+10, curY-78, "Strict Customer-Safe Commercial Copy — No Internal Surcharges Apply", 0.58, 0.64, 0.72)

	// Right: Financial Summary Box
	totalsX := margin + termsWidth + 15.0
	totalsWidth := contentWidth - termsWidth - 15.0
	c.rect(totalsX, curY-90, totalsWidth, 90, 0.973, 0.980, 0.988, 0.886, 0.910, 0.941, 1)

	c.text("F1", 8, totalsX+10, curY-18, "Subtotal Charges:", 0.39, 0.45, 0.55)
	c.text("F2", 8, totalsX+totalsWidth-65, curY-18, fmt.Sprintf("%s %.2f", quote.Currency, quote.Subtotal), 0.06, 0.09, 0.16)

	if quote.DiscountTotal > 0 {
		c.text("F1", 8, totalsX+10, curY-32, "Total Discounts:", 0.86, 0.15, 0.15)
		c.text("F2", 8, totalsX+totalsWidth-65, curY-32, fmt.Sprintf("-%s %.2f", quote.Currency, quote.DiscountTotal), 0.86, 0.15, 0.15)
	}

	c.text("F1", 8, totalsX+10, curY-46, "Applicable Taxes:", 0.39, 0.45, 0.55)
	c.text("F2", 8, totalsX+totalsWidth-65, curY-46, fmt.Sprintf("%s %.2f", quote.Currency, quote.TaxTotal), 0.06, 0.09, 0.16)

	c.line(totalsX+8, curY-56, totalsX+totalsWidth-8, curY-56, 0.8, 0.85, 0.9, 1)

	// Grand Total Highlight
	c.rect(totalsX+6, curY-84, totalsWidth-12, 24, 0.145, 0.388, 0.922, 0, 0, 0, 0) // Blue #2563EB
	c.text("F2", 9, totalsX+12, curY-72, "GRAND TOTAL:", 1, 1, 1)
	c.text("F2", 11, totalsX+totalsWidth-80, curY-72, fmt.Sprintf("%s %.2f", quote.Currency, quote.TotalAmount), 1, 1, 1)

	// 5. Bottom Acceptance & Signature Box
	curY -= 105.0
	c.line(margin, curY, pageWidth-margin, curY, 0.886, 0.910, 0.941, 1)

	c.text("F2", 8, margin, curY-14, "CUSTOMER AUTHORIZATION / ACCEPTANCE", 0.06, 0.09, 0.16)
	c.text("F1", 7.5, margin, curY-26, "To accept this quotation, please sign below or confirm via the secure customer portal link.", 0.39, 0.45, 0.55)

	c.text("F1", 8, margin, curY-48, "Authorized Signature: ___________________________", 0.28, 0.33, 0.41)
	c.text("F1", 8, margin+300, curY-48, "Date: _________________________", 0.28, 0.33, 0.41)

	// Footer Stamp
	c.text("F1", 7, margin, 18, fmt.Sprintf("Generated on %s | Freight OS LogisticsHQ Platform | Quotation # %s", time.Now().Format("2006-01-02 15:04:05"), quote.QuotationNumber), 0.6, 0.65, 0.72)

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
