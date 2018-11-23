# ecs-run-task

Runs an ECS task

```
usage: ecs-run-task --task-definition=TASK-DEFINITION --cluster=CLUSTER [<flags>]

Run a task on ECS.

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
      --task-definition=TASK-DEFINITION  
                         ECS task definition
      --cluster=CLUSTER  ECS cluster
```
