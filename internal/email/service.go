package email

import (
	"context"
	"errors"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
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

type payloadStorageInterface interface {
	Store(messageId string, payload []byte) (string, error)
	Delete(payloadPath string) error
}

type databaseInterface interface {
	Insert(ctx context.Context, id string, payloadPath string) error
	GetStaleEmails(ctx context.Context) ([]Email, error)
	GetInvalidEmails(ctx context.Context) ([]Email, error)
	RequeueEmail(ctx context.Context, id string) error
	ScanAndSetTTL(ctx context.Context, ttlTimestamp int64, maxRecords int) (*ScanAndSetTTLResult, error)
}

type ScanAndSetTTLResult struct {
	ProcessedRecords int  `json:"processed_records"`
	TotalRecords     int  `json:"total_records"`
	HasMoreRecords   bool `json:"has_more_records"`
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

func (s *Service) isTransientError(err error) bool {
	var provisionedThroughputErr *types.ProvisionedThroughputExceededException
	var requestLimitErr *types.RequestLimitExceeded
	var internalServerErr *types.InternalServerError

	return errors.As(err, &provisionedThroughputErr) ||
		errors.As(err, &requestLimitErr) ||
		errors.As(err, &internalServerErr)
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

		payloadPath, err := s.payloadStorage.Store(req.MessageId, req.PayloadBytes)
		if err != nil {
			log.Printf("failed to create payload file for '%s': %v", req.MessageId, err)
			result.Success = false
			result.ErrorCode = ErrorCodeStorageError
			result.ErrorMessage = ErrorMessageStorageError
			results[i] = result
			continue
		}

		if err := s.db.Insert(ctx, req.MessageId, payloadPath); err != nil {
			log.Printf("failed to insert record in database for '%s': %v", req.MessageId, err)

			s.tryDelete(payloadPath)

			result.Success = false

			var dupErr *types.DuplicateItemException
			if errors.As(err, &dupErr) {
				result.ErrorCode = ErrorCodeDuplicatedID
				result.ErrorMessage = ErrorMessageDuplicatedID
			} else if s.isTransientError(err) {
				result.ErrorCode = ErrorCodeTransientError
				result.ErrorMessage = ErrorMessageTransientError
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

func (s *Service) ScanAndSetTTL(ctx context.Context, ttlTimestamp int64, maxRecords int) (*ScanAndSetTTLResult, error) {
	return s.db.ScanAndSetTTL(ctx, ttlTimestamp, maxRecords)
}
