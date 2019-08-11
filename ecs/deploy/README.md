# ecs-deploy

Update an ECS service and its task definition with a new image and starts a deployment. It will fetch the existing version
and override the images with the ones in `--image`.

```
usage: ecs-deploy --cluster=CLUSTER --service=SERVICE [<flags>]

Update a task definition on ECS.

Flags:
      --help                 Show context-sensitive help (also try --help-long and --help-man).
      --task-name=TASK-NAME  ECS task name
      --cluster=CLUSTER      ECS cluster
      --service=SERVICE ...  ECS services
      --image=IMAGE ...      Change the images to the new ones. Format is container_name=image. Can be repeated.
      --timeout=300s         Timeout when waiting for services to update
      --task-json=TASK-JSON  Path to a JSON file with the task definition to use
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

# Examples

* Update a container image in a task
  ```
  $ ecs-deploy --cluster prod --task-name web --service web --image proxy=nginx:latest
  ```
* Update mutliple containers in the same task
  ```
  $ ecs-deploy --cluster prod --task-name web --service web --image proxy=nginx:latest --image api=django:latest
  ```
* Update the task JSON
  ```
  $ ecs-deploy --cluster prod --service web --task-json web.json  --image api=django:latest
  ```
