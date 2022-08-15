package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface"
	"golang.org/x/oauth2/google"
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

type GoogleAppCredsSecret struct {
	Type                    string `json:"type"`
	ProjectId               string `json:"project_id"`
	PrivateKeyId            string `json:"private_key_id"`
	PrivateKey              string `json:"private_key"`
	ClientEmail             string `json:"client_email"`
	ClientId                string `json:"client_id"`
	AuthUri                 string `json:"auth_uri"`
	TokenUri                string `json:"token_uri"`
	AuthProviderX509CertUrl string `json:"auth_provider_x509_cert_url"`
	ClientX509CertUrl       string `json:"client_x509_cert_url"`
}

const (
	secretNameGoogleAppCreds = "ocdtracker-google-app-creds"
	secretNamePostgresCreds  = "ocdtracker-postgres-creds"
	defaultRegion            = "eu-west-1"
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

func (sm *SecretsManager) GetGoogleAppCreds(ctx context.Context) (*google.Credentials, error) {
	secretString, err := sm.getSecret(secretNameGoogleAppCreds)
	if err != nil || secretString == nil {
		return nil, fmt.Errorf("failed to get %s secret: %w", secretNameGoogleAppCreds, err)
	}
	scopes := []string{"https://www.googleapis.com/auth/cloud-platform", "https://www.googleapis.com/auth/firebase"}
	googleAppCreds, err := google.CredentialsFromJSON(ctx, []byte(*secretString), scopes...)
	if err != nil {
		return nil, fmt.Errorf("failed to build google application credentials from json")
	}
	return googleAppCreds, nil
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
