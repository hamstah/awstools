# ecs-dashboard

## Overview

Creates a basic dashboard of ECS services between multiple accounts.

```
usage: ecs-dashboard [<flags>]

Flags:
      --help                 Show context-sensitive help (also try --help-long and --help-man).
  -c, --config-file="config.json"
                             Config filename
  -p, --port=8000            Server port
  -r, --refresh-interval=30  Refresh interval (seconds)
      --open                 Open browser
```

## Features

### Version tracking

If you have 3 AWS accounts `test`, `staging` and `prod` you want to see which version
runs where and know what needs deploying. List the accounts in order of the pipeline in
your config file and the dashboard will group the services by cluster and service name
and highlight services where the image does not match the image of the previous account.

### Health checks

The dashboard shows 2 metrics related to health
* Number of running tasks. It will show an alert if the running count does not match the desired count.
* Last ECS event for the service. It will show an alert if the last message is
  not about the service being in a steady state. This will show during a deployment.



## Configuration

The service loads its configuration from a file which needs to have the following format

```
{
   "accounts":[
      {
          "role":"arn:aws:iam::123456789012:role/ecs-monitor",
          "external_id": "",
          "account_name":"test",
          "prefix": "test-",
	        "region": "eu-west-1"
      },
      ...
   ]
}
```

Note: Only one region is supported right now but you can repeat the same account
with a different name and region.

The configuration will automatically reload on change.

## Docker

A docker image of the service is available on [Docker Hub](https://hub.docker.com/r/hamstah/ecs-dashboard/)

```
docker pull hamstah/ecs-dashboard
```

The image expects the config file to be `/usr/share/config.json`

Here is how to start it on port 8000 with a `config.json` in the current directory:

```
docker run -v $PWD/config.json:/usr/share/config.json -p "8000:8000" hamstah/ecs-dashboard
```

You can also create your own image with your config file in it if you don't want to use volumes.

```
FROM hamstah/ecs-dashboard:latest

COPY config.json /usr/share
```

## IAM Permissions

The roles in the config must have the following permissions

```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "",
            "Effect": "Allow",
            "Action": [
                "ecs:ListServices",
                "ecs:ListClusters",
                "ecs:DescribeTaskDefinition",
                "ecs:DescribeServices",
                "ecs:DescribeClusters"
            ],
            "Resource": "*"
        }
    ]
}
```
