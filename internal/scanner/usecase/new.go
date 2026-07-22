package usecase

import (
	"github.com/0Hoag/cryptocheck-api/internal/adapters/dexscreener"
	"github.com/0Hoag/cryptocheck-api/internal/adapters/etherscan"
	"github.com/0Hoag/cryptocheck-api/internal/adapters/solana"
	"github.com/0Hoag/cryptocheck-api/internal/core/scanner"
	pkgLog "github.com/0Hoag/cryptocheck-api/pkg/log"
)

type ScannerUC struct {
	l         pkgLog.Logger
	engine    *scanner.Engine
	dexClient *dexscreener.Client
	ethClient *etherscan.Client
	solClient *solana.Client
}

func New(l pkgLog.Logger, engine *scanner.Engine, dex *dexscreener.Client, eth *etherscan.Client) *ScannerUC {
	return &ScannerUC{
		l:         l,
		engine:    engine,
		dexClient: dex,
		ethClient: eth,
		solClient: solana.NewClient(),
	}
}
