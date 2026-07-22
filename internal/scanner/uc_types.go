package scanner

import (
	"github.com/0Hoag/cryptocheck-api/internal/core/scanner"
)

type ScanTokenInput struct {
	Token    string `json:"token"`
	Language string `json:"language"` // "en" or "vi"
}

type FindCandidatesInput struct {
	Query string `json:"query"`
}

type TokenCandidate struct {
	Address               string  `json:"address"`
	Network               string  `json:"network"`
	Name                  string  `json:"name"`
	Symbol                string  `json:"symbol"`
	LiquidityUSD          float64 `json:"liquidity_usd"`
	VolumeH24             float64 `json:"volume_h24"`
	ContractScanSupported bool    `json:"contract_scan_supported"`
}

type ScanTokenOutput struct {
	Network         string          `json:"network"`
	Name            string          `json:"name"`
	Address         string          `json:"address"`
	AnalysisType    string          `json:"analysis_type"`
	SourceAvailable bool            `json:"source_available"`
	ScoreAvailable  bool            `json:"score_available"`
	TrustScore      int             `json:"trust_score"`
	LiquidityUSD    float64         `json:"liquidity_usd,omitempty"`
	VolumeH24       float64         `json:"volume_h24,omitempty"`
	Issues          []scanner.Issue `json:"issues"`
	SafeFeatures    []string        `json:"safe_features"`
}
