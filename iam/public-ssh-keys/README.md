# iam-public-ssh-keys

```
usage: iam-public-ssh-keys [<flags>]

Return public SSH keys for an IAM user.

Flags:
      --help                 Show context-sensitive help (also try --help-long and --help-man).
  -u, --username=USERNAME    Username to fetch the keys for, otherwise default to the logged in user.
      --key-encoding=SSH     Encoding of the key to return (SSH or PEM)
      --allowed-group=ALLOWED-GROUP ...
                             Fetch the keys only if the user is in this group. You can use --allowed-group multiple times.
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

## IAM policy

You can use the `arn:aws:iam::aws:policy/IAMReadOnlyAccess` managed policy or use the custom one below

```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "iam:GetSSHPublicKey",
                "iam:ListSSHPublicKeys",
                "iam:ListGroupsForUser"
            ],
            "Resource": "*"
        }
    ]
}
```

## examples

### Use IAM SSH keys to connect to instances

* Make sure the binary is in `PATH` or make the path in the script below absolute
* Create a script `/usr/local/bin/ssh-authorized-keys.sh`
```
#!/usr/bin/env sh
iam-public-ssh-keys --username $1
```
* Make it executable
```
chmod +x /usr/local/bin/ssh-authorized-keys.sh
```
* Edit your `sshd` configuration (usually `/etc/sshd/sshd_config`) and add the lines. You can change the user, don't run it as root.
```
AuthorizedKeysCommand /usr/local/bin/ssh-authorized-keys.sh
AuthorizedKeysCommandUser ec2-user
```

Note: the user needs to already exist on the system.

You can use the other options in the script
* If the instance profile does not have the policy to access the SSH keys directly, just use `--assume-role-arn`. You probably need this when accessible users in a different AWS account.
* Restrict SSH access on the instance to users of a specific IAM group, add `--allowed-group`, here are some ideas to avoid hardcoding the group name so you can reuse the script
  * Have a `SSH_ALLOWED_GROUP` environment varible set in the userdata or AMI
  * Get the allowed group value from the instance tags
