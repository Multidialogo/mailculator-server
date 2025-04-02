package testutils

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"log"
	"multicarrier-email-api/internal/eml"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const (
	testEmailTableStatusIndex = "StatusIndex"
	testEmailTableStatusMeta  = "_META"
	testEmailTableName        = "Outbox"
)

func NewAwsTestConfigFromEnv() aws.Config {
	return aws.Config{
		Region: os.Getenv("AWS_REGION"),
		Credentials: credentials.NewStaticCredentialsProvider(
			os.Getenv("AWS_ACCESS_KEY_ID"),
			os.Getenv("AWS_SECRET_ACCESS_KEY"),
			"",
		),
		BaseEndpoint: aws.String(os.Getenv("AWS_BASE_ENDPOINT")),
	}
}

type EmailTestFixtureKeys struct {
	Id     string
	Status string
}

type EmailTestFixture struct {
	Id          string
	Status      string
	Latest      string
	EMLFilePath string
}

type dynamodbRecord struct {
	Id         string                 `dynamodbav:"Id"`
	Status     string                 `dynamodbav:"Status"`
	Attributes map[string]interface{} `dynamodbav:"Attributes"`
}

type EmailDatabaseFacade struct {
	db *dynamodb.Client
}

func NewEmailDatabaseFacade() *EmailDatabaseFacade {
	cfg := NewAwsTestConfigFromEnv()
	return &EmailDatabaseFacade{db: dynamodb.NewFromConfig(cfg)}
}

func (edf *EmailDatabaseFacade) unmarshalList(src []map[string]types.AttributeValue) ([]EmailTestFixture, error) {
	var items []dynamodbRecord
	err := attributevalue.UnmarshalListOfMaps(src, &items)
	if err != nil {
		return []EmailTestFixture{}, err
	}

	emails := make([]EmailTestFixture, len(items))

	for i, item := range items {
		emails[i] = EmailTestFixture{
			Id:          item.Id,
			Status:      item.Status,
			Latest:      fmt.Sprint(item.Attributes["Latest"]),
			EMLFilePath: fmt.Sprint(item.Attributes["EMLFilePath"]),
		}
	}

	return emails, nil
}

func (edf *EmailDatabaseFacade) Query(ctx context.Context, status string, limit int) ([]EmailTestFixture, error) {
	query := fmt.Sprintf("SELECT Id, Status, Attributes FROM \"%v\".\"%v\" WHERE Status=? AND Attributes.Latest =?", testEmailTableName, testEmailTableStatusIndex)
	params, err := attributevalue.MarshalList([]interface{}{testEmailTableStatusMeta, status})
	if err != nil {
		return []EmailTestFixture{}, err
	}

	stmt := &dynamodb.ExecuteStatementInput{
		Parameters: params,
		Statement:  aws.String(query),
		Limit:      aws.Int32(int32(limit)),
	}

	res, err := edf.db.ExecuteStatement(ctx, stmt)
	if err != nil {
		return []EmailTestFixture{}, err
	}

	return edf.unmarshalList(res.Items)
}

func (edf *EmailDatabaseFacade) RemoveFixtures(fixtures []EmailTestFixtureKeys) error {
	if len(fixtures) == 0 {
		log.Print("no fixtures to delete")
		return nil
	}

	log.Printf("deleting fixtures: %v", fixtures)

	var errorList []error

	query := fmt.Sprintf("DELETE FROM \"%v\" WHERE Id=? AND Status=?", testEmailTableName)
	for _, fixture := range fixtures {
		params, _ := attributevalue.MarshalList([]interface{}{fixture.Id, fixture.Status})
		stmt := &dynamodb.ExecuteStatementInput{Statement: aws.String(query), Parameters: params}

		if _, err := edf.db.ExecuteStatement(context.TODO(), stmt); err != nil {
			wrappedErr := fmt.Errorf("error while deleting fixture %s, error: %w", fixture.Id, err)
			log.Printf("%v", wrappedErr)
			errorList = append(errorList, wrappedErr)
		}
	}

	if len(errorList) == 0 {
		return nil
	}

	var err error
	for _, foundError := range errorList {
		err = fmt.Errorf("%w\n%w", err, foundError)
	}

	return err
}

func DummyEMLDataBatch(count int) []eml.EML {
	slice := make([]eml.EML, count)

	for i := 0; i < count; i++ {
		slice[i] = eml.EML{
			MessageId: uuid.NewString(),
			From:      "sender@test.multidialogo.it",
			ReplyTo:   "no-reply@test.multidialogo.it",
			To:        "recipient@test.multidialogo.it",
			Subject:   "Test Email with Reply-To",
			BodyHTML:  "<p>HTML format.</p>",
			BodyText:  "Plain text format.",
			Date:      time.Now(),
		}
	}

	return slice
}
