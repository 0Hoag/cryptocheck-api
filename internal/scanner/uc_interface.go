package scanner

import (
	"context"
)

//go:generate mockery --name=Usecase
type UseCase interface {
	ScanToken(ctx context.Context, input ScanTokenInput) (ScanTokenOutput, error)
	FindCandidates(ctx context.Context, input FindCandidatesInput) ([]TokenCandidate, error)
}
