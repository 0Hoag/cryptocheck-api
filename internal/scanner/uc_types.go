package scanner

import (
	"github.com/0Hoag/cryptocheck-api/internal/core/scanner"
)

type ScanTokenInput struct {
	Token    string `json:"token"`
	Language string `json:"language"` // "en" or "vi"
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
