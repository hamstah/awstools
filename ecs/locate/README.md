# ecs-locate

Find the IPs and host ports of a container port running in ECS. It returns the `ip:port` for every running ECS task in the cluster with the given name.

```
usage: ecs-locate --container-name=CONTAINER-NAME --container-port=CONTAINER-PORT --cluster=CLUSTER --service=SERVICE [<flags>]

Find an instance/port for a service

Flags:
      --help                 Show context-sensitive help (also try --help-long and --help-man).
      --container-name=CONTAINER-NAME
                             ECS container name
      --container-port=CONTAINER-PORT
                             ECS container port
      --cluster=CLUSTER      ECS cluster
      --service=SERVICE      ECS service
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
