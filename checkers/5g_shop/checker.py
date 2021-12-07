#!/usr/bin/env python3
import base64

import requests
import socket
import random
import time
import uuid
import string
import hashlib
import json
from traceback import print_exc
import logging
import contextlib
import ctypes

from http.client import HTTPConnection
from requests.exceptions import Timeout
from requests.adapters import HTTPAdapter
from requests.packages.urllib3.util.retry import Retry
from gornilo import \
    GetRequest, \
    CheckRequest, \
    PutRequest, \
    Checker, \
    Verdict

from user_agent_randomizer import get_random_user_agent

HTTPConnection.debuglevel = 1
logging.basicConfig(
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logging.getLogger().setLevel(logging.DEBUG)
log = logging.getLogger("requests.packages.urllib3")
log.setLevel(logging.DEBUG)
log.propagate = True

checker = Checker()


def get_sha256(s):
    return hashlib.sha256(s.encode('utf-8')).hexdigest()


flag_alphabet = string.digits + string.ascii_uppercase


def generate_some_flag():
    # [0-9A-Z]{31}=
    flag = []
    for i in range(31):
        flag.append(random.choice(flag_alphabet))
    flag.append('=')

    return ''.join(flag)


def get_session_with_retry(
        retries=3,
        backoff_factor=0.3,
        status_forcelist=(400, 404, 500, 502),
        session=None,
):
    session = session or requests.Session()
    retry = Retry(
        total=retries,
        read=retries,
        connect=retries,
        backoff_factor=backoff_factor,
        status_forcelist=status_forcelist,
    )
    adapter = HTTPAdapter(max_retries=retry)
    session.mount('http://', adapter)

    return session


calc = ctypes.CDLL('./libcalc.so')
calc.GetCashback.restype = ctypes.c_int
calc.GetCashback.argtypes = [ctypes.POINTER(ctypes.c_char)]

calc.GetServiceFee.restype = ctypes.c_int
calc.GetServiceFee.argtypes = [ctypes.c_int, ctypes.POINTER(ctypes.c_char), ctypes.POINTER(ctypes.c_char)]

API_PORT = 4040
AUTH_COOKIE_NAME = '5GAuth'

mumble = Verdict.MUMBLE('wrong server response')


@checker.define_put(vuln_num=1, vuln_rate=1)
def put(put_request: PutRequest) -> Verdict:
    try:
        login = str(uuid.uuid4())
        password = str(uuid.uuid4())
        password_hash = get_sha256(password)

        log.debug(f'creating user {login} with password {password} ({password_hash})')

        response = get_session_with_retry().post(
            f'http://{put_request.hostname}:{API_PORT}/api/users',
            headers={'User-Agent': get_random_user_agent()},
            json={
                'login': login,
                'password_hash': password_hash,
                'credit_card_info': put_request.flag
            }
        )

        log.debug(f'response: \ncode: {response.status_code}\nheaders:{response.headers}\ntext:{response.text}')

        if response.status_code != 201:
            log.error(f'unexpected code - {response.status_code}')
            return mumble

        auth_token = response.cookies[AUTH_COOKIE_NAME]
        log.debug(f'auth cookie will be {auth_token}')

        return Verdict.OK(auth_token)
    except Timeout as e:
        log.error(f'{e}, {print_exc()}')
        return Verdict.DOWN("service not responding")
    except Exception as e:
        log.error(f'{e}, {print_exc()}')
        return Verdict.DOWN('cant create user')


@checker.define_get(vuln_num=1)
def get(get_request: GetRequest) -> Verdict:
    try:
        response = get_session_with_retry().get(
            f'http://{get_request.hostname}:{API_PORT}/api/users',
            headers={"User-Agent": get_random_user_agent()},
            cookies={AUTH_COOKIE_NAME: get_request.flag_id},
            timeout=3
        )

        log.debug(f'response: \ncode: {response.status_code}\nheaders:{response.headers}\ntext:{response.text}')

        if response.json()["credit_card_info"] != get_request.flag:
            return Verdict.CORRUPT('wrong flag')

        response = get_session_with_retry().get(
            f'http://{get_request.hostname}:{API_PORT}/api/users/{response.json()["id"]}',
            headers={"User-Agent": get_random_user_agent()},
            cookies={AUTH_COOKIE_NAME: get_request.flag_id},
            timeout=3
        )

        log.debug(f'response: \ncode: {response.status_code}\nheaders:{response.headers}\ntext:{response.text}')

        if response.json()["credit_card_info"] != get_request.flag:
            return Verdict.CORRUPT('wrong flag')

        return Verdict.OK()
    except Timeout as e:
        log.error(f'{e}, {print_exc()}')
        return Verdict.DOWN("service not responding")
    except Exception as e:
        log.error(f'{e}, {print_exc()}')
        return Verdict.CORRUPT("service can't give a flag")


@checker.define_check
def check(check_request: CheckRequest) -> Verdict:
    try:
        

        return Verdict.OK()
    except Timeout as e:
        log.error(f'{e}, {print_exc()}')
        return Verdict.DOWN("service not responding")
    except Exception as e:
        log.error(f'{e}, {print_exc()}')
        return Verdict.DOWN('cant create user')


if __name__ == "__main__":
    checker.run()
