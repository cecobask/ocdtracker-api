package aws

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface"
)

type SecretsManager struct {
	client secretsmanageriface.SecretsManagerAPI
}

type PostgresCredsSecret struct {
	Username             string `json:"username"`
	Password             string `json:"password"`
	Engine               string `json:"engine"`
	Host                 string `json:"host"`
	Port                 int    `json:"port"`
	DBName               string `json:"dbname"`
	DBInstanceIdentifier string `json:"dbInstanceIdentifier"`
}

const (
	secretNamePostgresCreds = "postgres-creds"
	defaultRegion           = "eu-west-1"
)

func NewSecretsManager(sess *session.Session) *SecretsManager {
	return &SecretsManager{client: secretsmanager.New(sess, &aws.Config{Region: aws.String(defaultRegion)})}
}

func (sm *SecretsManager) getSecret(secretName string) (*string, error) {
	input := &secretsmanager.GetSecretValueInput{SecretId: aws.String(secretName)}
	secretOutput, err := sm.client.GetSecretValue(input)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret value: %w", err)
	}
	return secretOutput.SecretString, nil
}

func (sm *SecretsManager) GetPostgresCreds() (*PostgresCredsSecret, error) {
	secretString, err := sm.getSecret(secretNamePostgresCreds)
	if err != nil || secretString == nil {
		return nil, fmt.Errorf("failed to get %s secret: %w", secretNamePostgresCreds, err)
	}
	var postgresCredsSecret PostgresCredsSecret
	err = json.Unmarshal([]byte(*secretString), &postgresCredsSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal %s secret: %w", secretNamePostgresCreds, err)
	}
	return &postgresCredsSecret, nil
}
