package usecase

import (
	"context"
	"testing"

	"github.com/0Hoag/cryptocheck-api/internal/adapters/solana"
	scanDomain "github.com/0Hoag/cryptocheck-api/internal/scanner"
)

const validSolanaMint = "So11111111111111111111111111111111111111112"

type fixedSolanaMintClient struct {
	mint    solana.Mint
	address string
}

func (c *fixedSolanaMintClient) GetMint(_ context.Context, address string) (solana.Mint, error) {
	c.address = address
	return c.mint, nil
}

func TestIsSolanaMintAddress(t *testing.T) {
	tests := []struct {
		value string
		want  bool
	}{
		{value: validSolanaMint, want: true},
		{value: "0x0000000000000000000000000000000000000000", want: false},
		{value: "SOL", want: false},
		{value: "O0Il", want: false},
	}
	for _, tt := range tests {
		if got := isSolanaMintAddress(tt.value); got != tt.want {
			t.Errorf("isSolanaMintAddress(%q) = %t, want %t", tt.value, got, tt.want)
		}
	}
}

func TestScanTokenDirectSolanaMintUsesOnChainMintLookup(t *testing.T) {
	client := &fixedSolanaMintClient{mint: solana.Mint{Supply: "1000000", Decimals: 6}}
	uc := ScannerUC{solClient: client}

	report, err := uc.ScanToken(context.Background(), scanDomain.ScanTokenInput{Token: validSolanaMint})
	if err != nil {
		t.Fatalf("ScanToken() error = %v", err)
	}
	if client.address != validSolanaMint {
		t.Fatalf("mint lookup address = %q, want %q", client.address, validSolanaMint)
	}
	if report.AnalysisType != "solana_mint" || report.Network != "solana" {
		t.Fatalf("unexpected direct mint report: %+v", report)
	}
	if report.Name != "Solana SPL mint" || !report.ScoreAvailable {
		t.Fatalf("direct mint report must identify its limited on-chain scope: %+v", report)
	}
}
