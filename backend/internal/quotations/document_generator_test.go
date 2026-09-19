package quotations

import (
	"bytes"
	"context"
	"os"
	"testing"
	"time"
)

func TestDocumentGenerator_GenerateQuotationPDF(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "quote_pdf_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	gen := NewDocumentGenerator(tempDir)

	validUntil := time.Now().Add(7 * 24 * time.Hour)
	quote := &CustomerQuotationPreview{
		QuotationID:     101,
		QuotationNumber: "QT-2026-TEST-001",
		Status:          "APPROVED",
		ValidUntil:      &validUntil,
		Currency:        "USD",
		CustomerName:    "Global Acme Logistics Ltd",
		Origin:          "Shanghai Port",
		OriginCode:      "CNSHA",
		Destination:     "Los Angeles",
		DestinationCode: "USLAX",
		ServiceType:     "PORT_TO_PORT",
		TransportMode:   "OCEAN",
		PaymentTerms:    "NET_30",
		CommercialTerms: "Standard maritime FOB commercial terms.",
		CustomerNotes:   "Handle with care.",
		CompanyName:     "LogisticsHQ Global Forwarding",
		CompanyAddress:  "100 Ocean Blvd, Suite 400",
		CompanyContact:  "support@logisticshq.in",
		CompanyLogoURL:  "https://freel-bucket.s3.ap-south-1.amazonaws.com/organizations/1/branding/logo/brand.png",
		Charges: []CustomerQuotationChargeItem{
			{
				ChargeCode:     "BAS",
				ChargeName:     "Ocean Base Freight 40ft HC",
				ChargeCategory: "FREIGHT",
				Quantity:       1,
				UnitPrice:      3200.00,
				FinalAmount:    3200.00,
				Currency:       "USD",
			},
			{
				ChargeCode:     "THC",
				ChargeName:     "Origin Terminal Handling Charge",
				ChargeCategory: "ORIGIN",
				Quantity:       1,
				UnitPrice:      280.00,
				FinalAmount:    280.00,
				Currency:       "USD",
			},
		},
		Subtotal:    3480.00,
		TotalAmount: 3480.00,
	}

	ctx := context.Background()
	filePath, fileName, pdfBytes, err := gen.GenerateQuotationPDF(ctx, quote, 1)
	if err != nil {
		t.Fatalf("GenerateQuotationPDF failed: %v", err)
	}

	if fileName != "Quote_QT-2026-TEST-001_v1.pdf" {
		t.Errorf("unexpected filename: %s", fileName)
	}

	if len(pdfBytes) == 0 {
		t.Fatal("generated PDF bytes are empty")
	}

	// Verify standard PDF header %PDF-1.4
	if !bytes.HasPrefix(pdfBytes, []byte("%PDF-1.4")) {
		t.Errorf("PDF does not start with %%PDF-1.4 header: %s", string(pdfBytes[:8]))
	}

	// Verify PDF EOF marker
	if !bytes.Contains(pdfBytes, []byte("%%EOF")) {
		t.Error("PDF does not contain EOF trailer marker")
	}

	// Verify quotation metadata is rendered in PDF stream
	if !bytes.Contains(pdfBytes, []byte("QT-2026-TEST-001")) {
		t.Error("PDF does not contain quotation number")
	}
	if !bytes.Contains(pdfBytes, []byte("Global Acme Logistics Ltd")) {
		t.Error("PDF does not contain customer name")
	}
	if !bytes.Contains(pdfBytes, []byte("BRAND LOGO")) {
		t.Error("PDF does not contain BRAND LOGO badge")
	}
	if !bytes.Contains(pdfBytes, []byte("LogisticsHQ Global Forwarding")) {
		t.Error("PDF does not contain company name")
	}

	// Verify file was written to disk
	info, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("failed to stat generated file: %v", err)
	}
	if info.Size() == 0 {
		t.Error("file on disk has zero size")
	}
}
