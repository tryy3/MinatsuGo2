package github_release

import (
	"context"
	"log"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
)

func getSecretData(name string) string {
	vaultURL := os.Getenv("VAULT_URL")

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatalf("failed to get credentials: %v", err)
	}

	client, err := azsecrets.NewClient(vaultURL, cred, nil)
	if err != nil {
		log.Fatalf("failed to create a Key Vault client: %v", err)
	}

	version := ""
	secret, err := client.GetSecret(context.TODO(), name, version, nil)
	if err != nil {
		log.Fatalf("failed to get secret: %v", err)
	}

	return *secret.Value
}
