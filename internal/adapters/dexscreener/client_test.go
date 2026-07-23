package dexscreener

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSearchAssetsUsesExactMatchesAndPreservesMarketChains(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/latest/dex/search", r.URL.Path)
		require.Equal(t, "ENA", r.URL.Query().Get("q"))
		_, _ = fmt.Fprint(w, `{"pairs":[
          {"chainId":"ethereum","dexId":"uniswap","url":"https://dex/quote","pairCreatedAt":2,"baseToken":{"address":"0xnot","name":"Other","symbol":"NOT"},"quoteToken":{"address":"0xquote-ena","name":"Ethena","symbol":"ENA"},"liquidity":{"usd":500000},"volume":{"h24":90000}},
          {"chainId":"ethereum","dexId":"uniswap","url":"https://dex/base","pairCreatedAt":1,"baseToken":{"address":"0xbase-ena","name":"Ethena","symbol":"ENA"},"quoteToken":{"address":"0xusdc","name":"USD Coin","symbol":"USDC"},"liquidity":{"usd":120000},"volume":{"h24":50000}},
          {"chainId":"ethereum","dexId":"uniswap","url":"https://dex/duplicate","pairCreatedAt":3,"baseToken":{"address":"0xbase-ena","name":"Ethena","symbol":"ENA"},"quoteToken":{"address":"0xusdt","name":"Tether","symbol":"USDT"},"liquidity":{"usd":100000},"volume":{"h24":40000}},
          {"chainId":"solana","dexId":"raydium","url":"https://dex/sol","pairCreatedAt":4,"baseToken":{"address":"sol-ena","name":"Ethena Solana","symbol":"ENA"},"quoteToken":{"address":"sol-usdc","name":"USD Coin","symbol":"USDC"},"liquidity":{"usd":20000},"volume":{"h24":3000}}
        ]}`)
	}))
	defer server.Close()

	client := &Client{httpClient: server.Client(), baseURL: server.URL}
	assets, err := client.SearchAssets("ENA", 5)

	require.NoError(t, err)
	require.Len(t, assets, 3)
	require.Equal(t, "0xquote-ena", assets[0].Address, "exact quote-token match with highest liquidity wins")
	require.Equal(t, "eth", assets[0].Network)
	require.True(t, assets[0].ContractScanSupported)
	require.Equal(t, "0xbase-ena", assets[1].Address)
	require.Equal(t, "sol-ena", assets[2].Address)
	require.Equal(t, "solana", assets[2].Network)
	require.False(t, assets[2].ContractScanSupported)
}

func TestSearchAssetsReturnsProviderErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client := &Client{httpClient: server.Client(), baseURL: server.URL}
	_, err := client.SearchAssets("ENA", 1)
	require.ErrorContains(t, err, "dexscreener status: 429")
}
