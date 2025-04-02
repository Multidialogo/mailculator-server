package email

import (
	"context"
	"fmt"
	"log"

	"multicarrier-email-api/internal/eml"
)

type emlStorageInterface interface {
	Store(eml eml.EML) (string, error)
}

type databaseInterface interface {
	Insert(ctx context.Context, id string, emlFilePath string) error
	DeletePending(ctx context.Context, id string) error
}

type Service struct {
	emlStorage emlStorageInterface
	db         databaseInterface
}

func NewService(emlStorage emlStorageInterface, db databaseInterface) *Service {
	return &Service{
		emlStorage: emlStorage,
		db:         db,
	}
}

func (s *Service) tryDelete(ctx context.Context, ids []string) {
	for _, i := range ids {
		if deleteErr := s.db.DeletePending(ctx, i); deleteErr != nil {
			log.Printf("failed to delete pending email '%s': %v", i, deleteErr)
		}
	}
}

func (s *Service) Save(ctx context.Context, in []eml.EML) error {
	var inserted []string

	for _, emlData := range in {
		path, err := s.emlStorage.Store(emlData)
		if err != nil {
			s.tryDelete(ctx, inserted)
			return fmt.Errorf("failed to create EML file: %w", err)
		}

		if err := s.db.Insert(ctx, emlData.MessageId, path); err != nil {
			s.tryDelete(ctx, inserted)
			return fmt.Errorf("failed to insert record in database: %w", err)
		}

		inserted = append(inserted, emlData.MessageId)
	}

	return nil
}
