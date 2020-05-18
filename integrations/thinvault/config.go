package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

func getUser() *user.User {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	return usr
}

// User holds information we need even in simple scenarios.
var User = getUser()

// ConfigFilePath holds a path to search for confguration
var ConfigFilePath = filepath.Join(
	User.HomeDir, ".thinvault.yaml",
)

// DefaultConfigYml contains a yaml representation of the default configuration
const DefaultConfigYml string = `---
renew: true
emit:
- secret/services/shared/env
- cubbyhole/iam/aws-developer
fstring: |-
  %s
  export %s=%s
cubby:
- dest: cubbyhole/iam/aws-developer
  src: dev/aws/creds/aws-developer
files:
  KUBECONFIG: secret/services/k8s/tf-stage/kubeconfig_developer
`

// Config describes the settings of what the VaultThinClient will do
type Config struct {
	SelfRenew      bool                      `json:"renew" yaml:"renew"`
	EmittedPaths   *[]string                 `json:"emit" yaml:"emit"`
	EmitterFString string                    `json:"fstring" yaml:"fstring"`
	Cubby          *[]map[string]interface{} `json:"cubby" yaml:"cubby"`
	Files          *map[string]interface{}   `json:"files" yaml:"files"`
}

// LoadFromFile loads a config from file and returns it as struct as well.
func (c *Config) LoadFromFile(fp string) Config {
	if _, err := os.Stat(fp); err == nil {
		content, err := ioutil.ReadFile(fp)
		if err != nil {
			panic(err)
		}
		extension := filepath.Ext(fp)
		switch extension {
		case ".yaml":
			yaml.Unmarshal(content, &c)
		case ".json":
			json.Unmarshal(content, &c)
		}
	} else {
		err := yaml.Unmarshal([]byte(DefaultConfigYml), &c)
		if err != nil {
			panic(err)
		}
	}
	return *c
}

// String creates a yaml represention of the Config struct
func (c *Config) String() string {
	cfgBytes, err := yaml.Marshal(c)
	//cfgJson, err := json.Marshal(c)
	if err != nil {
		panic(err)
	}
	return string(cfgBytes)
}
