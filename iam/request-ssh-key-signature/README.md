# iam-request-ssh-key-signature

```
usage: iam-request-ssh-key-signature --lambda-arn=LAMBDA-ARN [<flags>]

Request a signature for a SSH key from lambda-sign-ssh-key.

Flags:
      --help                   Show context-sensitive help (also try --help-long and --help-man).
      --lambda-arn=LAMBDA-ARN  ARN of the lambda function signing the SSH key.
      --ssh-private-key-filename=SSH-PRIVATE-KEY-FILENAME
                               Path to the SSH key to add to the agent.
      --ssh-public-key-filename=SSH-PUBLIC-KEY-FILENAME
                               Path to the SSH key to sign.
      --environment=""         Name of the environment to sign the key for.
      --duration=1m            Duration of validity of the signature.
      --dump                   Dump the event JSON instead of calling lambda.
      --output=agent           Where to store the generated certificate.
      --source-address=SOURCE-ADDRESS ...
                               Set the IP restriction on the cert in CIDR format, can be repeated.
      --proxy-config=PROXY-CONFIG
                               Configuration for the ssh ProxyCommand host:port.
      --assume-role-arn=ASSUME-ROLE-ARN
                               Role to assume
      --assume-role-external-id=ASSUME-ROLE-EXTERNAL-ID
                               External ID of the role to assume
      --assume-role-session-name=ASSUME-ROLE-SESSION-NAME
                               Role session name
      --region=REGION          AWS Region
      --mfa-serial-number=MFA-SERIAL-NUMBER
                               MFA Serial Number
      --mfa-token-code=MFA-TOKEN-CODE
                               MFA Token Code
      --session-duration=1h    Session Duration
  -v, --version                Display the version
      --log-level=warn         Log level
      --log-format=text        Log format
```

## Setup

Create the [lambda-sign-ssh-key function](../../lambda/sign-ssh-key) Lambda with the associate config.

## Usage

### Default keys and ssh agent

The simplest usage is to sign the default SSH keys, the only required arguments are the Lambda ARN and the environment name:

```
$ iam-request-ssh-key-signature --environment prod \
  --lambda-arn arn:aws:lambda:eu-west-1:123456789012:function:sign-ssh-key
```

If all works the command won't print any output and the keys should be in the ssh agent. You can check the certificate is there with

```
$ ssh-add -L | ssh-keygen -L -f -
(stdin):1 is not a certificate
(stdin):2:
        Type: ssh-rsa-cert-v01@openssh.com user certificate
        Public key: RSA-CERT SHA256:CW8pTPIANte2HWEt+bwUs/DoH46utEEXQ5vUV60nMVg
        Signing CA: RSA SHA256:Bg3ycPpoLNi/NTT2nxgG5g5KMrO8ELMV0IBgnUZwACM
        Key ID: "arn:aws:iam::123456789012:user/hamstah/29c50c35-c5aa-67e7-c956-ca3bf2949f3b"
        Serial: 6878956844034323525
        Valid: from 2019-07-25T01:35:33 to 2019-07-25T01:37:33
        Principals:
                hamstah
        Critical Options:
                source-address 127.0.0.1/32,172.17.0.0/16
        Extensions:
                permit-pty
```

You can see the default values from the Lambda config are used for the source address and expiry.

### Store the certificate outside the ssh agent

Change the `--output` option to `stdout`

```
$ iam-request-ssh-key-signature \
  --environment prod   \
  --lambda-arn arn:aws:lambda:eu-west-1:123456789012:function:sign-ssh-key \
  --output stdout | jq .
{
  "certificate": "ssh-rsa-cert-v01@openssh.com AAAAHHNzaC1yc2EtY2VydC12MDFAb3BlbnNzaC5jb20AAAAgDCTmcNmWGQ+BSjhtSKYc4QyE8q1piiXAICHm16VLm1wAAAADAQABAAABAQC+agr/LBc2jrHjMdfmCnL7gy/zSB3WmIcpdCwUbadSRJc+ceLttVqoib7wf9VT5HqSWqlxdF0n9Ihh98jxwJZryM8LRM/IcW0LKtFKGJaSOtW6E0X6+G45TpCzyy2R5Vz7xf4zaZ3i784bDuUsjtimfGA3JYP8enfMmCHDXSORA/wL2mEyQsiPi7Bo+lom/qg8CGGlSqv+/S1yydL7F07nmYMytIVpby3Xv5355RuDEE2f+endrskv/QALyJzwBIjhZHD0ed5TNyNehPa89kTk9TblqZUPttBTu8fOHnihozejRoZp3jjDGR0jD8Nkvh3T46ACyuQcZXjWShQHvUFJtuqM+IzOAQwAAAABAAAAS2Fybjphd3M6aWFtOjozMzA0Mjg5MTM2ODM6dXNlci9oYW1zdGFoLzc2Mzk5OWM1LTU1NjUtMjk0Ni1jMTRjLWExMGE3NzVhNjA4ZQAAAAsAAAAHaGFtc3RhaAAAAABdOOwFAAAAAF047H0AAABCAAAADnNvdXJjZS1hZGRyZXNzAAAALAAAACgxMjcuMC4wLjEvMzIsMTcyLjE3LjAuMC8xNiw5MS4xMjYuMC4wLzE2AAAAEgAAAApwZXJtaXQtcHR5AAAAAAAAAAAAAAEXAAAAB3NzaC1yc2EAAAADAQABAAABAQDBrCsmu3dTNJVHL+/3VVWYBnvgUTvanLrmuG7kCYrMx16AUGFIeoUVV9ulzu+M+DPqGb9pNBihAENpFm4Zfo2e3pPqkP3mpqfOkfdLxpw72zsadKhTClJue0Kd45d/0aPDFB1S6Ok9O0RCmy5720CHGdFY0ocRRpfQaGakI8+xQxuJ26FYAGns0VmoQFv7SZII/WDV00OotWo4908Qc+OMKbwaWH5c3gpzfID6x+XUG/+wesK5WmJ6mVWeGFfHhbugLdfWcXcNfsHVrL/KdUEQRjCMXpy7EUSAnf2WwEnLaGH42G9jeVcUBNED/aNI9QG7sl4YY5bDkgbU3bHKuRN1AAABDwAAAAdzc2gtcnNhAAABAHFMtUjn/DfdR9Hm8sNd77lIZqXz45zqnDZbwMGEWUiNz5js33FK0YmdkO0RwZb6/18X0JyO8jUZF6Cqwc63pkii65WRQpmHQfYWgT2BoL6G6A5h8XzvqYNX5yDUtwLp3wp0wuKuFU0o0u08jCpaoOsdQ53dJTchaz13OwYFgCCr8I8Smvl1BoRw0Mlb2/vxxA99mtAKagmgy3efey1l+cjctLS7dpEVp64T/mREM/d37ookgbqa04Kqc/brM/tty7s9rkeO5g53WWrSL4v7Kb5jTEFKj1hHBSjsFiVAHZ66xnyoFGVzGLZ3glWZ15wQAfLby71VB+9p4KJFpyBePAE=\n",
  "duration": 60,
  "valid_before": "2019-07-24T23:40:45Z"
}
```

The certificate can be extracted by piping the command to  `jq .certificate > cert.pem` then connect with `-i cert.pem`.

### Overwrite source addresses

Use `--source-address` with a CIDR to restrict the source addresses to a specific address. The option can be repeated to
have more than one source addresses. This is useful for jumping through a bastion with an external IP for the
bastion and an internal one for the internal servers.

The CIDR must be a subset of the source addresses set in the environment configuration in the Lambda.

### Overwrite the session duration

Use `--duration` with the duration to use. The value must be shorter than the default duration in the environment configuration in the Lambda.


### Using ProxyCommand

To request the certificate as you connect the command can be used as an SSH ProxyCommand

```
$ ssh -o ProxyCommand="iam-request-ssh-key-signature --environment prod --lambda-arn arn:aws:lambda:eu-west-1:330428913683:function:sign-ssh-key --proxy-config %h:%p" hamstah@server.com
```

Or add the `ProxyCommand` to `~/.ssh/config`

```
Host server.com
     ProxyCommand iam-request-ssh-key-signature --environment prod --lambda-arn arn:aws:lambda:eu-west-1:330428913683:function:sign-ssh-key --proxy-config %h:%p
```


## View the certificates

By default the certificates are stored in the ssh-agent, you can view them with `ssh-add -L | ssh-keygen -L -f -`. Note that the
certificates are automatically removed from the ssh agent when expired.
