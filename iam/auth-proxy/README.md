# iam-auth-proxy

*BETA do not use for production yet*

```
usage: iam-auth-proxy [<flags>]

Proxy to generate IAM auth token

Flags:
      --help                     Show context-sensitive help (also try --help-long and --help-man).
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
  -v, --version                  Display the version
      --socks-proxy=SOCKS-PROXY  Socks proxy host:port to use.
      --bind=":8080"             Address to bind to
```

## Use case

You want to use IAM as your identity provider for a service as you already use IAM to manage access to AWS.

## Comparison to other AWS Solutions

* Cognito: Cognito lets you use give IAM access to Cognito users but uses its own pool of users, it's the opposite
* AWS SSO: Is not globally available

# Authentication Flow

The proxy runs on the local machine of the user and negotiate an auth cookie with the remote server.

```
   +-----------+                +---------+                +----------+                +-------+
   |  Browser  |                |  Proxy  |                |  Server  |                |  AWS  |
   |           |                |         |                |          |                |       |
   |           |    1. GET /private       |                |          |                |       |
   |           |  +---------->  |         |                |          |                |       |
   |           |                |         |                |          |                |       |
   |           |                |         |    2. GET /private        |                |       |
   |           |                |         |  +---------->  |          |                |       |
   |           |                |         |                |          |                |       |
   |           |                |         |    3. 401 + auth headers  |                |       |
   |           |                |         |  <----------+  |          |                |       |
   |           |                |         |                |          |                |       |
   |           |                |         |    4. STS Get Session Token                |       |
   |           |                |         |  +-------------------------------------->  |       |
   |           |                |         |                |          |                |       |
   |           |                |         |    5. Session Token       |                |       |
   |           |                |         |  <--------------------------------------+  |       |
   |           |                |         |                |          |                |       |
   |           |                |         |    6. KMS Encrypt Session Token            |       |
   |           |                |         |  +-------------------------------------->  |       |
   |           |                |         |                |          |                |       |
   |           |                |         |    7. Encrypted Session Token              |       |
   |           |                |         |  <--------------------------------------+  |       |
   |           |                |         |                |          |                |       |
   |           |    8. 302 /auth + token  |                |          |                |       |
   |           |  <----------+  |         |                |          |                |       |
   |           |                |         |                |          |                |       |
   |           |    9. GET /auth + token  |                |          |                |       |
   |           |  +---------->  |         |                |          |                |       |
   |           |                |         |                |          |                |       |
   |           |                |         |    10. GET /auth + token  |                |       |
   |           |                |         |  +---------->  |          |                |       |
   |           |                |         |                |          |                |       |
   |           |                |         |                |          |    11. KMS Decrypt Session Token
   |           |                |         |                |          |  +---------->  |       |
   |           |                |         |                |          |                |       |
   |           |                |         |                |          |    12. Decrypted Session Token
   |           |                |         |                |          |  <----------+  |       |
   |           |                |         |                |          |                |       |
   |           |                |         |                |          |    13. STS Get Caller Identity
   |           |                |         |                |          |  +---------->  |       |
   |           |                |         |                |          |                |       |
   |           |                |         |                |          |    14. STS Identity    |
   |           |                |         |                |          |  <----------+  |       |
   |           |                |         |                |          |                |       |
   |           |                |         |    15. 302 /private + cookie               |       |
   |           |                |         |  <----------+  |          |                |       |
   |           |                |         |                |          |                |       |
   |           |    16. 302 /private + cookie              |          |                |       |
   |           |  <----------+  |         |                |          |                |       |
   |           |                |         |                |          |                |       |
   |           |    17. GET /private + cookie              |          |                |       |
   |           |  +---------->  |         |                |          |                |       |
   |           |                |         |                |          |                |       |
   |           |                |         |   18. GET /private + cookie                |       |
   |           |                |         |  +---------->  |          |                |       |
   |           |                |         |                |          |                |       |
   |           |                |         |   19. 200      |          |                |       |
   |           |                |         |  <----------+  |          |                |       |
   |           |                |         |                |          |                |       |
   |           |    20. 200     |         |                |          |                |       |
   |           |  <----------+  |         |                |          |                |       |
   |           |                |         |                |          |                |       |
   +-----------+                +---------+                +----------+                +-------+
```

1. The user requests a protected page `/private`
2. The proxy forwards the request to the server
3. The server has not received a valid session so it returns a 401 response with the following headers
   ```
   Www-Authenticate: IAM realm="<aws_account_id (string)>"
   Iam-Auth-Kms-Encryption-Context: <encryption_context (base64 string)>
   Iam-Auth-Kms-Key-Id: <kms_key_id (arn string)
   ```
   * The realm is the AWS account ID where the IAM user should exists.
   * The encryption context is a base64 encoded json object of key/values to give KMS. Its contents are not used by the proxy itself.
   * The KMS Key Id is used to encrypt the auth token.
4. The proxy calls the STS `GetSessionToken` API to get temporary credentials.
6. The proxy serializes and encrypts the temporary credentials with the KMS Key and Encryption Context returned by the server
8. The proxy redirects the browser to the `/auth` endpoint with the `token` and `return_url` arguments
10. When the server receives the auth token it decrypts it with the same encryption context given in the 401
13. The server then uses STS with the client credentials and calls `GetCallerIdentity` to validate them and get the user arn
15. The server issues a session cookie and redirects the user to `return_url`
17. The client fetches the `/private` page successfully

# Security

* The proxy generates temporary credentials with the minimum allowed lifetime (15 minutes).
  The server can check the AWS expiration information returned by STS and decide to further restrict the validity window further.
* The temporary credentials have [no permission](https://docs.aws.amazon.com/STS/latest/APIReference/API_GetSessionToken.html) beyond calling STS `GetCallerIdentity` as they don't have MFA info associated with them.
* The credentials are encrypted with KMS so only the server can decrypt them (See policies in `Setup`).
* The credentials are encrypted with an encryption context that needs to be known to decrypt them so tokens from log files are not useful. The server is free to rotate
  the encryption context as often as needed.
* The server can use any session mechanism

# IAM Setup

You can use IAM users in a different account (for example a centralised identity account shared with multiple other accounts).
You need to add a trust relationship between the KMS key in the central account and each separate account.

## KMS

* Create a KMS key in the account where the server runs
* The role used by the server needs `kms:Decrypt`
* The user needs to have `kms:Encrypt`

*Examples*

* KMS key policy
  ```
     {
        "Sid": "OtherAccountCanEncrypt",
        "Effect": "Allow",
        "Principal": {
          "AWS": [
            "arn:aws:iam::<cental account id>:root"
          ]
        },
        "Action": [
          "kms:Encrypt"
        ],
        "Resource": "*"
      }
  ```
* Server role policy
  ```
  {
     "Sid": "ServerCanDecrypt",
     "Effect": "Allow",
     "Action": [
       "kms:Decrypt",
       "kms:DescribeKey"
     ],
     "Resource": "<key arn>"
   }
  ```
* User policy
  ```
  {
     "Sid": "UserCanEncrypt",
     "Effect": "Allow",
     "Action": [
       "kms:Encrypt"
     ],
     "Resource": "<key arn>"
   }
  ```

## STS

* The user needs to have `sts:GetSessionToken`

# Examples

## Example server

[example-server.py](./example-server.py) has an example basic implementation of a server using flask
```
pip install boto3 flask
AUTH_AWS_ACCOUNT_ID=<account id> AWS_DEFAULT_REGION=eu-west-1 SECRET_KEY=<long random string> KMS_KEY_ID=<kms key arn> KMS_SALT=<long random string> python example-server.py
```

*DO not use for anything serious*

* It does not reduce the lifetime of credentials
* The secrets passed on the command line are visible to other users on the machine
* ...

You can use a navigator to access it through the http proxy. Go to `http://server/` and if successful you should see your user ARN displayed.

## Access a server inside a VPC

```
ssh -D8081 user@bastion
AWS_PROFILE=<profile> iam-auth-proxy --socks-proxy 127.0.0.1:8081
````

# Known limitations

* Does not work with (some version of?) cURL as it doesn't set cookies when receiving a 302.
