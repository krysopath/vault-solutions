# direnv

direnv is a tool which loads environment variables based based on your present
working directory or `$PWD`. As you descend into or ascend back through the
hierarchy of resources on your filesystem.

- this allows for simple auto provisioning of important values
- this also supports arbitrary code execution (like anything on shell)
- an integration with vault could access secrets for you

## example

```
Last login: Mon Mar 30 17:56:38 2020 from 192.168.121.1
direnv: loading .envrc
direnv: export -PS2

$  export VAULT_TOKEN=s.******************
$ cat .envrc
#!/usr/bin/env bash

# setting some import shell variables here
export VAULT_ADDR=https://vault.internal.3yourmind.com:443

# lastly enrich the session with vault data
function vaultified() {
    for leafnode in "$@"; do
        for values in $(vault kv get $leafnode\
        |sed -E "1,3d;s,^([A-Z_]+)[[:space:]]+(.*)$,\1=\2,"); do
            K=$(awk -F= '{ print $1 }' <<<$values)
            V=$(awk -F= '{ print $2 }' <<<$values)
            export $K=$V
        done;
     done;
}
if [ -n $VAULT_TOKEN ]; then
    vaultified ${VAULT_PATHS}
fi
$ export VAULT_PATHS='secret/services/shared/env secret/services/shared/anchore'
$ echo $VAULT_PATHS
secret/services/shared/env secret/services/shared/anchore
$ cd ..
direnv: unloading
$ cd $HOME
direnv: loading .envrc
direnv: export +ANCHORE_CLI_PASS +ANCHORE_CLI_URL +ANCHORE_CLI_USER +AWS_ACCESS_KEY_ID +AWS_DEFAULT_REGION +AWS_ECR_REPOSITORY +AWS_SECRET_ACCESS_KEY +CROWDIN_PROJECT_ID +CROWDIN_TOKEN +NPM_TOKEN +SONAR_TOKEN +TAXJAR_TOKEN -PS2
$ env | grep ANCHORE
ANCHORE_CLI_USER=shared
ANCHORE_CLI_PASS=O******************3
ANCHORE_CLI_URL=https://anchore.ops.3yourmind.com/v1/
$ cd my-special-dir/
direnv: loading .envrc
direnv: export +PORTAINER_ADDR +PORTAINER_PASSWORD +PORTAINER_USER -PS2
$ cd ..
direnv: loading .envrc
direnv: export +ANCHORE_CLI_PASS +ANCHORE_CLI_URL +ANCHORE_CLI_USER +AWS_ACCESS_KEY_ID +AWS_DEFAULT_REGION +AWS_ECR_REPOSITORY +AWS_SECRET_ACCESS_KEY +CROWDIN_PROJECT_ID +CROWDIN_TOKEN +NPM_TOKEN +SONAR_TOKEN +TAXJAR_TOKEN -PS2
```
