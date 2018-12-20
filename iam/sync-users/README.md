# iam-sync-users

```
Note: Only works on Linux with sudo
```

```
usage: iam-sync-users [<flags>]

Create users from IAM

Flags:
      --help             Show context-sensitive help (also try --help-long and --help-man).
      --assume-role-arn=ASSUME-ROLE-ARN
                         Role to assume
      --assume-role-external-id=ASSUME-ROLE-EXTERNAL-ID
                         External ID of the role to assume
      --assume-role-session-name=ASSUME-ROLE-SESSION-NAME
                         Role session name
      --region=REGION    AWS Region
      --mfa-serial-number=MFA-SERIAL-NUMBER
                         MFA Serial Number
      --mfa-token-code=MFA-TOKEN-CODE
                         MFA Token Code
  -v, --version          Display the version
      --group=GROUP ...  Add users from this group. You can use --group multiple times.
```

## IAM policy

You can use the `arn:aws:iam::aws:policy/IAMReadOnlyAccess` managed policy or use the custom one below

```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "iam:GetGroup"
            ],
            "Resource": "*"
        }
    ]
}
```
