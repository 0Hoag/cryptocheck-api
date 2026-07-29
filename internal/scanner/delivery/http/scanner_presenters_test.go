package http

import (
	"testing"
	"time"

	"github.com/0Hoag/cryptocheck-api/internal/scanner"
)

func TestScannerTokenInputValidationAndNormalization(t *testing.T) {
	if (scannerTokenInput{Token: "   "}).validate() == nil {
		t.Fatal("whitespace-only scanner input must be rejected")
	}

	input := (scannerTokenInput{Token: "  ENA  "}).ToScanTokenInput()
	if input.Token != "ENA" || input.Language != "en" {
		t.Fatalf("unexpected normalized scanner input: %+v", input)
	}

	localized := (scannerTokenInput{Token: "BTC", Language: "vi"}).ToScanTokenInput()
	if localized.Language != "vi" {
		t.Fatalf("explicit language was lost: %+v", localized)
	}
}

func TestScannerOutputCarriesAnalysisTimestamp(t *testing.T) {
	analyzedAt := time.Date(2026, time.July, 29, 9, 0, 0, 0, time.FixedZone("UTC+7", 7*60*60))
	output := (handler{}).ToScanTokenOutput(scanner.ScanTokenOutput{Name: "ENA"}, analyzedAt)
	if !output.AnalyzedAt.Equal(analyzedAt.UTC()) {
		t.Fatalf("analysis timestamp = %v, want %v", output.AnalyzedAt, analyzedAt.UTC())
	}
}
