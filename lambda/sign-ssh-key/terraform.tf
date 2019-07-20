// Create a secretsmanager secret to store the CA information
locals {
  // do not fill these they are just used to initialise the secret structure
  ca = {
    private_key            = ""
    private_key_passphrase = ""
    public_key             = ""
  }
}

resource "aws_secretsmanager_secret" "ca" {
  name = "sign-ssh-key-ca"
}

resource "aws_secretsmanager_secret_version" "ca" {
  secret_id     = aws_secretsmanager_secret.ca.id
  secret_string = jsonencode(local.ca)

  lifecycle {
    ignore_changes = [secret_string]
  }
}

// Create an execution role for the lambda
data "aws_iam_policy_document" "lambda_assume" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "ssh_key_signer" {
  name               = "ssh-key-signer"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume.json
}

// Create a policy granting read access to the CA secret
data "aws_iam_policy_document" "read_ca_secret" {
  statement {
    actions   = ["secretsmanager:GetSecretValue"]
    resources = [aws_secretsmanager_secret.ca.id]
  }
}

resource "aws_iam_policy" "read_ca_secret" {
  name   = "read-sign-ssh-key-ca"
  policy = data.aws_iam_policy_document.read_ca_secret.json
}

// Attach the policy to the lambda execution role
resource "aws_iam_role_policy_attachment" "read_ca_secret" {
  role       = aws_iam_role.ssh_key_signer.name
  policy_arn = aws_iam_policy.read_ca_secret.arn
}

// Build the lambda binary
resource "null_resource" "build" {

  triggers = {
    build  = sha1(file("main.go"))
    config = sha1(file("prod.json"))
  }

  provisioner "local-exec" {
    command = "GOOS=linux GOARCH=amd64 go build -o lambda-sign-ssh-key && chmod +x lambda-sign-ssh-key && zip lambda lambda-sign-ssh-key prod.json"
  }
}

resource "aws_lambda_function" "lambda" {
  filename         = "${path.module}/lambda.zip"
  function_name    = "sign-ssh-key"
  role             = aws_iam_role.ssh_key_signer.arn
  handler          = "lambda-sign-ssh-key"
  source_code_hash = base64sha256("${path.module}/lambda.zip")
  runtime          = "go1.x"
  depends_on       = [null_resource.build]
}

output "lambda_arn" {
  value = aws_lambda_function.lambda.arn
}

// Create a role to be assumed by users requesting signatures
variable "ssh_signature_requester_account_ids" {
  type    = list(string)
  default = []
}

data "aws_iam_policy_document" "requester_assume" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "AWS"
      identifiers = var.ssh_signature_requester_account_ids
    }
  }
}

resource "aws_iam_role" "ssh_key_signature_requester" {
  name               = "ssh-key-signature-requester"
  assume_role_policy = data.aws_iam_policy_document.requester_assume.json
}

output "ssh_key_signature_requester_arn" {
  value = aws_iam_role.ssh_key_signature_requester.arn
}

// Create a policy to allow the role to call the lambda
data "aws_iam_policy_document" "ssh_key_signature_request" {
  statement {
    actions = ["lambda:InvokeFunction"]
    resources = [
      aws_lambda_function.lambda.arn
    ]
  }
}

resource "aws_iam_policy" "ssh_key_signature_request" {
  name   = "ssh-key-signature-request"
  policy = data.aws_iam_policy_document.ssh_key_signature_request.json
}

// Attach the policy to the requester role
resource "aws_iam_role_policy_attachment" "ssh_key_signature_request" {
  role       = aws_iam_role.ssh_key_signature_requester.name
  policy_arn = aws_iam_policy.ssh_key_signature_request.arn
}
