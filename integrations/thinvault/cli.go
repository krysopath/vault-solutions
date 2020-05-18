package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

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
	for _, sourcePath := range *v.Config.Files {
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

func emitFile(v *VaultThinClient, name string) {
	path, ok := (*v.Config.Files)[name]
	if ok {
		fmt.Fprintln(
			os.Stdout,
			strings.TrimSpace(*v.EmitFileAction(path.(string))),
		)
	}
}

func usage() {
	fmt.Fprintln(
		os.Stderr,
		`Usage: thinvault identity|env|renew|file|config

    This helper utility can create and maintain an IAM identity for you. It
    also creates a valid local kubernetes authentication.

    Our dependencies are:
    1. kubectl (create k8s api requests)
    2. aws-iam-authenticator (authenticate to k8s via IAM)
    3. vault login (to create IAM & store secrets in the cubbyhole)

    identity: creates an IAM user that expires in one week (if not renewed) and
    saves it into the cubbyhole secret engine.

    env: This command emits a list of 'export K=V' statements that can be
    evaluated by any shell. They contain secrets for the aws-iam-authenticator.
         example: "$(thinvault env)"

    renew: this commands attempts to renew the parent vault token and the lease
    on the IAM user that is saved in its cubbyhole.		

    file: takes another argument that is a key to defined file
         example: "thinvault file KUBECONFIG > ./kubeconf"

    config: render the config as evaluated by the process as yaml.
		`)
}

func tokencheck() {
	if len(os.Getenv("VAULT_TOKEN")) == 0 {
		fmt.Fprintf(os.Stderr, "warn: no VAULT_TOKEN\n")
	}

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
	tokencheck()
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
		case "file":
			emitFile(v, os.Args[2])
		default:
			usage()
			os.Exit(1)
		}
	} else {
		usage()
	}
}
