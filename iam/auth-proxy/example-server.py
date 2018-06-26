from flask import Flask, session, request, Response, abort, redirect
import boto3
import base64
import json
import os

app = Flask(__name__)
app.secret_key = os.environ["SECRET_KEY"]
kms = boto3.client('kms')
kms_key_id = kms.describe_key(KeyId=os.getenv("KMS_KEY_ID"))["KeyMetadata"]["Arn"]

kms_encryption_context = {"salt": os.environ["KMS_SALT"]}
kms_encryption_context_header = base64.b64encode(json.dumps(kms_encryption_context))

sts = boto3.client('sts')
account_id = os.getenv("AWS_ACCOUNT_ID")
if account_id is None:
    account_id = sts.get_caller_identity()["Account"]
auth_headers = {
    "Iam-Auth-Kms-Encryption-Context": kms_encryption_context_header,
    "Iam-Auth-Kms-Key-Id": kms_key_id,
    "WWW-Authenticate": 'IAM realm="%s"' % account_id,
}

@app.route("/auth")
def auth_iam():
    if "token" not in request.args:
        abort(403)

    return_url = request.args.get("return_url", "/")
    if not return_url.startswith("/"):
        abort(403)

    try:
        token = request.args["token"]
        blob = base64.b64decode(token)
        res = kms.decrypt(
            CiphertextBlob=blob,
            EncryptionContext=kms_encryption_context,
        )
        creds = json.loads(res["Plaintext"])
        sts = boto3.client(
            'sts',
            aws_access_key_id=creds["AccessKeyId"],
            aws_secret_access_key=creds["SecretAccessKey"],
            aws_session_token=creds["SessionToken"],
        )
        identity = sts.get_caller_identity()['Arn']
        if identity.split(":")[4] != account_id:
            abort(403)

        session["user_id"] = identity
        return redirect(return_url)
    except:
        abort(403)

    abort(400)

@app.route("/check")
def check():
    if "user_id" in session:
        return session["user_id"]
    else:
        return ":("


@app.route("/")
def home():
    if session.get("user_id") is None:
        return Response("Please log in", 401, auth_headers)

    return session["user_id"]

if __name__ == "__main__":
    app.run(debug=True, host="0.0.0.0")
