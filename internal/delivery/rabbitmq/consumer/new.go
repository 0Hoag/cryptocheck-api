package consumer

import (
	"github.com/0Hoag/cryptocheck-api/internal/post"
	"github.com/0Hoag/cryptocheck-api/pkg/log"
	"github.com/0Hoag/cryptocheck-api/pkg/rabbitmq"
)

type Consumer struct {
	l      log.Logger
	postUC post.UseCase
	conn   *rabbitmq.Connection
}

func New(l log.Logger, conn *rabbitmq.Connection) Consumer {
	return Consumer{
		l:    l,
		conn: conn,
	}
}
