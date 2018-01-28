#!/usr/bin/env python3
import github3
import getpass
import os
import mimetypes

base = os.path.join(os.path.dirname(__file__), "..")
version = open(os.path.join(base, "VERSION")).read().strip()

last_2fa = None

def my_two_factor_function():
    global last_2fa
    code = ''
    while not code:
        code = input('Enter 2FA code: [%s] ' % last_2fa) or last_2fa
    last_2fa = code
    return code


mimetypes.init()

current_user = os.getenv("USER")
username = input('Username: [%s] ' % current_user) or current_user
password = getpass.getpass('Password: ')

client = github3.login(username, password, two_factor_callback=my_two_factor_function)

repo = client.repository(username, 'awstools')

release = repo.release_from_tag('v%s' % version)
if release is None:
    release = repo.create_release('v%s' % version, draft=False, prerelease=False)

bin_dir = os.path.join(base, "bin")
for file in os.listdir(bin_dir):
    rel_file = os.path.join(bin_dir, file)
    content_type, _ = mimetypes.guess_type(rel_file)
    if content_type is None:
        content_type = "application/octet-stream"

    print(rel_file, content_type)
    try:
        asset = release.upload_asset(
            content_type=content_type,
            name=file,
            asset=open(rel_file, 'rb').read(),
        )
    except Exception as e:
        print(e)

    print("\n\n\n")
