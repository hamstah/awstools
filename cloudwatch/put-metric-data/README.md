# cloudwatch-put-metric-data

Put a single cloudwatch metric value.

```
usage: cloudwatch-put-metric-data --metric-name=METRIC-NAME --namespace=NAMESPACE --value=VALUE [<flags>]

Put a cloudwatch metric value.

Flags:
      --help                     Show context-sensitive help (also try --help-long and --help-man).
      --metric-name=METRIC-NAME  Name of the Cloudwatch metric
      --namespace=NAMESPACE      Name of the Cloudwatch namespace
      --dimension=DIMENSION ...  Dimensions name=value
      --value=VALUE              Metric value
      --assume-role-arn=ASSUME-ROLE-ARN
                                 Role to assume
      --assume-role-external-id=ASSUME-ROLE-EXTERNAL-ID
                                 External ID of the role to assume
      --assume-role-session-name=ASSUME-ROLE-SESSION-NAME
                                 Role session name
      --region=REGION            AWS Region
      --mfa-serial-number=MFA-SERIAL-NUMBER
                                 MFA Serial Number
      --mfa-token-code=MFA-TOKEN-CODE
                                 MFA Token Code
      --session-duration=1h      Session Duration
  -v, --version                  Display the version
      --log-level=warn           Log level
      --log-format=text          Log format

```
