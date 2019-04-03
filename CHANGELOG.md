# Changelog

## v7.3.0 (27/03/2018)

**New**

* `iam-sync-users`: Manage sudo with IAM tags

## v7.2.0 (26/03/2018)

**New**

* `iam-sync-users`
  * Added locking users not in IAM with `--lock-missing` and `--lock-ignore-user`
  * Added setting groups from IAM tags with `--iam-tags-prefix`
  * Made sudo optional with `--sudo`/`--no-sudo`

## v7.1.0 (19/03/2018)

**New**

* `aws-dump`: Added LastUsed to metadata

**Fix**

* `elb-resolve-alb-external-url`: Fix incorrect resolution

## v7.0.0 (17/01/2019)

**Breaking changes**

* `aws-dump` changed the report argument format

**New**

* `aws-dump` added more resources



## v6.3.0 (08/01/2019)

**New**

* Added `ec2-describe-instances`
* `aws-dump` added more resources

## v6.2.0 (07/01/2019)

**New**

* Added `aws-dump`

## v6.1.0 (21/12/2018)

**New**

* Added `iam-sync-users`

## v6.0.0 (17/12/2018)

**Fix**

* `iam-session` returns the command exit code
* `kms-env` returns the command exit code

## v5.11.0 (03/12/2018)

**New**

* Added `--task-json` to `ecs-deploy`

**Fix**

* `ecs-deploy` would return before the deployment was completed

## v5.10.0 (23/11/2018)

**New**

* Added support for secrets manager in `kms-env`

## v5.9.0 (01/10/2018)

**New**

* Added `ecs-locate`

## v5.8.1 (26/06/2018)

**Fix**

* MFA support in `iam-auth-proxy`

## v5.8.0 (26/06/2018)

**New**

* Added `iam-auth-proxy`

## v5.7.0 (19/06/2018)

**New**

* Added `--version` to all commands except `ecs-dashboard`

## v5.6.0 (19/06/2018)

**New**

* Added SSM support in `kms-env`

## v5.5.0 (13/06/2018)

**New**

* Added `--dns-prefix` to `elb-resolve-alb-external-url` to filter the right dns when there is more than one alias for an ALB

## v5.4.1 (11/06/2018)

**Fix**

* Fix a crash when using `ec2-ip-from-name` when a terminated instance exists

## v5.4 (29/05/2018)

**New**

* `lambda-ping`: Ping a URL with lambda and publish a cloudwatch custom metric

## v5.3 (29/05/2018)

**New**

* `cloudwatch-put-metric-data`: New dimension argument

**Fixes**

* `ecs-dashboard`: Use the same code to open session as the other tools. Fixes an issue where role assumption wasn't working sometimes from ECS.

## v5.2 (01/05/2018)

**New**

* Added `kms-env`

## v5.1 (29/04/2018)

**New**

* Added support for profiles with MFA and role assumption in all tools
* Added a check for trailing spaces in env vars to catch copy paste mistakes
* Added `common.OpenSession` to make it easier to open a session with config

## v5.0 (21/04/2018)

**Breaking Changes**

* `ecs-run-task`: `--cluster-name` renamed to `--cluster` to match `ecs-deploy`

**New**

* Added support for role assumption and MFA

## v4.5 (21/04/2018)

**New**

* Added `iam-public-ssh-keys`

## v4.4 (20/04/2018)

**New**

* Added `iam-session`

## v4.3 (29/01/2018)

**Breaking Changes**

* Changed flag in `ecs-deploy`

**New**

* Added `ecr-get-login`

## v3.2 (28/01/2018)

No changelog for older versions, check commit logs
