package email

import (
	"context"
	"errors"
	"log"
)

const (
	ErrorCodeDuplicatedID   = "DUPLICATED_ID"
	ErrorCodeStorageError   = "STORAGE_ERROR"
	ErrorCodeDatabaseError  = "DATABASE_ERROR"
	ErrorCodeTransientError = "TRANSIENT_ERROR"
)

const (
	ErrorMessageDuplicatedID   = "Email with this ID already exists"
	ErrorMessageStorageError   = "Failed to store email payload"
	ErrorMessageDatabaseError  = "Failed to save email to database"
	ErrorMessageTransientError = "Temporary database error, retry possible"
)

type EmailRequest struct {
	MessageId    string
	PayloadBytes []byte
}

type SaveResult struct {
	MessageId    string
	Success      bool
	ErrorCode    string
	ErrorMessage string
}

type storageError struct {
	err error
}

func (e *storageError) Error() string { return e.err.Error() }
func (e *storageError) Unwrap() error { return e.err }

type payloadStorageInterface interface {
	ResolvePath(messageId string) (string, error)
	Write(path string, payload []byte) error
	Delete(payloadPath string) error
}

type databaseInterface interface {
	Insert(ctx context.Context, id string, payloadPath string, onEmailInserted func() error) error
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

func (s *Service) tryDelete(payloadPath string) {
	if deleteErr := s.payloadStorage.Delete(payloadPath); deleteErr != nil {
		log.Printf("failed to delete payload file '%s': %v", payloadPath, deleteErr)
	}
}

func (s *Service) Save(ctx context.Context, emailRequests []EmailRequest) []SaveResult {
	results := make([]SaveResult, len(emailRequests))

	for i, req := range emailRequests {
		result := SaveResult{
			MessageId: req.MessageId,
			Success:   true,
		}

		payloadPath, err := s.payloadStorage.ResolvePath(req.MessageId)
		if err != nil {
			log.Printf("failed to resolve payload path for '%s': %v", req.MessageId, err)
			result.Success = false
			result.ErrorCode = ErrorCodeStorageError
			result.ErrorMessage = ErrorMessageStorageError
			results[i] = result
			continue
		}

		err = s.db.Insert(ctx, req.MessageId, payloadPath, func() error {
			if writeErr := s.payloadStorage.Write(payloadPath, req.PayloadBytes); writeErr != nil {
				return &storageError{err: writeErr}
			}
			return nil
		})

		if err != nil {
			var cErr *commitError
			if errors.As(err, &cErr) {
				s.tryDelete(payloadPath)
			}

			log.Printf("failed to save email '%s': %v", req.MessageId, err)
			result.Success = false

			var stErr *storageError
			if IsDuplicateEntryError(err) {
				result.ErrorCode = ErrorCodeDuplicatedID
				result.ErrorMessage = ErrorMessageDuplicatedID
			} else if errors.As(err, &stErr) {
				result.ErrorCode = ErrorCodeStorageError
				result.ErrorMessage = ErrorMessageStorageError
			} else {
				result.ErrorCode = ErrorCodeDatabaseError
				result.ErrorMessage = ErrorMessageDatabaseError
			}

			results[i] = result
			continue
		}

		results[i] = result
	}

	return results
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
