# Changelog

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
