package main

import (
	"log"

	"github.com/hashicorp/vault/api"
)

// NewClient creates a new VaultThinClient
func NewClient(thinConfig *Config) *VaultThinClient {
	vaultConfig := api.DefaultConfig()
	client, err := api.NewClient(vaultConfig)
	if err != nil {
		log.Fatalf("err: %s", err)
	}
	return &VaultThinClient{
		Vault:  client,
		Config: thinConfig,
	}
}

// VaultThinClient sits atop Vault and simplifies common workflows.
type VaultThinClient struct {
	Vault  *api.Client
	Config *Config
}
