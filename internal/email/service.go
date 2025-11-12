package email

import (
	"context"
	"fmt"
	"log"
)

type EmailRequest struct {
	MessageId    string
	PayloadBytes []byte
}


type payloadStorageInterface interface {
	Store(messageId string, payload []byte) (string, error)
}

type databaseInterface interface {
	Insert(ctx context.Context, id string, payloadPath string) error
	DeletePending(ctx context.Context, id string) error
	GetStaleEmails(ctx context.Context) ([]Email, error)
	GetInvalidEmails(ctx context.Context) ([]Email, error)
	RequeueEmail(ctx context.Context, id string) error
}

type Service struct {
	payloadStorage payloadStorageInterface
	db             databaseInterface
}

func NewService(payloadStorage payloadStorageInterface, db databaseInterface) *Service {
	return &Service{
		payloadStorage: payloadStorage,
		db:             db,
	}
}

func (s *Service) tryDelete(ctx context.Context, ids []string) {
	for _, i := range ids {
		if deleteErr := s.db.DeletePending(ctx, i); deleteErr != nil {
			log.Printf("failed to delete pending email '%s': %v", i, deleteErr)
		}
	}
}

func (s *Service) Save(ctx context.Context, emailRequests []EmailRequest) error {
	var insertedIds []string

	for _, req := range emailRequests {
		payloadPath, err := s.payloadStorage.Store(req.MessageId, req.PayloadBytes)
		if err != nil {
			s.tryDelete(ctx, insertedIds)
			return fmt.Errorf("failed to create payload file: %w", err)
		}

		if err := s.db.Insert(ctx, req.MessageId, payloadPath); err != nil {
			s.tryDelete(ctx, insertedIds)
			return fmt.Errorf("failed to insert record in database: %w", err)
		}

		insertedIds = append(insertedIds, req.MessageId)
	}

	return nil
}

func (s *Service) GetStaleEmails(ctx context.Context) ([]Email, error) {
	return s.db.GetStaleEmails(ctx)
}

func (s *Service) GetInvalidEmails(ctx context.Context) ([]Email, error) {
	return s.db.GetInvalidEmails(ctx)
}

func (s *Service) RequeueEmail(ctx context.Context, id string) error {
	return s.db.RequeueEmail(ctx, id)
}
