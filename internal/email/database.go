package email

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const (
	StatusMeta                  = "_META"
	StatusAccepted              = "ACCEPTED"
	StatusIntaking              = "INTAKING"
	StatusProcessing            = "PROCESSING"
	StatusCallingSentCallback   = "CALLING-SENT-CALLBACK"
	StatusCallingFailedCallback = "CALLING-FAILED-CALLBACK"
	StatusReady                 = "READY"
	StatusSent                  = "SENT"
	StatusFailed                = "FAILED"
	StatusInvalid               = "INVALID"
	StatusSentAcknowledged      = "SENT-ACKNOWLEDGED"
	StatusFailedAcknowledged    = "FAILED-ACKNOWLEDGED"
)

const (
	statusMeta    = StatusMeta
	statusInitial = StatusAccepted

	statusIntaking              = StatusIntaking
	statusProcessing            = StatusProcessing
	statusCallingSentCallback   = StatusCallingSentCallback
	statusCallingFailedCallback = StatusCallingFailedCallback
	statusReady                 = StatusReady
	statusSent                  = StatusSent
	statusFailed                = StatusFailed
)

type Database struct {
	dynamo                      *dynamodb.Client
	tableName                   string
	staleEmailsThresholdMinutes int
}

func NewDatabase(dynamo *dynamodb.Client, tableName string, staleEmailsThresholdMinutes int) *Database {
	return &Database{
		dynamo:                      dynamo,
		tableName:                   tableName,
		staleEmailsThresholdMinutes: staleEmailsThresholdMinutes,
	}
}

func (db *Database) getMetaAttributes(status string, payloadFilePath string, createdAt string, ttl int64) map[string]interface{} {
	return map[string]interface{}{
		"Latest":          status,
		"CreatedAt":       createdAt,
		"UpdatedAt":       createdAt,
		"PayloadFilePath": payloadFilePath,
	}
}

func (db *Database) Insert(ctx context.Context, id string, payloadFilePath string) error {
	ttl := time.Now().Add(30 * 24 * time.Hour).Unix()

	metaStmt := fmt.Sprintf("INSERT INTO \"%v\" VALUE {'Id': ?, 'Status': ?, 'Attributes': ?, 'TTL': ?}", db.tableName)
	metaAttrs := db.getMetaAttributes(statusInitial, payloadFilePath, time.Now().Format(time.RFC3339), ttl)
	metaParams, _ := attributevalue.MarshalList([]interface{}{id, statusMeta, metaAttrs, ttl})

	inStmt := fmt.Sprintf("INSERT INTO \"%v\" VALUE {'Id': ?, 'Status': ?, 'Attributes': ?, 'TTL': ?}", db.tableName)
	inAttrs := map[string]interface{}{}
	inParams, _ := attributevalue.MarshalList([]interface{}{id, statusInitial, inAttrs, ttl})

	ti := &dynamodb.ExecuteTransactionInput{
		TransactStatements: []types.ParameterizedStatement{
			{Statement: aws.String(metaStmt), Parameters: metaParams},
			{Statement: aws.String(inStmt), Parameters: inParams},
		},
	}

	_, err := db.dynamo.ExecuteTransaction(ctx, ti)
	return err
}

func (db *Database) GetStaleEmails(ctx context.Context) ([]Email, error) {
	thresholdTime := time.Now().Add(-time.Duration(db.staleEmailsThresholdMinutes) * time.Minute)
	thresholdStr := thresholdTime.Format(time.RFC3339)

	// Query for stale emails with Latest in (INTAKING, PROCESSING, CALLING-SENT-CALLBACK, CALLING-FAILED-CALLBACK)
	query := fmt.Sprintf(`SELECT Id, Status, Attributes.Latest, Attributes.CreatedAt, Attributes.UpdatedAt 
		FROM "%v" 
		WHERE Status=? 
		AND (Attributes.Latest=? OR Attributes.Latest=? OR Attributes.Latest=? OR Attributes.Latest=?)
		AND Attributes.UpdatedAt < ?`,
		db.tableName)

	params, err := attributevalue.MarshalList([]interface{}{
		statusMeta,
		statusIntaking,
		statusProcessing,
		statusCallingSentCallback,
		statusCallingFailedCallback,
		thresholdStr,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal parameters: %w", err)
	}

	var allRecords []struct {
		Id         string                 `dynamodbav:"Id"`
		Status     string                 `dynamodbav:"Status"`
		Attributes map[string]interface{} `dynamodbav:"Attributes"`
	}

	// Paginate through all results using NextToken
	var nextToken *string
	for {
		stmt := &dynamodb.ExecuteStatementInput{
			Statement:  aws.String(query),
			Parameters: params,
			NextToken:  nextToken,
		}

		res, err := db.dynamo.ExecuteStatement(ctx, stmt)
		if err != nil {
			return nil, fmt.Errorf("failed to execute statement: %w", err)
		}

		var records []struct {
			Id         string                 `dynamodbav:"Id"`
			Status     string                 `dynamodbav:"Status"`
			Attributes map[string]interface{} `dynamodbav:"Attributes"`
		}

		if err := attributevalue.UnmarshalListOfMaps(res.Items, &records); err != nil {
			return nil, fmt.Errorf("failed to unmarshal records: %w", err)
		}

		allRecords = append(allRecords, records...)

		// Check if there are more results
		if res.NextToken == nil {
			break
		}
		nextToken = res.NextToken
	}

	staleEmails := make([]Email, 0, len(allRecords))
	for _, record := range allRecords {
		latest, _ := record.Attributes["Latest"].(string)
		createdAtStr, _ := record.Attributes["CreatedAt"].(string)
		updatedAtStr, _ := record.Attributes["UpdatedAt"].(string)

		createdAt, _ := time.Parse(time.RFC3339, createdAtStr)
		updatedAt, _ := time.Parse(time.RFC3339, updatedAtStr)

		staleEmails = append(staleEmails, Email{
			Id:        record.Id,
			Status:    latest, // Map Latest to Status as per requirements
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		})
	}

	return staleEmails, nil
}

func (db *Database) GetInvalidEmails(ctx context.Context) ([]Email, error) {
	query := fmt.Sprintf(`SELECT Id, Status, Attributes
		FROM "%v"
		WHERE Status=?
		AND Attributes.Latest=?`,
		db.tableName)

	params, err := attributevalue.MarshalList([]interface{}{
		StatusMeta,
		StatusInvalid,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal parameters: %w", err)
	}

	var allRecords []struct {
		Id         string                 `dynamodbav:"Id"`
		Status     string                 `dynamodbav:"Status"`
		Attributes map[string]interface{} `dynamodbav:"Attributes"`
	}

	// Paginate through all results using NextToken
	var nextToken *string
	for {
		stmt := &dynamodb.ExecuteStatementInput{
			Statement:  aws.String(query),
			Parameters: params,
			NextToken:  nextToken,
		}

		res, err := db.dynamo.ExecuteStatement(ctx, stmt)
		if err != nil {
			return nil, fmt.Errorf("failed to execute statement: %w", err)
		}

		var records []struct {
			Id         string                 `dynamodbav:"Id"`
			Status     string                 `dynamodbav:"Status"`
			Attributes map[string]interface{} `dynamodbav:"Attributes"`
		}

		if err := attributevalue.UnmarshalListOfMaps(res.Items, &records); err != nil {
			return nil, fmt.Errorf("failed to unmarshal records: %w", err)
		}

		allRecords = append(allRecords, records...)

		// Check if there are more results
		if res.NextToken == nil {
			break
		}
		nextToken = res.NextToken
	}

	invalidEmails := make([]Email, 0, len(allRecords))
	for _, record := range allRecords {
		createdAtStr, _ := record.Attributes["CreatedAt"].(string)
		updatedAtStr, _ := record.Attributes["UpdatedAt"].(string)
		errorMessage, _ := record.Attributes["ErrorMessage"].(string)

		createdAt, _ := time.Parse(time.RFC3339, createdAtStr)
		updatedAt, _ := time.Parse(time.RFC3339, updatedAtStr)

		invalidEmails = append(invalidEmails, Email{
			Id:           record.Id,
			Status:       record.Status,
			CreatedAt:    createdAt,
			UpdatedAt:    updatedAt,
			ErrorMessage: errorMessage,
		})
	}

	return invalidEmails, nil
}

func (db *Database) RequeueEmail(ctx context.Context, id string) error {
	// First, get the _META record to check the Latest status
	getStmt := fmt.Sprintf(`SELECT Id, Status, Attributes, TTL FROM "%v" WHERE Id=? AND Status=?`, db.tableName)
	getParams, err := attributevalue.MarshalList([]interface{}{id, statusMeta})
	if err != nil {
		return fmt.Errorf("failed to marshal get parameters: %w", err)
	}

	getRes, err := db.dynamo.ExecuteStatement(ctx, &dynamodb.ExecuteStatementInput{
		Statement:  aws.String(getStmt),
		Parameters: getParams,
	})
	if err != nil {
		return fmt.Errorf("failed to get meta record: %w", err)
	}

	if len(getRes.Items) == 0 {
		return fmt.Errorf("email with id %s not found", id)
	}

	var metaRecord struct {
		Id         string                 `dynamodbav:"Id"`
		Status     string                 `dynamodbav:"Status"`
		Attributes map[string]interface{} `dynamodbav:"Attributes"`
		TTL        int64                  `dynamodbav:"TTL"`
	}

	if err := attributevalue.UnmarshalMap(getRes.Items[0], &metaRecord); err != nil {
		return fmt.Errorf("failed to unmarshal meta record: %w", err)
	}

	currentLatest, ok := metaRecord.Attributes["Latest"].(string)
	if !ok {
		return fmt.Errorf("latest status not found in meta record")
	}

	// Map the current Latest to the new status
	var newStatus string
	switch currentLatest {
	case statusIntaking:
		newStatus = statusInitial // ACCEPTED
	case statusProcessing:
		newStatus = statusReady
	case statusCallingSentCallback:
		newStatus = statusSent
	case statusCallingFailedCallback:
		newStatus = statusFailed
	default:
		return fmt.Errorf("cannot requeue email with Latest status: %s", currentLatest)
	}

	// Delete the record where Status = currentLatest
	deleteStmt := fmt.Sprintf("DELETE FROM \"%v\" WHERE Id=? AND Status=?", db.tableName)
	deleteParams, _ := attributevalue.MarshalList([]interface{}{id, currentLatest})

	// Update the Latest field in the _META record
	updateStmt := fmt.Sprintf("UPDATE \"%v\" SET Attributes.Latest=?, Attributes.UpdatedAt=? WHERE Id=? AND Status=?", db.tableName)
	updateParams, _ := attributevalue.MarshalList([]interface{}{newStatus, time.Now().Format(time.RFC3339), id, statusMeta})

	ti := &dynamodb.ExecuteTransactionInput{
		TransactStatements: []types.ParameterizedStatement{
			{Statement: aws.String(deleteStmt), Parameters: deleteParams},
			{Statement: aws.String(updateStmt), Parameters: updateParams},
		},
	}

	if _, err := db.dynamo.ExecuteTransaction(ctx, ti); err != nil {
		return fmt.Errorf("failed to requeue email: %w", err)
	}

	return nil
}
