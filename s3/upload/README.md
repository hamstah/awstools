# s3-upload

Upload a single file to S3.

```
usage: s3-upload --bucket=BUCKET --key=KEY --file=FILE [<flags>]

Upload a file to S3.

Flags:
      --help                 Show context-sensitive help (also try --help-long and --help-man).
      --bucket=BUCKET        Name of the bucket
      --key=KEY              Key of the uploaded file
      --file=FILE            File to upload
      --acl=ACL              ACL of the uploaded file
      --metadata=METADATA    Metadata of the uploaded file
      --assume-role-arn=ASSUME-ROLE-ARN
                             Role to assume
      --assume-role-external-id=ASSUME-ROLE-EXTERNAL-ID
                             External ID of the role to assume
      --assume-role-session-name=ASSUME-ROLE-SESSION-NAME
                             Role session name
      --assume-role-policy=ASSUME-ROLE-POLICY
                             IAM policy to use when assuming the role
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
