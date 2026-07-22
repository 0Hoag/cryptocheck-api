package usecase

import (
	"github.com/0Hoag/cryptocheck-api/internal/models"
	"github.com/0Hoag/cryptocheck-api/internal/resource/notification"
)

type getPostNotiContent struct {
	P            models.Post
	Type         notification.SourceType
	TaggedTarget []string
	Content      string
	TaggerName   string
}
