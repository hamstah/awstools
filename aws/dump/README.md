# aws-dump

Dumps AWS resources metadata to JSON and optionally check if they are managed by Terraform.

```
usage: aws-dump --accounts-config=ACCOUNTS-CONFIG --output=OUTPUT [<flags>]

Dump AWS resources

Flags:
      --help                 Show context-sensitive help (also try --help-long and --help-man).
  -c, --accounts-config=ACCOUNTS-CONFIG
                             Configuration file with the accounts to list resources for.
  -t, --terraform-backends-config=TERRAFORM-BACKENDS-CONFIG
                             Configuration file with the terraform backends to compare with.
  -o, --output=OUTPUT        Filename to store the results in.
      --only-unmanaged       Only return resources not managed by terraform.
      --report=REPORT ...    Only run the specified report. Can be repeated.
      --list-reports         Prints the list of available reports and exits.
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

## Supported resources

You can see available reports with `--list-reports`.

```
acm:certificates
autoscaling:groups
autoscaling:launch-configurations
cloudwatch:alarms
ec2:images
ec2:instances
ec2:key-pairs
ec2:launch-templates
ec2:nat-gateways
ec2:security-groups
ec2:vpcs
iam:groups
iam:instance-profiles
iam:policies
iam:roles
iam:users-and-access-keys
kms:aliases
kms:keys
lambda:event-source-mappings
lambda:functions
rds:db-clusters
rds:db-instance-automated-backups
rds:db-instances
rds:db-parameter-groups
rds:db-security-groups
rds:db-snapshots
rds:db-subnet-groups
rds:event-subscriptions
rds:events
rds:global-clusters
rds:option-groups
rds:reserved-db-instances
route53:zones-and-records
s3:buckets
```

## Configuration

### AWS Accounts

Create a JSON file with the following structure

```js
{
  "accounts": [
    {
      "role_arn": "arn:aws:iam::123456789012:role/Role",
      "regions": ["us-east-1", "us-east-2", "eu-west-1"]
    },
    {
      "role_arn": "arn:aws:iam::234567890123:role/Role",
      "regions": ["us-east-1"]
    }
  ]
}
```

Then pass the filename to the `--accounts-config` flag.

### Terraform

Currently only S3 backends are supported.

#### s3 backends

Create a JSON file with the following structure

```
{
  "destination": "./terraform-states/",
  "options": {
    "path_substitutions": [
      {
        "old": "/",
        "new": "-"
      }
    ],
    "overwrite": false
  },
  "s3":[
    {
      "bucket":"terraform-bucket",
      "keys":[
        "test.tfstate",
        "prod.tfstate"
      ],
      "region":"eu-west-1",
      "role_arn":"arn:aws:iam::123456789012:role/Role"
    }
  ]
}
```

## Output

The output file contains a JSON array of resources

```
[
  ...
  {
    "id": "test-bucket",
    "arn": "arn:aws:s3:::test-bucket",
    "service": "s3",
    "type": "bucket",
    "account_id": "123456789012",
    "region": "",
    "metadata": null,
    "managed_by": {
      "state": "arn:aws:s3:::terraform-bucket/test.tfstate",
      "type": "terraform"
    }
  },
  {
    "id": "prod-bucket",
    "arn": "arn:aws:s3:::prod-bucket",
    "service": "s3",
    "type": "bucket",
    "account_id": "123456789012",
    "region": "",
    "metadata": null,
    "managed_by": null,
  },
  ...
]
```

If `--only-unmanaged` is used only resources with `managed_by: null` will be returned.
