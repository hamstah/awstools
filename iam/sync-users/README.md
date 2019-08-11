# iam-sync-users

```
Note: Only works on Linux with sudo
```

```
usage: iam-sync-users [<flags>]

Sync local users with IAM

Flags:
      --help                 Show context-sensitive help (also try --help-long and --help-man).
      --group=GROUP ...      Add users from this IAM group. You can use --group multiple times.
      --iam-tags-prefix="iam-sync-users"
                             Prefix for tags in IAM
      --lock-missing         Lock local users not in IAM.
      --lock-ignore-user=LOCK-IGNORE-USER ...
                             Ignore local user when locking.
      --sudo                 Add users to sudoers file.
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
                "iam:GetGroup"
            ],
            "Resource": "*"
        }
    ]
}
```


## Configuration

### Sudo

There are two options to manage sudo
* Use the `--sudo` flag to give sudo access to all users in the IAM groups
* Use the `--no-sudo` flag and use IAM tags on the users to give sudo only to specific users
  * Pick a tag prefix for the machine, for example your environment, for example com.company.prod
  * Add the `com.company.prod:sudo` tag with value `true` to the users who should have sudo
  * call `iam-sync-users` with `--no-sudo` and `--iam-tags-prefix com.company.prod`

### Linux groups

Similar to sudo, we can set the groups the users are in with a tag on the IAM user.

The IAM tag name is `<iam-tags-prefix>:groups`, for example `--iam-tags-prefix com.company.prod:groups`
The IAM tag value is a space separated list of Linux groups, for example `wheel docker`

### User locking

Use `--lock-missing` and `--lock-ignore-user` to enable locking of users account not found in IAM.

Locking will do two things
- Lock the user with usermod to prevent password logins
- Expire the user to prevent to prevent other logs with PAM such as SSH

Use `--lock-ignore-user` to exclude certain users from the users to lock. You probably want to specify `--lock-ignore-user ubuntu` or `--lock-ignore-user ec2-user` when using EC2.

The locking only applies to non system users, based on the `MIN_UID` and `MAX_UID` values in `/etc/login.defs`.

## Notes

* If your IAM usernames have a `.` in them, you might need to configure `/etc/adduser.conf` to allow it. Centos/Amazon Linux should work, Ubuntu (tested in 18.04) needs the change below.
  * Edit `/etc/adduser.conf`
  * Look for `NAME_REGEX` and see if it's commented.
  * Uncomment or insert ```NAME_REGEX='^[a-z][-.a-z0-9.]*$'```
