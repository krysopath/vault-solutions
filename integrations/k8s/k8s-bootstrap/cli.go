package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"strings"
)

func getUser() *user.User {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	return usr
}

var User = getUser()

var ConfigFilePath = fmt.Sprintf(
	"%s/.secrets.json",
	User.HomeDir)

type Config struct {
	SelfRenew    bool                      `json:"self_renew"`
	EmittedPaths *[]string                 `json:"emitted_paths"`
	Cubby        *[]map[string]interface{} `json:"cubby"`
	Files        map[string]interface{}    `json:"files"`
}

func (c *Config) LoadFromFile(fp string) Config {
	var cfg Config
	if _, err := os.Stat(fp); err == nil {
		content, err := ioutil.ReadFile(fp)
		if err != nil {
			panic(err)
		}
		json.Unmarshal(content, &c)
	}
	return cfg
}

func (c *Config) String() string {
	cfgJson, err := json.Marshal(c)
	if err != nil {
		panic(err)
	}
	return string(cfgJson)
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
		emit = fmt.Sprintf("%s\nexport %s=%s", emit, key, value)
	}
	return strings.TrimSpace(emit)
}

func newIdentity(v *VaultThinClient) {
	for _, cubby := range *v.Config.Cubby {
		v.Write(
			cubby["dest"].(string),
			*v.IAMIdentity(cubby["src"].(string)),
		)
	}
}

func initial(v *VaultThinClient) {
	for _, sourcePath := range v.Config.Files {
		fmt.Fprintf(os.Stdout, *v.EmitFileAction(sourcePath.(string)))
	}
}

func emit(v *VaultThinClient) {
	if v.Config.SelfRenew {
		go renew(v)
	}
	fmt.Fprintln(os.Stdout, EmitEnv(v))
}

func renew(v *VaultThinClient) {
	lease := os.Getenv("VAULT_IAM_LEASE")
	v.Renew(lease)
}

func renderConfig(v *VaultThinClient) {
	fmt.Fprintln(os.Stdout, v.Config)
}

func usage() {
	fmt.Fprintln(os.Stderr, "Usage: k8s-bootstrap initial|identity|env|renew|config")
}

func main() {
	cfg := &Config{}
	cfg.LoadFromFile(ConfigFilePath)
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
