# aws-dump

Dumps AWS resources metadata to JSON and optionally check if they are managed by Terraform.

```
usage: aws-dump --accounts-config=ACCOUNTS-CONFIG --output=OUTPUT [<flags>]

Dump AWS resources

Flags:
      --help            Show context-sensitive help (also try --help-long and --help-man).
      --assume-role-arn=ASSUME-ROLE-ARN  
                        Role to assume
      --assume-role-external-id=ASSUME-ROLE-EXTERNAL-ID  
                        External ID of the role to assume
      --assume-role-session-name=ASSUME-ROLE-SESSION-NAME  
                        Role session name
      --region=REGION   AWS Region
      --mfa-serial-number=MFA-SERIAL-NUMBER  
                        MFA Serial Number
      --mfa-token-code=MFA-TOKEN-CODE  
                        MFA Token Code
  -v, --version         Display the version
  -c, --accounts-config=ACCOUNTS-CONFIG  
                        Configuration file with the accounts to list resources for.
  -t, --terraform-backends-config=TERRAFORM-BACKENDS-CONFIG  
                        Configuration file with the terraform backends to compare with.
  -o, --output=OUTPUT   Filename to store the results in.
      --only-unmanaged  Only return resources not managed by terraform.
```

## Supported resources

* EC2
  * VPC
  * Security Groups
* IAM (Does not include attachments)
  * Users
  * Access keys
  * Roles
  * Policies
* S3
  * Buckets

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
