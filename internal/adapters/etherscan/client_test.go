package etherscan

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetContractSourceParsesExplorerResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "1", r.URL.Query().Get("chainid"))
		require.Equal(t, "contract", r.URL.Query().Get("module"))
		require.Equal(t, "getsourcecode", r.URL.Query().Get("action"))
		require.Equal(t, "0xabc", r.URL.Query().Get("address"))
		require.Equal(t, "test-key", r.URL.Query().Get("apikey"))
		_, _ = fmt.Fprint(w, `{"status":"1","message":"OK","result":[{"SourceCode":"contract Token {}","ContractName":"Token","ABI":"[]"}]}`)
	}))
	defer server.Close()

	client := &Client{apiKeys: map[string]string{NetworkETH: "test-key"}, baseURLs: map[string]string{NetworkETH: server.URL}, httpClient: server.Client()}
	source, name, err := client.GetContractSource(NetworkETH, "0xabc")

	require.NoError(t, err)
	require.Equal(t, "contract Token {}", source)
	require.Equal(t, "Token", name)
}

func TestGetContractSourceReturnsExplorerErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, `{"status":"0","message":"NOTOK","result":"Contract source code not verified"}`)
	}))
	defer server.Close()

	client := &Client{apiKeys: map[string]string{NetworkBSC: "test-key"}, baseURLs: map[string]string{NetworkBSC: server.URL}, httpClient: server.Client()}
	_, _, err := client.GetContractSource(NetworkBSC, "0xabc")

	require.ErrorContains(t, err, "Contract source code not verified")
}
