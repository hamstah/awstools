# ec2-ip-from-name

Returns a list of IP addresses associated with ec2 instances with a given Name tag.

```
usage: ec2-ip-from-name --name=NAME [<flags>]

Returns a list of instances IP with a given name.

Flags:
      --help                 Show context-sensitive help (also try --help-long and --help-man).
      --name=NAME            Name of the EC2 instance
      --max-results=9        Max number of IPs to return
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
