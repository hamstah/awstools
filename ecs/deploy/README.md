# ecs-deploy

Update an ECS service and its task definition with a new image and starts a deployment. It will fetch the existing version
and override the images with the ones in `--image`.

```
usage: ecs-deploy --task-name=TASK-NAME --cluster=CLUSTER --service=SERVICE [<flags>]

Update a task definition on ECS.

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
  -v, --version              Display the version
      --task-name=TASK-NAME  ECS task name
      --cluster=CLUSTER      ECS cluster
      --service=SERVICE ...  ECS services
      --image=IMAGE ...      Change the images to the new ones. Container name=image
      --timeout=300s         Timeout when waiting for services to update
```
