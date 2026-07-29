package http

import "testing"

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
