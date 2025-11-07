package email

import (
	"context"
	"fmt"
	"log"

	"multicarrier-email-api/internal/eml"
)

type EmailRequest struct {
	EML          eml.EML
	PayloadBytes []byte
}

type emlStorageInterface interface {
	Store(eml eml.EML) (string, error)
}

type payloadStorageInterface interface {
	Store(messageId string, payload []byte) (string, error)
}

type databaseInterface interface {
	Insert(ctx context.Context, id string, emlFilePath string, payloadPath string) error
	DeletePending(ctx context.Context, id string) error
	GetStaleEmails(ctx context.Context) ([]StaleEmail, error)
	RequeueEmail(ctx context.Context, id string) error
}

type Service struct {
	emlStorage     emlStorageInterface
	payloadStorage payloadStorageInterface
	db             databaseInterface
}

func NewService(emlStorage emlStorageInterface, payloadStorage payloadStorageInterface, db databaseInterface) *Service {
	return &Service{
		emlStorage:     emlStorage,
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
	var insertedEmls []string

	for _, req := range emailRequests {
		path, err := s.emlStorage.Store(req.EML)
		if err != nil {
			s.tryDelete(ctx, insertedEmls)
			return fmt.Errorf("failed to create EML file: %w", err)
		}

		payloadPath, err := s.payloadStorage.Store(req.EML.MessageId, req.PayloadBytes)
		if err != nil {
			s.tryDelete(ctx, insertedEmls)
			return fmt.Errorf("failed to create payload file: %w", err)
		}

		if err := s.db.Insert(ctx, req.EML.MessageId, path, payloadPath); err != nil {
			s.tryDelete(ctx, insertedEmls)
			return fmt.Errorf("failed to insert record in database: %w", err)
		}

		insertedEmls = append(insertedEmls, req.EML.MessageId)
	}

	return nil
}

func (s *Service) GetStaleEmails(ctx context.Context) ([]StaleEmail, error) {
	return s.db.GetStaleEmails(ctx)
}

func (s *Service) RequeueEmail(ctx context.Context, id string) error {
	return s.db.RequeueEmail(ctx, id)
}
