package github_release

import (
	"context"
	"log"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
)

var client azsecrets.Client

func getSecretData(name string) string {
	vaultURL := os.Getenv("VAULT_URL")

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatalf("failed to get credentials: %v", err)
	}

	client, err := azsecrets.NewClient(vaultURL, cred, nil)

	version := ""
	secret, err := client.GetSecret(context.TODO(), "mySecretName", version, nil)
	if err != nil {
		log.Fatalf("failed to get secret: %v", err)
	}

	// Establish a connection to the Key Vault client
	// azsecrets.NewClient
	// _, err = azsecrets.NewClient(vaultURL, cred, nil)
	// if err != nil {
	// 	log.Fatalf("failed to create a Key Vault client: %v", err)
	// }

	// basicClient := keyvault.New()
	// basicClient.Authorizer = authorizer

	// secret, err := basicClient.GetSecret(context.Background(), vaultURL, name, "")
	// return os.Getenv(string(name))
	return *secret.Value
}
