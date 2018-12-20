# AWS tools

Collection of tools to make working with AWS a bit easier without having to depend on `awscli` and python.

## List of tools

| Tool                                                     | Overview                                                                                                        |
|----------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------|
| [iam-session](iam/session/)                              | Creates new IAM session with role assumption and MFA support.                                                   |
| [iam-public-keys](iam/public-ssh-keys)                   | Returns the public SSH keys of an IAM user.                                                                     |
| [iam-auth-proxy](iam/auth-proxy)                         | Use IAM as identity provider for services.                                                                      |
| [iam-sync-users](iam/sync-users)                         | Create Linux users from IAM                                                                                     |
| [cloudwatch-put-metric-data](cloudwatch/put-metric-data) | Basic sending a metric value to cloudwatch                                                                      |
| [ec2-ip-from-name](ec2/ip-from-name)                     | Given an EC2 name, list up to `-max-results` IPs associated with instances with that name                       |
| [ecr-get-login](ecr/get-login)                           | Prints out the command to run to auth with docker ECR. Check output flag for other options                      |
| [ecs-dashboard](ecs/dashboard)                           | Shows ECS services and their version across multiple AWS accounts.                                              |
| [ecs-locate](ecs/locate)                                 | Returns `ip:port` for containers of an ECS service                                                              |
| [ecs-deploy](ecs/dashboard)                              | Update the container images of a task and update services to use it                                             |
| [ecs-run-task](ecs/run-task)                             | Runs a task definition                                                                                          |
| `elb-resolve-elb-external-url`                           | ELB classic only (no ALB). Given a name returns the zone53 record associated with the ELB, including scheme (https returned if both available) and port.|
| `elb-resolve-alb-external-url`                           | Both ELB classic and ALB. Given a name, returns route53 record associated with the ELB. Does not include scheme or port as it doesn't check listeners. |
| `lambda-ping`                                            | Pings a URL with lambda and publish a custom cloudwatch metric with the result.                                 |
| `s3-download`                                            | Download a single file from s3.                                                                                 |
| [kms-env](kms/env/)                                      | Decrypts environment variables from SSM, KMS or Secret Manager and runs a command.                              |

## Authentication

Every tool supports the standard AWS authentication as well as sts sessions with the following options

* `--region`: Choose the aws-region to use
* `--assume-role-arn`: Assume the role before running. This is useful for cross account access.
* `--mfa-serial-number`: The new session will have its 2FA flag set.
* `--mfa-token-code`: The token code to use when using `--mfa-serial-number`. If not provided the tool will prompt for it.

## Releases

The releases are only available for linux amd64 at the moment.

### Checking release signatures

Download the signature from the release and use GPG to verify it

```
#!/usr/bin/env bash
version=5.0

wget https://github.com/hamstah/awstools/releases/download/v${version}/ec2-ip-from-name
wget https://github.com/hamstah/awstools/releases/download/v${version}/ec2-ip-from-name.asc
gpg --verify ec2-ip-from-name.asc ec2-ip-from-name
```

The signing key is

```
Primary key fingerprint: 5FC5 40A9 A2F2 B87B 9C49  3D9E 7D40 F516 7D5C 7058
```
