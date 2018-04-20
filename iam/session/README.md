# iam-session

```
usage: iam-session [<flags>] [<command>...]

Start a new session under a different role.

Flags:
      --help               Show context-sensitive help (also try --help-long and --help-man).
      --assume-role-arn=ASSUME-ROLE-ARN  
                           Role to assume
      --assume-role-external-id=ASSUME-ROLE-EXTERNAL-ID  
                           External ID of the role to assume
      --region=REGION      AWS Region
      --mfa-serial-number=MFA-SERIAL-NUMBER  
                           MFA Serial Number
      --mfa-token-code=MFA-TOKEN-CODE  
                           MFA Token Code
  -q, --quiet              Do not output anything
  -s, --save-profile=SAVE-PROFILE  
                           Save the profile in the AWS credentials storage
      --overwrite-profile  Overwrite the profile if it already exists

Args:
  [<command>]  Command to run, prefix with -- to pass args
```

## Features

* Supports the standard AWS authentication
  * credentials file (choose different profiles with `AWS_PROFILE`)
  * environment variables
  * instance profiles
* Use it to assume role between different AWS accounts

## Examples

Here are some examples of invocation. All parameters can be used together.

### Run a command under a different role

```
iam-session --assume-role-arn arn:aws:iam::123456789012:role/my-role aws ec2 describe-instances
```
Will describe ec2 instances accessible to the role `my-role`. When the command is run the session doesn't exist anymore.
The command has its environment set to the parent environment with the new AWS environment variables injected.

### Open a new shell authenticated as a different role

```
iam-session --assume-role-arn arn:aws:iam::123456789012:role/my-role bash
```

Will open a new shell under a new session for `my-role`. Like the previous example the AWS environment variables will be set.
Note that the session is only valid for 15 minutes.

### Create a new temporary session with MFA

```
iam-session --mfa-serial-number arn:aarn:aws:iam::123456789012:mfa/nico/my-mfa bash
```

The new session will have its MFA age set in IAM. Valid for 15 mins. You will be prompted for the MFA token code, but you can also pass it with `--mfa-token-code`

```
iam-session --mfa-serial-number arn:aarn:aws:iam::123456789012:mfa/nico/my-mfa --mfa-token-code 012345 bash
```

### Save a temporary session to the AWS credentials file to reuse it

```
iam-session --mfa-serial-number arn:aarn:aws:iam::123456789012:mfa/nico/my-mfa --save-profile new-profile
aws --profile=new-profile ec2 describe-instances
```

The new profile will be added to `~/.aws/credentials` and `~/.aws/config`

If the profile already exists you will be prompted to confirm its replacement. You can avoid the prompt by using `--overwrite-profile`
