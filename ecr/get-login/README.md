# ecr-get-login

Get an username/password to authenticate with an ECR repository. The default output is the docker login command to use to match the same command
in awscli. There is also an option to output the raw credentials to use in other tools.

```
usage: ecr-get-login [<flags>]

Returns an authorization token from ECR.

Flags:
      --help           Show context-sensitive help (also try --help-long and --help-man).
      --assume-role-arn=ASSUME-ROLE-ARN  
                       Role to assume
      --assume-role-external-id=ASSUME-ROLE-EXTERNAL-ID  
                       External ID of the role to assume
      --assume-role-session-name=ASSUME-ROLE-SESSION-NAME  
                       Role session name
      --region=REGION  AWS Region
      --mfa-serial-number=MFA-SERIAL-NUMBER  
                       MFA Serial Number
      --mfa-token-code=MFA-TOKEN-CODE  
                       MFA Token Code
  -v, --version        Display the version
      --output=shell   Return the credentials instead of docker command
```
