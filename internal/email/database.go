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
	tableName     = "Outbox"
	statusMeta    = "_META"
	statusInitial = "PENDING"
)

type Database struct {
	dynamo *dynamodb.Client
}

func New(dynamo *dynamodb.Client) *Database {
	return &Database{dynamo: dynamo}
}

func (db *Database) getMetaAttributes(status string, emlFilePath string, createdAt string) map[string]interface{} {
	return map[string]interface{}{
		"Latest":      status,
		"CreatedAt":   createdAt,
		"EMLFilePath": emlFilePath,
	}
}

func (db *Database) Insert(ctx context.Context, id string, emlFilePath string) error {
	metaStmt := fmt.Sprintf("INSERT INTO \"%v\" VALUE {'Id': ?, 'Status': ?, 'Attributes': ?}", tableName)
	metaAttrs := db.getMetaAttributes(statusInitial, emlFilePath, time.Now().Format(time.RFC3339))
	metaParams, _ := attributevalue.MarshalList([]interface{}{id, statusMeta, metaAttrs})

	inStmt := fmt.Sprintf("INSERT INTO \"%v\" VALUE {'Id': ?, 'Status': ?}", tableName)
	inParams, _ := attributevalue.MarshalList([]interface{}{id, statusInitial, map[string]interface{}{}})

	ti := &dynamodb.ExecuteTransactionInput{
		TransactStatements: []types.ParameterizedStatement{
			{Statement: aws.String(metaStmt), Parameters: metaParams},
			{Statement: aws.String(inStmt), Parameters: inParams},
		},
	}

	_, err := db.dynamo.ExecuteTransaction(ctx, ti)
	return err
}

func (db *Database) DeletePending(ctx context.Context, id string) error {
	metaStmt := fmt.Sprintf("DELETE FROM \"%v\" WHERE Id=? AND Status=?", tableName)
	metaParams, _ := attributevalue.MarshalList([]interface{}{id, statusMeta})

	inStmt := fmt.Sprintf("DELETE FROM \"%v\" WHERE Id=? AND Status=?", tableName)
	inParams, _ := attributevalue.MarshalList([]interface{}{id, statusInitial})

	ti := &dynamodb.ExecuteTransactionInput{
		TransactStatements: []types.ParameterizedStatement{
			{Statement: aws.String(metaStmt), Parameters: metaParams},
			{Statement: aws.String(inStmt), Parameters: inParams},
		},
	}

	_, err := db.dynamo.ExecuteTransaction(ctx, ti)
	return err
}
