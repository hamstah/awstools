# kms-env

```
usage: kms-env [<flags>] <command>...

Decrypt environment variables encrypted with KMS, SSM or Secret Manager.

Flags:
      --help                    Show context-sensitive help (also try --help-long and --help-man).
      --kms-prefix="KMS_"       Prefix for the KMS environment variables
      --ssm-prefix="SSM_"       Prefix for the SSM environment variables
      --secrets-manager-prefix="SECRETS_MANAGER_"
                                Prefix for the secrets manager environment variables
      --secrets-manager-version-stage="AWSCURRENT"
                                The version stage of secrets from secrets manager
      --refresh-interval=0      Refresh interval
      --refresh-action=RESTART  Action to take when values have changed
      --refresh-max-retries=5   Number of retries when failing to refresh the config
      --assume-role-arn=ASSUME-ROLE-ARN
                                Role to assume
      --assume-role-external-id=ASSUME-ROLE-EXTERNAL-ID
                                External ID of the role to assume
      --assume-role-session-name=ASSUME-ROLE-SESSION-NAME
                                Role session name
      --region=REGION           AWS Region
      --mfa-serial-number=MFA-SERIAL-NUMBER
                                MFA Serial Number
      --mfa-token-code=MFA-TOKEN-CODE
                                MFA Token Code
      --session-duration=1h     Session Duration
  -v, --version                 Display the version
      --log-level=warn          Log level
      --log-format=text         Log format

Args:
  <command>  Command to run, prefix with -- to pass args
```

## Features

* Scans environment variables with the key prefixes `--kms-prefix`, `--ssm-prefix` or `--secrets-manager-prefix`, fetches and decrypts the values
  then injects them into the environment of the sub command to run.
* Scans environment variable values with the prefix `kms://`, `ssm://`, `secrets-manager://` or `file://`.
* KMS values should be base64 encoded in the value of the variable
* SSM values should be the path to the parameter store parameter. If the path ends in `/*` it will fetch the values
  under that path (non-recursively) and prefix them with the original env var. If the name is prefix with an extra `_`.
  no prefix is used
* Secret manager values should be the name of the secret. It supports JSON encoded values in either SecretString or SecretBinary and will fetch the AWSCURRENT version by default (override with `--secrets-manager-version-stage`). JSON keys are upper cased and prefixed with the name of the environment variable excluding the prefix. To not include the prefix, use an extra `_` before the
variable name, for example `SSM__A=` or `_A=ssm://...`.
* Refresh configuration and restart or exit the child process

## Examples

All sources except `file://` are supported as both key prefixes (`KMS_A`) or value prefixes (`A=kms://...`).

### KMS

```
# key prefix
export KMS_A=<base64>

# value prefix
export A=kms://<base54>

kms-env program
```

`program` will be called with its environment set to the parent process environment with the additional env var `A` with
its value set to the decrypted value of `KMS_A`. `KMS_A` is not passed to the child process.

### SSM Single parameter

```
# key prefix
export SSM_B=/path/to/value

# value prefix
export B=ssm:///path/to/value

kms-env program
```

`program` will be called with its environment set to the parent process environment with the additional env var `B` with
its value set to the value of the parameter under `/path/to/value`. If the parameter was encrypted with KMS, it is automatically
decrypted.

### SSM wildcard

Assuming the following parameters exist
```
/path/to/values/foo-bar
/path/to/values/flip
```

```
# key prefix
export SSM_C=/path/to/values/*

# value prefix
export C=ssm:///path/to/values/*

kms-env program
```

Similar to the previous case, but the environment will have the following variables set
* `C_FOO_BAR`
* `C_FLIP`

### SSM wildcard without prefix

Use a double `_` to ignore the prefix

```
# key prefix
export SSM__C=/path/to/values/*

# value prefix
export _C=ssm:///path/to/values/*

kms-env program
```

Will have
* `FOO_BAR`
* `FLIP`

If multiple variables have a double `_` they all get merged. (Note: there is no guarantee of the order in which they are processed).

### Secrets Manager with prefix

Assuming the secret `name/of/secret` exists
```
{"foo": 123, "bar": "test"}
```

```
# key prefix
export SECRETS_MANAGER_ABC=name/of/secret

# value prefix
ABC=secrets-manager://name/of/secret

kms-env program
```

Will have
* `ABC_FOO=123`
* `ABC_BAR=test`

### Secrets Manager without prefix

Add an extra `_` in the environment variable name:

```
# key prefix
export SECRETS_MANAGER__ABC=name/of/secret

# value prefix
export _ABC=name/of/secret

kms-env program
```

Will have
* `FOO=123`
* `BAR=test`

### Load a file

```
# value prefix
export ABC=file:///path/to/file

kms-env program
```

Will have the content of /path/to/file in the ABC environment variable.

### Refresh the configuration and restart the process if it changed

```
export SECRETS_MANAGER_ABC=name/of/secret

kms-env --refresh-interval 10m --refresh-action RESTART program
```
