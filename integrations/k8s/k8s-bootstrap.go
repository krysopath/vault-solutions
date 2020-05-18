package main

import (
	"fmt"
	"os"
	"strings"
)

var (
	identityPath   string = os.Getenv("VAULT_IAM")        //`dev/aws/creds/aws-developer`
	kubeConfigPath string = os.Getenv("VAULT_KUBECONFIG") //`secret/services/k8s/tf-stage/kubeconfig_developer`
	selfRenew      bool   = true
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
		emit = fmt.Sprintf("%s\nexport %s=%s", emit, key, value)
	}
	return strings.TrimSpace(emit)
}
