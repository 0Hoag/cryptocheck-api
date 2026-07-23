package dexscreener

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/0Hoag/cryptocheck-api/internal/adapters/etherscan"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    "https://api.dexscreener.com",
	}
}

// Response structures matching DexScreener API
type SearchResponse struct {
	Pairs []Pair `json:"pairs"`
}

type Pair struct {
	ChainId       string    `json:"chainId"`
	DexID         string    `json:"dexId"`
	URL           string    `json:"url"`
	PriceUSD      string    `json:"priceUsd"`
	PairCreatedAt int64     `json:"pairCreatedAt"`
	BaseToken     Token     `json:"baseToken"`
	QuoteToken    Token     `json:"quoteToken"`
	Liquidity     Liquidity `json:"liquidity"`
	Volume        Volume    `json:"volume"`
	Info          PairInfo  `json:"info"`
}

// PairInfo contains optional presentation metadata supplied by DexScreener.
type PairInfo struct {
	ImageURL string `json:"imageUrl"`
}

type Token struct {
	Address string `json:"address"`
	Name    string `json:"name"`
	Symbol  string `json:"symbol"`
}

type Liquidity struct {
	Usd float64 `json:"usd"`
}

type Volume struct {
	H24 float64 `json:"h24"`
}

// Asset is a market-discovery result. ContractScanSupported means the chain is
// currently supported by CryptoCheck's source-code security analyser.
type Asset struct {
	Address               string
	Network               string
	Name                  string
	Symbol                string
	LiquidityUSD          float64
	VolumeH24             float64
	PriceUSD              float64
	ImageURL              string
	ContractScanSupported bool
	DexID                 string
	PairURL               string
	PairCreatedAt         int64
}

// SearchTopToken finds the best matching token for a query (Symbol or Name)
// Returns: address, network, name, error
func (c *Client) SearchTopToken(query string) (string, string, string, error) {
	asset, err := c.SearchTopAsset(query)
	if err != nil {
		return "", "", "", err
	}
	if !asset.ContractScanSupported {
		return "", "", "", fmt.Errorf("found token on %s, which is not yet source-code scan supported", asset.Network)
	}
	return asset.Address, asset.Network, asset.Name, nil
}

// SearchTopAsset returns the strongest DexScreener match on any indexed chain.
// It deliberately does not imply that a security source-code scan is available.
func (c *Client) SearchTopAsset(query string) (Asset, error) {
	assets, err := c.SearchAssets(query, 1)
	if err != nil {
		return Asset{}, err
	}
	return assets[0], nil
}

// SearchAssets returns a short, de-duplicated list of the strongest matches.
// Callers can let a user choose the correct chain when a symbol is ambiguous.
func (c *Client) SearchAssets(query string, limit int) ([]Asset, error) {
	url := fmt.Sprintf("%s/latest/dex/search?q=%s", c.baseURL, url.QueryEscape(query))

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("dexscreener request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("dexscreener status: %d", resp.StatusCode)
	}

	var result SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode json: %w", err)
	}

	if len(result.Pairs) == 0 {
		return nil, fmt.Errorf("no token found for query: %s", query)
	}

	// Filter and Sort: prioritize high liquidity and specific chains
	// Supported chains in our bot: ethereum, bsc, base, arbitrum, polygon
	supportedChains := map[string]string{
		"ethereum": etherscan.NetworkETH,
		"bsc":      etherscan.NetworkBSC,
		"base":     etherscan.NetworkBase,
		"arbitrum": etherscan.NetworkArbitrum,
		"polygon":  etherscan.NetworkPolygon,
	}

	// 1. Keep every DexScreener chain and identify the token that actually
	// matches the user's query. A pair can expose the requested token as base
	// or quote; assuming base was the reason a symbol could resolve to a wrong
	// contract.
	type candidate struct {
		pair  Pair
		token Token
		exact bool
	}
	var candidates []candidate
	for _, p := range result.Pairs {
		baseExact := tokenMatches(p.BaseToken, query)
		quoteExact := tokenMatches(p.QuoteToken, query)
		switch {
		case baseExact:
			candidates = append(candidates, candidate{pair: p, token: p.BaseToken, exact: true})
		case quoteExact:
			candidates = append(candidates, candidate{pair: p, token: p.QuoteToken, exact: true})
		default:
			candidates = append(candidates, candidate{pair: p, token: p.BaseToken})
		}
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no usable token result for query: %s", query)
	}

	// 2. Exact symbol/address/name matches always win. Liquidity breaks ties.
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].exact != candidates[j].exact {
			return candidates[i].exact
		}
		return candidates[i].pair.Liquidity.Usd > candidates[j].pair.Liquidity.Usd
	})

	assets := make([]Asset, 0, limit)
	seen := make(map[string]struct{})
	for _, candidate := range candidates {
		if limit > 0 && len(assets) >= limit {
			break
		}
		key := strings.ToLower(candidate.pair.ChainId + ":" + candidate.token.Address)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}

		// Keep the original chain ID for market-only profiles; map supported EVM
		// chains to the identifier expected by the source-code adapters.
		network, contractScanSupported := supportedChains[candidate.pair.ChainId]
		if !contractScanSupported {
			network = candidate.pair.ChainId
		}

		priceUSD, _ := strconv.ParseFloat(candidate.pair.PriceUSD, 64)
		assets = append(assets, Asset{
			Address: candidate.token.Address, Network: network, Name: candidate.token.Name,
			Symbol: candidate.token.Symbol, LiquidityUSD: candidate.pair.Liquidity.Usd,
			VolumeH24: candidate.pair.Volume.H24, PriceUSD: priceUSD, ContractScanSupported: contractScanSupported,
			ImageURL: candidate.pair.Info.ImageURL, DexID: candidate.pair.DexID, PairURL: candidate.pair.URL, PairCreatedAt: candidate.pair.PairCreatedAt,
		})
	}
	if len(assets) == 0 {
		return nil, fmt.Errorf("no usable token result for query: %s", query)
	}
	return assets, nil
}

func tokenMatches(token Token, query string) bool {
	query = strings.TrimSpace(query)
	return strings.EqualFold(token.Address, query) || strings.EqualFold(token.Symbol, query) || strings.EqualFold(token.Name, query)
}
