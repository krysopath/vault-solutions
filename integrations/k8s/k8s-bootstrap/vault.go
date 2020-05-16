package main

import (
	"encoding/base64"
	"log"

	vault "github.com/hashicorp/vault/api"
)

// NewClient creates a new VaultThinClient
func NewClient(thinConfig *Config) *VaultThinClient {
	vault_config := vault.DefaultConfig()
	client, err := vault.NewClient(vault_config)
	if err != nil {
		log.Fatalf("err: %s", err)
	}
	return &VaultThinClient{
		Vault:  client,
		Config: thinConfig,
	}
}

//type Vault interface {
//	Write(string, map[string]interface{}) (*api.Secret, error)
//	Read(string) (*vault.Secret, error)
//}

// VaultThinClient sits atop Vault and simplifies common workflows.
type VaultThinClient struct {
	Vault  *vault.Client
	Config *Config
}

// GetCubby gets the Data of vault.Secret from tokens cubbyhole/
func (c *VaultThinClient) GetCubby(cubbypath string) map[string]interface{} {
	leased, err := c.Read(cubbypath)
	if err != nil {
		log.Fatalf("err: %s", err)
	}
	return leased.Data
}

// Renew can renew any lease as long vault can renew the lease
func (c *VaultThinClient) Renew(leaseID string) (*vault.Secret, error) {
	return c.Vault.Sys().Renew(leaseID, 604800)
}

// Read can read a vault.Secret if the token can read it
func (c *VaultThinClient) Read(p string) (*vault.Secret, error) {
	return c.Vault.Logical().Read(p)
}

// Write can write a vault.Secret
func (c *VaultThinClient) Write(
	p string,
	data map[string]interface{}) (*vault.Secret, error) {

	return c.Vault.Logical().Write(p, data)
}

// Identity creates an IAM persona via vault middleman
func (c *VaultThinClient) IAMIdentity(path string) *map[string]interface{} {
	identity, err := c.Read(path)
	if err != nil {
		log.Fatalf("err: %s", err)
	}
	return &map[string]interface{}{
		"AWS_ACCESS_KEY_ID":     identity.Data["access_key"],
		"AWS_SECRET_ACCESS_KEY": identity.Data["secret_key"],
		"VAULT_IAM_LEASE":       identity.LeaseID,
	}
}

// CreateKubeConfig creates a valid KUBECONFIG file
func (c *VaultThinClient) EmitFileAction(src string) *string {
	secret, err := c.Read(src)
	fileB64 := secret.Data["data"].(string)
	newFile, err := base64.StdEncoding.DecodeString(fileB64)
	if err != nil {
		log.Fatal("decode error:", err)
	}
	fileContent := string(newFile)
	return &fileContent
}
