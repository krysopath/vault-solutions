#!/bin/bash

set -euo pipefail

VAULT_KUBECONFIG=secret/path/to/kube_config
VAULT_IAM=dev/aws/creds/aws-developer
CLUSTER="$(sed -E 's,^(\w+\/)+(tf-(ops|prod|stage))\/(\w+)$,\2,' <<< $VAULT_KUBECONFIG)"
ROLE="$(sed -E 's,^(\w+\/)+(tf-(ops|prod|stage))\/(\w+)$,\4,' <<< $VAULT_KUBECONFIG)"

test_binaries () {
    test -x "$(command -v kubectl)" || exit 50
    test -x "$(command -v aws-iam-authenticator)" || exit 51
    test -x "$(command -v vault)" || exit 52
    test -x "$(command -v jq)" || exit 53
    test -x "$(command -v k9s)" || exit 54
}


fetch_kubeconfig() {
    vault read -format=json "$VAULT_KUBECONFIG"\
	| jq -r .data.data \
    | base64 -d > "$1"

}

initial_config() {
	mkdir -p "$HOME/.kube/$CLUSTER/"
	fetch_kubeconfig "$HOME/.kube/$CLUSTER/$ROLE"
}

create_cubbyhole() {
	DATA=$(vault read -format=json $VAULT_IAM); 
	vault write "cubbyhole/iam" \
		AWS_ACCESS_KEY_ID="$(jq -r '.data.access_key' <<<"$DATA")"\
		AWS_SECRET_ACCESS_KEY="$(jq -r '.data.secret_key' <<<"$DATA")"\
		VAULT_IDENTITY_LEASE="$(jq -r .lease_id <<<"$DATA")"
    echo "+ Created cubbyhole/iam" >/dev/stderr
}

read_cubbyhole() {
	vault read "cubbyhole/iam" \
    | sed -E "1,2d;s,^([A-Z_]+)[[:space:]]+(.*)$,export \1=\2,"
    echo "export KUBECONFIG=$HOME/.kube/$CLUSTER/$ROLE"

}

lease_renew() {
    VAULT_TOKEN="${VAULT_TOKEN:-$(cat ~/.vault-token)}"
    vault token renew "$VAULT_TOKEN" \
        || echo "! Can not renew this vault token." > /dev/stderr

	if [ -n "$VAULT_IDENTITY_LEASE" ]; then
		vault lease renew "$VAULT_IDENTITY_LEASE" >/dev/null\
            && echo "+ Renewed the IAM lease."
	else
        echo "! Error: run 'eval \$($(basename "$0") env)' first"
	fi
}

usage() {
    echo "$(basename "$0") intitial|cubby|env|renew" > /dev/stderr
	echo "

	"
}

case $1 in
	initial) 
		test_binaries
        initial_config
		;;
	env) 
		read_cubbyhole
		;;
	renew)
		lease_renew
		;;
	cubby)
		create_cubbyhole
		;;
	*) 
		usage 
		;;
esac

