# s3-download

Download a single file from S3.

```
usage: s3-download --bucket=BUCKET --key=KEY --filename=FILENAME [<flags>]

Download a file from S3.

Flags:
      --help                 Show context-sensitive help (also try --help-long and --help-man).
      --bucket=BUCKET        Name of the bucket
      --key=KEY              Key to download
      --filename=FILENAME    Output filename
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
