package usecase

import (
	"context"
	"strings"

	"github.com/0Hoag/cryptocheck-api/internal/adapters/dexscreener"
	"github.com/0Hoag/cryptocheck-api/internal/adapters/etherscan"
	coreScanner "github.com/0Hoag/cryptocheck-api/internal/core/scanner"
	scanDomain "github.com/0Hoag/cryptocheck-api/internal/scanner"
)

func (uc ScannerUC) ScanToken(ctx context.Context, input scanDomain.ScanTokenInput) (scanDomain.ScanTokenOutput, error) {
	query := strings.TrimSpace(input.Token)
	if native, ok := nativeAssetReports[strings.ToUpper(query)]; ok {
		return native.toOutput(), nil
	}
	address := query
	network := "eth"
	name := query
	var marketAsset *dexscreener.Asset

	// 1. Resolve Symbol if needed
	isAddress := (strings.HasPrefix(query, "0x") && len(query) == 42) || query == "0xMOCK"
	if !isAddress {
		asset, err := uc.dexClient.SearchTopAsset(query)
		if err != nil {
			uc.l.Errorf(ctx, "Token not found on DexScreener: %v", err)
			return scanDomain.ScanTokenOutput{}, scanDomain.ErrTokenNotFound
		}
		marketAsset = &asset
		address = asset.Address
		network = asset.Network
		name = asset.Name
		if !asset.ContractScanSupported {
			return marketProfile(asset), nil
		}
	}

	// 2. Fetch Source Code (Try all networks like Telegram Bot)
	var sourceCode string
	var err error
	var networkFound string

	networks := networksToTry(network)

	for _, net := range networks {
		sourceCode, name, err = uc.ethClient.GetContractSource(net, address)
		if err == nil && sourceCode != "" {
			networkFound = net
			break
		}
	}

	if networkFound == "" {
		if marketAsset != nil {
			return marketProfile(*marketAsset), nil
		}
		uc.l.Errorf(ctx, "scanner.usecase.scanner.ScanToken: source code not found on any network")
		return scanDomain.ScanTokenOutput{}, scanDomain.ErrSourceCodeNotFound
	}

	network = networkFound

	// 3. Analyze (pass language preference)
	result := uc.engine.Scan(sourceCode, address, input.Language)

	return scanDomain.ScanTokenOutput{
		Network:         network,
		Name:            name,
		Address:         address,
		AnalysisType:    "contract",
		SourceAvailable: true,
		ScoreAvailable:  true,
		TrustScore:      result.TrustScore,
		Issues:          result.Issues,
		SafeFeatures:    result.SafeFeatures,
	}, nil
}

func (uc ScannerUC) FindCandidates(ctx context.Context, input scanDomain.FindCandidatesInput) ([]scanDomain.TokenCandidate, error) {
	query := strings.TrimSpace(input.Query)
	assets, err := uc.dexClient.SearchAssets(query, 5)
	if err != nil {
		uc.l.Errorf(ctx, "scanner.usecase.scanner.FindCandidates: %v", err)
		return nil, scanDomain.ErrTokenNotFound
	}
	candidates := make([]scanDomain.TokenCandidate, 0, len(assets))
	for _, asset := range assets {
		candidates = append(candidates, scanDomain.TokenCandidate{
			Address: asset.Address, Network: asset.Network, Name: asset.Name, Symbol: asset.Symbol,
			LiquidityUSD: asset.LiquidityUSD, VolumeH24: asset.VolumeH24, ContractScanSupported: asset.ContractScanSupported,
			DexID: asset.DexID, PairURL: asset.PairURL, PairCreatedAt: asset.PairCreatedAt,
		})
	}
	return candidates, nil
}

func marketProfile(asset dexscreener.Asset) scanDomain.ScanTokenOutput {
	return scanDomain.ScanTokenOutput{
		Network:          asset.Network,
		Name:             asset.Name,
		Address:          asset.Address,
		AnalysisType:     "market_asset",
		SourceAvailable:  false,
		ScoreAvailable:   false,
		LiquidityUSD:     asset.LiquidityUSD,
		VolumeH24:        asset.VolumeH24,
		MarketProvider:   "DexScreener",
		DexID:            asset.DexID,
		PairURL:          asset.PairURL,
		PairCreatedAt:    asset.PairCreatedAt,
		MarketConfidence: marketConfidence(asset.LiquidityUSD, asset.VolumeH24),
		Issues: []coreScanner.Issue{{
			Type:        coreScanner.IssueInfo,
			Name:        "Market profile only",
			Description: "This asset was found on DexScreener, but its chain or source code is not currently supported for a contract security scan.",
			Impact:      0,
		}},
		SafeFeatures: []string{"DexScreener market pair found", "Liquidity and 24h volume are available"},
	}
}

func marketConfidence(liquidityUSD, volumeH24 float64) string {
	if liquidityUSD >= 100_000 && volumeH24 >= 10_000 {
		return "high"
	}
	if liquidityUSD >= 10_000 || volumeH24 >= 1_000 {
		return "medium"
	}
	return "low"
}

type nativeAssetReport struct {
	Symbol       string
	Name         string
	Network      string
	TrustScore   int
	Issues       []coreScanner.Issue
	SafeFeatures []string
}

func (r nativeAssetReport) toOutput() scanDomain.ScanTokenOutput {
	return scanDomain.ScanTokenOutput{
		Network: r.Network, Name: r.Name, Address: r.Symbol, AnalysisType: "native_asset", SourceAvailable: false,
		ScoreAvailable: true, TrustScore: r.TrustScore, Issues: r.Issues, SafeFeatures: r.SafeFeatures,
	}
}

var nativeAssetReports = map[string]nativeAssetReport{
	"BTC": {Symbol: "BTC", Name: "Bitcoin", Network: "bitcoin", TrustScore: 92,
		Issues:       []coreScanner.Issue{{Type: coreScanner.IssueInfo, Name: "Native asset", Description: "BTC runs on the Bitcoin network and has no EVM smart-contract source code to inspect.", Impact: 0}},
		SafeFeatures: []string{"Native Bitcoin asset", "Proof-of-work network", "No token admin contract"}},
	"ETH": {Symbol: "ETH", Name: "Ethereum", Network: "ethereum", TrustScore: 90,
		Issues:       []coreScanner.Issue{{Type: coreScanner.IssueInfo, Name: "Native asset", Description: "ETH is Ethereum's native asset, so contract source-code checks do not apply.", Impact: 0}},
		SafeFeatures: []string{"Native Ethereum asset", "No token admin contract", "Network gas asset"}},
	"BNB": {Symbol: "BNB", Name: "BNB", Network: "bsc", TrustScore: 88,
		Issues:       []coreScanner.Issue{{Type: coreScanner.IssueInfo, Name: "Native asset", Description: "BNB is the BNB Smart Chain native asset; use a token contract address to run source-code checks.", Impact: 0}},
		SafeFeatures: []string{"Native BNB Chain asset", "No token admin contract"}},
	"SOL": {Symbol: "SOL", Name: "Solana", Network: "solana", TrustScore: 88,
		Issues:       []coreScanner.Issue{{Type: coreScanner.IssueInfo, Name: "Native asset", Description: "SOL is Solana's native asset and cannot be evaluated with EVM contract-source rules.", Impact: 0}},
		SafeFeatures: []string{"Native Solana asset", "No EVM token contract"}},
}

func networksToTry(preferred string) []string {
	all := []string{etherscan.NetworkETH, etherscan.NetworkBSC, etherscan.NetworkBase, etherscan.NetworkArbitrum, etherscan.NetworkPolygon}
	result := make([]string, 0, len(all))
	if preferred != "" {
		result = append(result, preferred)
	}
	for _, network := range all {
		if network != preferred {
			result = append(result, network)
		}
	}
	return result
}
