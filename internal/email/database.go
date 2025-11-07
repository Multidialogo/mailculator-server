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
	statusMeta    = "_META"
	statusInitial = "ACCEPTED"
)

type Database struct {
	dynamo    *dynamodb.Client
	tableName string
}

func NewDatabase(dynamo *dynamodb.Client, tableName string) *Database {
	return &Database{
		dynamo:    dynamo,
		tableName: tableName,
	}
}

func (db *Database) getMetaAttributes(status string, emlFilePath string, payloadFilePath string, createdAt string, ttl int64) map[string]interface{} {
	return map[string]interface{}{
		"Latest":          status,
		"CreatedAt":       createdAt,
		"EMLFilePath":     emlFilePath,
		"PayloadFilePath": payloadFilePath,
		"TTL":             ttl,
	}
}

func (db *Database) Insert(ctx context.Context, id string, emlFilePath string, payloadFilePath string) error {
	ttl := time.Now().Add(30 * 24 * time.Hour).Unix()

	metaStmt := fmt.Sprintf("INSERT INTO \"%v\" VALUE {'Id': ?, 'Status': ?, 'Attributes': ?}", db.tableName)
	metaAttrs := db.getMetaAttributes(statusInitial, emlFilePath, payloadFilePath, time.Now().Format(time.RFC3339), ttl)
	metaParams, _ := attributevalue.MarshalList([]interface{}{id, statusMeta, metaAttrs})

	inStmt := fmt.Sprintf("INSERT INTO \"%v\" VALUE {'Id': ?, 'Status': ?, 'Attributes': ?}", db.tableName)
	inAttrs := map[string]interface{}{"TTL": ttl}
	inParams, _ := attributevalue.MarshalList([]interface{}{id, statusInitial, inAttrs})

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
	metaStmt := fmt.Sprintf("DELETE FROM \"%v\" WHERE Id=? AND Status=?", db.tableName)
	metaParams, _ := attributevalue.MarshalList([]interface{}{id, statusMeta})

	inStmt := fmt.Sprintf("DELETE FROM \"%v\" WHERE Id=? AND Status=?", db.tableName)
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
