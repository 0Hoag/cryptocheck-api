package producer

import (
	"context"

	rabb "github.com/0Hoag/cryptocheck-api/internal/delivery/rabbitmq"
	"github.com/0Hoag/cryptocheck-api/pkg/log"
	"github.com/0Hoag/cryptocheck-api/pkg/rabbitmq"
)

//go:generate mockery --name=Producer
type Producer interface {
	PublishDeletePostRelationMsg(ctx context.Context, msg rabb.DeletePostRelationMsg) error
	PublishNotiMsg(ctx context.Context, msg rabb.PublishNotiMsg) error
	Run() error
	Close()
}

type implProducer struct {
	l                        log.Logger
	conn                     rabbitmq.Connection
	pushNotiWriter           *rabbitmq.Channel
	deletePostRelationWriter *rabbitmq.Channel
}

// New creates a new producer
func New(l log.Logger, conn rabbitmq.Connection) Producer {
	return &implProducer{
		l:    l,
		conn: conn,
	}
}

// NewNoop returns a producer that safely discards asynchronous messages. It is
// used only when RabbitMQ is unavailable so the HTTP API can still serve CRUD
// requests without dereferencing an empty AMQP connection.
func NewNoop() Producer {
	return noopProducer{}
}

type noopProducer struct{}

func (noopProducer) PublishDeletePostRelationMsg(context.Context, rabb.DeletePostRelationMsg) error {
	return nil
}

func (noopProducer) PublishNotiMsg(context.Context, rabb.PublishNotiMsg) error {
	return nil
}

func (noopProducer) Run() error { return nil }

func (noopProducer) Close() {}
