#!/bin/bash

set -euo pipefail

VAULT_KUBECONFIG="${VAULT_KUBECONFIG:-secret/path/to/tf-stage/kube_config}"
VAULT_IAM="${VAULT_IAM:-dev/aws/creds/aws-developer}"
VAULT_RENEW="${VAULT_RENEW:-true}"

CLUSTER="$(sed -E 's,^(\w+\/)+(tf-(ops|prod|stage))\/(\w+)$,\2,' <<<"$VAULT_KUBECONFIG")"
ROLE="$(sed -E 's,^(\w+\/)+(tf-(ops|prod|stage))\/(\w+)$,\4,' <<<"$VAULT_KUBECONFIG")"

test_binaries() {
	test -x "$(command -v kubectl)" || exit 51
	test -x "$(command -v aws-iam-authenticator)" || exit 52
	test -x "$(command -v vault)" || exit 53
	test -x "$(command -v jq)" || exit 54
}

fetch_kubeconfig() {
    echo "+ Creating KUBECONFIG  in '$1'" > /dev/stderr
	vault read -format=json "$VAULT_KUBECONFIG" |
		jq -r .data.data |
		base64 -d |
        tee /dev/stderr >"$1"
}

initial_config() {
	mkdir -p "$HOME/.kube/$CLUSTER/"
	fetch_kubeconfig "$HOME/.kube/$CLUSTER/$ROLE"
}

identity() {
	DATA=$(vault read -format=json "$VAULT_IAM")
    echo "+ Created IAM lease '$(jq -r .lease_id<<<"$DATA")'" > /dev/stderr
	vault write "cubbyhole/iam" \
		AWS_ACCESS_KEY_ID="$(jq -r '.data.access_key' <<<"$DATA")" \
		AWS_SECRET_ACCESS_KEY="$(jq -r '.data.secret_key' <<<"$DATA")" \
		VAULT_IAM_LEASE="$(jq -r .lease_id <<<"$DATA")" &>/dev/null
	echo "+ Saved secrets in 'cubbyhole/iam'" >/dev/stderr
}

emit_env_statements() {
	vault read "cubbyhole/iam" |
		sed -E "1,2d;s,^([A-Z_]+)[[:space:]]+(.*)$,export \1=\2,"
	echo "export KUBECONFIG=$HOME/.kube/$CLUSTER/$ROLE"
}

lease_renew() {
	if [ -n "${VAULT_IAM_LEASE:-}" ]; then
		vault lease renew "$VAULT_IAM_LEASE" >/dev/null \
            && echo "+ Renewed the IAM lease."
	    VAULT_TOKEN="${VAULT_TOKEN:-$(cat ~/.vault-token)}"
	    vault token renew "$VAULT_TOKEN" &>/dev/null ||
	    	echo "! Can not renew this vault token." >/dev/stderr

	else
		echo "! Error: run 'eval \$($(basename "$0") env)' first"
	fi

}

usage() {
	echo -e "Usage:\n\t$(basename "$0") initial|identity|env|renew" >/dev/stderr
	echo "
    This helper utility can create and maintain an IAM identity for you. It
    also creates a valid local kubernetes authentication.

    Our dependencies are:
    1. kubectl (create k8s api requests)
    2. aws-iam-authenticator (authenticate to k8s via IAM)
    3. vault (to create IAM & store secrets in the cubbyhole)
    4. jq (to parse json in a simple way)

    We fail if we do not find out dependencies with exits 50+N, where N is the
    dependency in the list above.

    initial: This subcommands checks if all dependencies are in place. It also
    fetches the proper KUBECONFIG file.

    identity: creates an IAM user that expires in one week (if not renewed) and
    saves it into the cubbyhole secret engine.

    env: This command emits a list of \`export K=V\` statements that can be
    evaluated by any shell. They contain secrets for the aws-iam-authenticator.

    renew: this commands attempts to renew the parent vault token and the lease
    on the IAM user that is saved in its cubbyhole.

	"
}

case ${1:-?} in
initial)
	test_binaries
	initial_config
	;;
env)
	test_binaries
	emit_env_statements
	$VAULT_RENEW && lease_renew &>/dev/null &
	;;
renew)
	test_binaries
	lease_renew
	;;
identity)
	test_binaries
	identity
	;;
*)
	usage
	;;
esac
