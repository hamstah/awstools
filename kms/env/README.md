# kms-env

```
usage: kms-env [<flags>] <command>...

Decrypt environment variables encrypted with KMS, SSM or Secret Manager.

Flags:
      --help               Show context-sensitive help (also try --help-long and --help-man).
      --assume-role-arn=ASSUME-ROLE-ARN
                           Role to assume
      --assume-role-external-id=ASSUME-ROLE-EXTERNAL-ID
                           External ID of the role to assume
      --assume-role-session-name=ASSUME-ROLE-SESSION-NAME
                           Role session name
      --region=REGION      AWS Region
      --mfa-serial-number=MFA-SERIAL-NUMBER
                           MFA Serial Number
      --mfa-token-code=MFA-TOKEN-CODE
                           MFA Token Code
  -v, --version            Display the version
      --kms-prefix="KMS_"  Prefix for the KMS environment variables
      --ssm-prefix="SSM_"  Prefix for the SSM environment variables
      --secrets-manager-prefix="SECRETS_MANAGER_"
                           Prefix for the secrets manager environment variables
      --secrets-manager-version-stage="AWSCURRENT"
                           The version stage of secrets from secrets manager

Args:
  <command>  Command to run, prefix with -- to pass args
```

## Features

* Scans environment variables with the prefix `--kms-prefix`, `--ssm-prefix` or `--secrets-manager-prefix`, fetches and decrypts the values
  then injects them into the environment of the sub command to run.
* KMS values should be base64 encoded in the value of the variable
* SSM values should be the path to the parameter store parameter. If the path ends in `/*` it will fetch the values
  under that path (non-recursively) and prefix them with the original env var. If the name is prefix with an extra `_`.
  no prefix is used
* Secret manager values should be the name of the secret. It supports JSON encoded values in either SecretString or SecretBinary and will fetch
  the AWSCURRENT version by default (override with `--secrets-manager-version-stage`). If JSON keys are upper cased and prefixed with the name of the
  environment variable excluding the prefix. To not include the prefix, use an extra `_` after the prefix.

## Examples

### KMS

```
export KMS_A=<base64>
kms-env program
```

`program` will be called with its environment set to the parent process environment with the additional env var `A` with
its value set to the decrypted value of `KMS_A`. `KMS_A` is not passed to the child process.

### SSM Single parameter

```
export SSM_B=/path/to/value
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
export SSM_C=/path/to/values/*
kms-env program
```

Similar to the previous case, but the environment will have the following variables set
* `C_FOO_BAR`
* `C_FLIP`

### SSM wildcard without prefix

Use a double `_` to ignore the prefix

```
export SSM_C=/path/to/values/*
kms-env program
```

Will have
* `FOO_BAR`
* `FLIP`

If multiple variables have a double `_` they all get merged. (Note: there is no guarantee of the order in which they are processed)

### Secrets Manager with prefix

Assuming the secret `name/of/secret` exists
```
{"foo": 123, "bar": "test"}
```

```
export SECRETS_MANAGER_ABC=name/of/secret
kms-env program
```

Will have
* `ABC_FOO=123`
* `ABC_BAR=test`

### Secrets Manager without prefix

Add an extra `_` in the environment variable name

```
export SECRETS_MANAGER__ABC=name/of/secret
kms-env program
```

Will have
* `FOO=123`
* `BAR=test
