# direnv

`direnv` is a tool which loads environment variables based on your present
working directory or `$PWD`. As you descend into or ascend back through the
hierarchy of resources on your filesystem.

- this allows for simple auto provisioning of namespaced values
- this also supports arbitrary code execution (like anything on shell)
- an integration with vault could access secrets for you

# problem

lets say we have a certain infrastructure layout with respective files. e.g.:
```
iac
├── dev
│   ├── .envrc
│   ├── project-a
│   │   └── .envrc
│   └── project-b
│       └── .envrc
├── .envrc
└── ops
    ├── accounts
    │   └── .envrc
    ├── deploy
    │   └── .envrc
    └── .envrc
```

> Note the .envrc in all subdirectories

By setting specific environment variables automatically on descending into
these specific directories, we can stop caring about micromanagment of shell
variables according to context.

## caveats
- error messages when you dont reach vault
- added execution time to `cd /directory/with/vaultified/direnv`

## benefits
- bring namespaced and automatically provisioned values to your development
- never forget to load your secrets anymore :)
- define arbitrary code to run when you descend into its parent directory (:


## setup instructions:

0. install direnv (available via apt and yum)
1. choose a `directory/`
2. create a file  `directory/.envrc` and open it with an editor
3. export variables like in a shell script: `export MY_KEY=some-value`, save.
4. run `direnv allow directory/`
5. `cd directory/` and see the magic at work

## setup for vault integration

To provide yourself with secrets while working in a simple manner, you can
connect vault with direnv. Since you talk to vault via simple http, you do not
need any special software.

> I used `vault` cli binary here for convenience reasons.

Append to your `.envrc`:
```
export VAULT_ADDR=https://vault.internal.domain.tld:443

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
```

save and `direnv allow` it.

Next make sure to export a valid `VAULT_TOKEN` and export
`VAULT_PATH=secret/services/shared/env` or similar.

Whenever you `cd` into the directory the code will run.

## execution example

```
Last login: Mon Mar 30 17:56:38 2020 from 192.168.121.1
direnv: loading .envrc
direnv: export -PS2

$  export VAULT_TOKEN=s.******************
$ export VAULT_PATHS='secret/services/shared/env secret/services/shared/anchore'
$ echo $VAULT_PATHS
secret/services/shared/env secret/services/shared/anchore
$ cd ..
direnv: unloading
$ cd $HOME
direnv: loading .envrc
direnv: export +ANCHORE_CLI_PASS +ANCHORE_CLI_URL +ANCHORE_CLI_USER
+AWS_ACCESS_KEY_ID +AWS_DEFAULT_REGION +AWS_ECR_REPOSITORY
+AWS_SECRET_ACCESS_KEY +CROWDIN_PROJECT_ID +CROWDIN_TOKEN +NPM_TOKEN
+SONAR_TOKEN +TAXJAR_TOKEN -PS2
$ env | grep ANCHORE
ANCHORE_CLI_USER=shared
ANCHORE_CLI_PASS=O******************3
ANCHORE_CLI_URL=https://anchore.ops.domain.tld/v1/
$ cd my-special-dir/
direnv: loading .envrc
direnv: export +PORTAINER_ADDR +PORTAINER_PASSWORD +PORTAINER_USER -PS2
$ cd ..
direnv: loading .envrc
direnv: export +ANCHORE_CLI_PASS +ANCHORE_CLI_URL +ANCHORE_CLI_USER
+AWS_ACCESS_KEY_ID +AWS_DEFAULT_REGION +AWS_ECR_REPOSITORY
+AWS_SECRET_ACCESS_KEY +CROWDIN_PROJECT_ID +CROWDIN_TOKEN +NPM_TOKEN
+SONAR_TOKEN +TAXJAR_TOKEN -PS2
```

