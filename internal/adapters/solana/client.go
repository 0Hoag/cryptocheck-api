package solana

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const defaultRPCURL = "https://api.mainnet-beta.solana.com"

// Mint is the limited set of on-chain SPL metadata needed for a basic token
// authority report. It is not a full program audit.
type Mint struct {
	MintAuthority   string
	FreezeAuthority string
	Supply          string
	Decimals        int
}

type Client struct {
	rpcURL     string
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{rpcURL: defaultRPCURL, httpClient: &http.Client{Timeout: 12 * time.Second}}
}

func (c *Client) GetMint(ctx context.Context, address string) (Mint, error) {
	payload, err := json.Marshal(map[string]any{
		"jsonrpc": "2.0", "id": 1, "method": "getAccountInfo",
		"params": []any{address, map[string]string{"encoding": "jsonParsed"}},
	})
	if err != nil {
		return Mint{}, fmt.Errorf("encode solana request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.rpcURL, bytes.NewReader(payload))
	if err != nil {
		return Mint{}, fmt.Errorf("create solana request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Mint{}, fmt.Errorf("solana rpc request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Mint{}, fmt.Errorf("solana rpc status: %d", resp.StatusCode)
	}

	var body struct {
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
		Result struct {
			Value *struct {
				Data struct {
					Parsed struct {
						Type string `json:"type"`
						Info struct {
							MintAuthority   *string `json:"mintAuthority"`
							FreezeAuthority *string `json:"freezeAuthority"`
							Supply          string  `json:"supply"`
							Decimals        int     `json:"decimals"`
						} `json:"info"`
					} `json:"parsed"`
				} `json:"data"`
			} `json:"value"`
		} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return Mint{}, fmt.Errorf("decode solana response: %w", err)
	}
	if body.Error != nil {
		return Mint{}, fmt.Errorf("solana rpc: %s", body.Error.Message)
	}
	if body.Result.Value == nil || body.Result.Value.Data.Parsed.Type != "mint" {
		return Mint{}, fmt.Errorf("address is not an SPL token mint")
	}
	info := body.Result.Value.Data.Parsed.Info
	mint := Mint{Supply: info.Supply, Decimals: info.Decimals}
	if info.MintAuthority != nil {
		mint.MintAuthority = *info.MintAuthority
	}
	if info.FreezeAuthority != nil {
		mint.FreezeAuthority = *info.FreezeAuthority
	}
	return mint, nil
}
