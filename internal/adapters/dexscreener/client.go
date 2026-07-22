package dexscreener

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/0Hoag/cryptocheck-api/internal/adapters/etherscan"
)

type Client struct {
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// Response structures matching DexScreener API
type SearchResponse struct {
	Pairs []Pair `json:"pairs"`
}

type Pair struct {
	ChainId    string    `json:"chainId"`
	BaseToken  Token     `json:"baseToken"`
	QuoteToken Token     `json:"quoteToken"`
	Liquidity  Liquidity `json:"liquidity"`
	Volume     Volume    `json:"volume"`
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
	ContractScanSupported bool
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
	url := fmt.Sprintf("https://api.dexscreener.com/latest/dex/search?q=%s", url.QueryEscape(query))

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return Asset{}, fmt.Errorf("dexscreener request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Asset{}, fmt.Errorf("dexscreener status: %d", resp.StatusCode)
	}

	var result SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return Asset{}, fmt.Errorf("failed to decode json: %w", err)
	}

	if len(result.Pairs) == 0 {
		return Asset{}, fmt.Errorf("no token found for query: %s", query)
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
		return Asset{}, fmt.Errorf("no usable token result for query: %s", query)
	}

	// 2. Exact symbol/address/name matches always win. Liquidity breaks ties.
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].exact != candidates[j].exact {
			return candidates[i].exact
		}
		return candidates[i].pair.Liquidity.Usd > candidates[j].pair.Liquidity.Usd
	})

	topMatch := candidates[0]

	// Keep the original chain ID for market-only profiles; map supported EVM
	// chains to the identifier expected by the source-code adapters.
	network, contractScanSupported := supportedChains[topMatch.pair.ChainId]
	if !contractScanSupported {
		network = topMatch.pair.ChainId
	}

	return Asset{
		Address: topMatch.token.Address, Network: network, Name: topMatch.token.Name,
		Symbol: topMatch.token.Symbol, LiquidityUSD: topMatch.pair.Liquidity.Usd,
		VolumeH24: topMatch.pair.Volume.H24, ContractScanSupported: contractScanSupported,
	}, nil
}

func tokenMatches(token Token, query string) bool {
	query = strings.TrimSpace(query)
	return strings.EqualFold(token.Address, query) || strings.EqualFold(token.Symbol, query) || strings.EqualFold(token.Name, query)
}
