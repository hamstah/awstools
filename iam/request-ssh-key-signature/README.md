# iam-request-ssh-key-signature

```
usage: iam-request-ssh-key-signature --lambda-arn=LAMBDA-ARN --ssh-public-key-filename=SSH-PUBLIC-KEY-FILENAME [<flags>]

Request a signature for a SSH key from lambda-sign-ssh-key.

Flags:
      --help                   Show context-sensitive help (also try --help-long and --help-man).
      --lambda-arn=LAMBDA-ARN  ARN of the lambda function signing the SSH key.
      --ssh-public-key-filename=SSH-PUBLIC-KEY-FILENAME
                               Path to the SSH key to sign.
      --environment=""         Name of the environment to sign the key for.
      --duration=1m            Duration of validity of the signature.
      --dump                   Dump the event JSON instead of calling lambda
      --source-address=SOURCE-ADDRESS ...
                               Set the IP restriction on the cert in CIDR format, can be repeated
      --assume-role-arn=ASSUME-ROLE-ARN
                               Role to assume
      --assume-role-external-id=ASSUME-ROLE-EXTERNAL-ID
                               External ID of the role to assume
      --assume-role-session-name=ASSUME-ROLE-SESSION-NAME
                               Role session name
      --region=REGION          AWS Region
      --mfa-serial-number=MFA-SERIAL-NUMBER
                               MFA Serial Number
      --mfa-token-code=MFA-TOKEN-CODE
                               MFA Token Code
      --session-duration=1h    Session Duration
  -v, --version                Display the version
      --log-level=warn         Log level
      --log-format=text        Log format
```
