package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

func getUser() *user.User {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	return usr
}

// User holds information we need in even in simple scenarios.
var User = getUser()

// ConfigFilePath holds a path to search for confguration instructions
var ConfigFilePath = filepath.Join(
	User.HomeDir, ".secrets.yaml",
)

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
	Files          map[string]interface{}    `json:"files" yaml:"files"`
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

// EmitEnv emits a string that can be evaluated by any shell
func EmitEnv(v *VaultThinClient) string {
	Emitted := make(map[string]interface{})

	for _, p := range *v.Config.EmittedPaths {
		for key, value := range v.GetCubby(p) {
			Emitted[key] = value
		}
	}

	var emit string = ""
	for key, value := range Emitted {
		emit = fmt.Sprintf(
			v.Config.EmitterFString,
			emit,
			key,
			value)
	}
	return strings.TrimSpace(emit)
}

func newIdentity(v *VaultThinClient) {
	for _, cubby := range *v.Config.Cubby {
		v.Write(
			cubby["dest"].(string),
			v.IAMIdentity(cubby["src"].(string)),
		)
	}
}

func initial(v *VaultThinClient) {
	for _, sourcePath := range v.Config.Files {
		fmt.Fprintln(
			os.Stdout,
			strings.TrimSpace(
				*v.EmitFileAction(sourcePath.(string)),
			),
		)
	}
}

func emit(v *VaultThinClient) {
	if v.Config.SelfRenew {
		go renew(v)
	}
	fmt.Fprintln(
		os.Stdout, EmitEnv(v))
}

func renew(v *VaultThinClient) {
	lease := os.Getenv("VAULT_IAM_LEASE")
	v.Renew(lease)
}

func renderConfig(v *VaultThinClient) {
	fmt.Fprintln(os.Stdout, v.Config)
}

func usage() {
	fmt.Fprintln(
		os.Stderr,
		`Usage: k8s-bootstrap initial|identity|env|renew|config

    This helper utility can create and maintain an IAM identity for you. It
    also creates a valid local kubernetes authentication.

    Our dependencies are:
    1. kubectl (create k8s api requests)
    2. aws-iam-authenticator (authenticate to k8s via IAM)
    3. vault login (to create IAM & store secrets in the cubbyhole)

    initial: This subcommands fetches the proper KUBECONFIG file.

    identity: creates an IAM user that expires in one week (if not renewed) and
    saves it into the cubbyhole secret engine.

    env: This command emits a list of 'export K=V' statements that can be
    evaluated by any shell. They contain secrets for the aws-iam-authenticator.

    renew: this commands attempts to renew the parent vault token and the lease
    on the IAM user that is saved in its cubbyhole.		

    config: render the config as evaluated by the process as yaml.
		`)
}

func main() {
	cfgPtr := flag.String("cfg",
		ConfigFilePath,
		"a path to the configuration file (if any; can be json and yaml)",
	)
	flag.Parse()
	cfg := &Config{}
	cfg.LoadFromFile(*cfgPtr)
	v := NewClient(cfg)

	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "initial":
			initial(v)
		case "identity":
			newIdentity(v)
		case "env":
			emit(v)
		case "renew":
			renew(v)
		case "config":
			renderConfig(v)
		default:
			usage()
			os.Exit(1)
		}
	} else {
		usage()
	}
}
