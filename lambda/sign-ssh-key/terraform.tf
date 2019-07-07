resource "null_resource" "build" {

  triggers = {
    build = sha1(file("main.go"))
  }

  provisioner "local-exec" {
    command = "GOOS=linux GOARCH=amd64 go build -o lambda-sign-ssh-key && chmod +x lambda-sign-ssh-key && zip lambda lambda-sign-ssh-key prod.json"
  }
}


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

resource "aws_secretsmanager_secret" "ca" {
  name = "sign-ssh-key-ca"
}

locals {
  ca = {
    private_key            = ""
    private_key_passphrase = ""
    public_key             = ""
  }
}

resource "aws_secretsmanager_secret_version" "ca" {
  secret_id     = aws_secretsmanager_secret.ca.id
  secret_string = jsonencode(local.ca)

  lifecycle {
    ignore_changes = [secret_string]
  }
}

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

resource "aws_iam_role_policy_attachment" "read_ca_secret" {
  role       = aws_iam_role.ssh_key_signer.name
  policy_arn = aws_iam_policy.read_ca_secret.arn
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
