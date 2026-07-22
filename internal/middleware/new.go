package middleware

import (
	"github.com/0Hoag/cryptocheck-api/internal/users"
	"github.com/0Hoag/cryptocheck-api/pkg/log"

	pkgCrt "github.com/0Hoag/cryptocheck-api/pkg/encrypter"

	"github.com/0Hoag/cryptocheck-api/pkg/jwt"
)

type Middleware struct {
	l           log.Logger
	userUC      users.UseCase
	jwtManager  jwt.Manager
	encrypter   pkgCrt.Encrypter
	internalKey string
}

func New(
	l log.Logger,
	userUC users.UseCase,
	jwtManager jwt.Manager,
	encrypter pkgCrt.Encrypter,
	internalKey string,
) Middleware {
	return Middleware{
		l:           l,
		userUC:      userUC,
		jwtManager:  jwtManager,
		encrypter:   encrypter,
		internalKey: internalKey,
	}
}
