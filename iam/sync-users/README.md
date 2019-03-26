# iam-sync-users

```
Note: Only works on Linux with sudo
```

```
usage: iam-sync-users [<flags>]

Sync local users with IAM

Flags:
      --help                 Show context-sensitive help (also try --help-long and --help-man).
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
      --group=GROUP ...      Add users from this IAM group. You can use --group multiple times.
      --iam-tags-prefix=IAM-TAGS-PREFIX
                             Prefix for tags in IAM
      --lock-missing         Lock local users not in IAM.
      --lock-ignore-user=LOCK-IGNORE-USER ...
                             Ignore local user when locking.
      --sudo                 Add users to sudoers file.
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
