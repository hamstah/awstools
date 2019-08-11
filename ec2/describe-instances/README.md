# ec2-describe-instances

Describe instances from either instance ids or a filter. If called with no argument, it will resolve the instance id from the metadata service.

```
usage: ec2-describe-instances [<flags>] [<identifiers>...]

Returns metadata of one or more EC2 instances

Flags:
      --help                 Show context-sensitive help (also try --help-long and --help-man).
      --filter=FILTER        The filter to use for the identifiers. eg tag:Name
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

Args:
  [<identifiers>]  If omitted the instance is fetched from the EC2 metadata.
```

## Output

```
[
  {
    "instance_id": "i-0123456789012345",
    "iam_instance_profile": "arn:aws:iam::123456789012:instance-profile/base",
    "private_dns_name": "ip-10-0-0-1.eu-west-1.compute.internal",
    "private_ip_address": "10.0.0.1",
    "public_dns_name": "ec2-39-119-120-1.eu-west-1.compute.amazonaws.com",
    "public_ip_address": "39.119.120.1",
    "image_id": "ami-0123456789012345",
    "instance_type": "t2.micro",
    "key_name": "root",
    "launch_time": "2018-09-13T19:24:34Z",
    "subnet_id": "subnet-1234501",
    "tags": {
      "Environment": "test",
      "Name": "bastion"
    },
    "security_groups": [
      {
        "id": "sg-0843026f",
        "name": "bastion-ssh"
      }
    ],
    "state": "running",
    "vpc_id": "vpc-087456d",
    "auto_scaling_group": ""
  }
]
```

## Usage

### Describe current instance

```
$ ec2-describe-instance
```

### Describe multiple instances by id

```
$ ec2-describe-instances instanceid1 instanceid2
```

### Describe by filter

```
$ ec2-describe-instances --filter tag:Name name-of-instance
```
