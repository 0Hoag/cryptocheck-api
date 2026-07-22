package http

import (
	scan "github.com/0Hoag/cryptocheck-api/internal/core/scanner"
	"github.com/0Hoag/cryptocheck-api/internal/scanner"
)

type scannerTokenInput struct {
	Token    string `form:"token"`
	Language string `form:"lang"` // From query param or header
}

func (r scannerTokenInput) ToScanTokenInput() scanner.ScanTokenInput {
	lang := r.Language
	if lang == "" {
		lang = "en" // Default to English
	}
	return scanner.ScanTokenInput{
		Token:    r.Token,
		Language: lang,
	}
}

func (r scannerTokenInput) validate() error {
	if len(r.Token) == 0 {
		return errWrongBody
	}

	return nil
}

type issue struct {
	Type        scan.IssueType `json:"type"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Impact      int            `json:"impact"`
}

func toIssues(issues []scan.Issue) []issue {
	result := []issue{}
	for _, i := range issues {
		result = append(result, issue{
			Type:        i.Type,
			Name:        i.Name,
			Description: i.Description,
			Impact:      i.Impact,
		})
	}
	return result
}

type scannerTokenOutput struct {
	Network         string   `json:"network"`
	Name            string   `json:"name"`
	Address         string   `json:"address"`
	AnalysisType    string   `json:"analysis_type"`
	SourceAvailable bool     `json:"source_available"`
	ScoreAvailable  bool     `json:"score_available"`
	TrustScore      int      `json:"trust_score"`
	LiquidityUSD    float64  `json:"liquidity_usd,omitempty"`
	VolumeH24       float64  `json:"volume_h24,omitempty"`
	Issues          []issue  `json:"issues"`
	SafeFeatures    []string `json:"safe_features"`
}

func (h handler) ToScanTokenOutput(token scanner.ScanTokenOutput) scannerTokenOutput {
	return scannerTokenOutput{
		Network:         token.Network,
		Name:            token.Name,
		Address:         token.Address,
		AnalysisType:    token.AnalysisType,
		SourceAvailable: token.SourceAvailable,
		ScoreAvailable:  token.ScoreAvailable,
		TrustScore:      token.TrustScore,
		LiquidityUSD:    token.LiquidityUSD,
		VolumeH24:       token.VolumeH24,
		Issues:          toIssues(token.Issues),
		SafeFeatures:    token.SafeFeatures,
	}
}
