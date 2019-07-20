# lambda-sign-ssh-key

```
usage: lambda-sign-ssh-key [<flags>]

Signs SSH keys.

Flags:
      --help                 Show context-sensitive help (also try --help-long and --help-man).
      --config-filename-template="%s.json"
                             Filename of the configuration file.
      --event-filename=EVENT-FILENAME
                             Filename with the event payload. Will process the event and exit if present.
      --identity-url-max-age=10s
                             Maximum age of the identity URL signature
      --assume-role-arn=ASSUME-ROLE-ARN
                             Role to assume
      --assume-role-external-id=ASSUME-ROLE-EXTERNAL-ID
                             External ID of the role to assume
      --assume-role-session-name=ASSUME-ROLE-SESSION-NAME
                             Role session name
      --region=REGION        AWS Region
      --mfa-serial-number=MFA-SERIAL-NUMBER
                             MFA Serial Number
      --mfa-token-code=MFA-TOKEN-CODE
                             MFA Token Code
      --session-duration=1h  Session Duration
  -v, --version              Display the version
      --log-level=warn       Log level
      --log-format=text      Log format
```

This lambda is to be used with [iam-request-ssh-key-signature](../../iam/request-ssh-key-signature) to generate
SSH key certificates.

## Configuration

### CA keys

The lambda will sign the ssh keys with a keypair. Follow the instructions from [Create SSH CA Certificate Signing Keys](https://access.redhat.com/documentation/en-us/red_hat_enterprise_linux/6/html/deployment_guide/sec-creating_ssh_ca_certificate_signing-keys)
to generate the keypair. The lambda will need access to both the public and private key while the SSH servers will need access to the public key only. You should
store them separately in order to grant different permissions for the lambda and the servers.

The keys can be looked up from either AWS Secrets Manager, AWS SSM or decrypted from AWS KMS (encrypted with `kms:Encrypt` or encrypted with `kms:GenerateDataKey` and secret box).

### Environent configuration file

The signing request will contain an `environment` value which will be used to determine which parameters to use when signing the certificate.

The environent configuration is looked up from a json file with the following structure

```
{
    "ca": {
      "private_key": "<private key>",
      "public_key": "<public key>",
      "private_key_passphrase": "<optional passphrase>"
    },
    "validity_max_duration": <max duration for the certificate in seconds>,
    "validity_start_offset": <offset in seconds for the valid after certificate to deal with clock drift>,
    "source_addreses": [
      "<source ip cidr restriction>",
      ...
    ]
}
```

Each string parameter can be loaded from Secret Manager, SSM Parameter Store, plain text file or decrypted from KMS using
the corresponding prefix

* `ssm://parameter-name`
* `secrets-manager://name-of-secret`
* `file://name-of-file`
* `kms://base64-blob`

Note that `secret-manager://` and `ssm://` with a wildcard will expand to a json object.

## Deploying

The lambda function does not need any particular permissions. If you use KMS, Secrets Manager or SSM the lambda needs
a role with a policy allowing access to those resources.

[terraform.tf](./terraform.tf) gives an example of building and deploying the lambda using Secrets Manager to manage the CA
configuration. Apply terraform then set the fields in Secrets Manager. The private key should be base64 encoded to avoid issues
with new lines.

## Configuring the SSH servers

* Store the CA public key on your SSH server under `/etc/ssh/ca_user_key.pub`
* Edit `/etc/ssh/sshd_config` and append the following line `TrustedUserCAKeys /etc/ssh/ca_user_key.pub`
* Reload the SSH service `sudo /etc/init.d/ssh reload` or `sudo systemctl reload ssh`

## Creating the SSH users

The SSH certificate will use the IAM username as principal so the users need to exist on the SSH server.
The easiest way to do it is to use [iam-sync-users](../sync-users) on a CRON job.
