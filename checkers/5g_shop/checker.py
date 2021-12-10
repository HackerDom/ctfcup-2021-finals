#!/usr/bin/env python3.9
import base64

import ctypes
import hashlib
import logging
import os
import random
import string
import uuid
from datetime import datetime
from http.client import HTTPConnection
from traceback import print_exc

import requests
from dateutil import parser
from gornilo import \
    GetRequest, \
    CheckRequest, \
    PutRequest, \
    Checker, \
    Verdict
from requests.adapters import HTTPAdapter
from requests.exceptions import Timeout
from requests.packages.urllib3.util.retry import Retry

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


file_dir = os.path.dirname(os.path.realpath(__file__))
calc = ctypes.CDLL(f'{file_dir}/libcalc.so')
calc.GetCashback.restype = ctypes.c_int
calc.GetCashback.argtypes = [ctypes.POINTER(ctypes.c_char)]

calc.GetServiceFee.restype = ctypes.c_int
calc.GetServiceFee.argtypes = [ctypes.c_int, ctypes.POINTER(ctypes.c_char), ctypes.POINTER(ctypes.c_char)]

API_PORT = 4040
AUTH_COOKIE_NAME = '5GAuth'

mumble = Verdict.MUMBLE('wrong server response')
down = Verdict.DOWN("service not responding")


def nicehost(h):
    if h.find(':') != -1:
        return h

    return h + ':' + str(API_PORT)


def down_status_code(response):
    return response.status_code == 502 or response.status_code == 503


def select_random_image():
    name = random.choice(
        [
            'laptop1.jpg',
            'laptop2.jpg',
            'laptop3.jpg',
            'laptop4.jpeg',
            'phon1.jpg',
            'phon2.jpeg',
            'phon3.jpeg',
            'phon4.jpeg'
        ])
    return file_dir + "/" + name, name


@checker.define_put(vuln_num=1, vuln_rate=1)
def put(put_request: PutRequest) -> Verdict:
    try:
        login = str(uuid.uuid4())
        password = str(uuid.uuid4())
        password_hash = get_sha256(password)

        log.debug(f'creating user {login} with password {password} ({password_hash})')

        response = get_session_with_retry().post(
            f'http://{nicehost(put_request.hostname)}/api/users',
            headers={'User-Agent': get_random_user_agent()},
            json={
                'login': login,
                'password_hash': password_hash,
                'credit_card_info': put_request.flag
            }
        )

        log.debug(f'response: \ncode: {response.status_code}\nheaders:{response.headers}\ntext:{response.text}')

        if down_status_code(response):
            return down

        if response.status_code != 201:
            log.error(f'unexpected code - {response.status_code}')
            return mumble

        auth_token = response.cookies[AUTH_COOKIE_NAME]
        log.debug(f'auth cookie will be {auth_token}')

        return Verdict.OK(auth_token)
    except Timeout as e:
        log.error(f'{e}, {print_exc()}')
        return down
    except Exception as e:
        log.error(f'{e}, {print_exc()}')
        return down


@checker.define_get(vuln_num=1)
def get(get_request: GetRequest) -> Verdict:
    try:
        response = get_session_with_retry().get(
            f'http://{nicehost(get_request.hostname)}/api/users',
            headers={"User-Agent": get_random_user_agent()},
            cookies={AUTH_COOKIE_NAME: get_request.flag_id},
            timeout=3
        )

        log.debug(f'response: \ncode: {response.status_code}\nheaders:{response.headers}\ntext:{response.text}')

        if down_status_code(response):
            return down

        if response.status_code != 200 or response.json()["credit_card_info"] != get_request.flag:
            return Verdict.CORRUPT('wrong flag')

        response = get_session_with_retry().get(
            f'http://{nicehost(get_request.hostname)}/api/users/{response.json()["id"]}',
            headers={"User-Agent": get_random_user_agent()},
            cookies={AUTH_COOKIE_NAME: get_request.flag_id},
            timeout=3
        )

        if down_status_code(response):
            return down

        if response.status_code != 200:
            return mumble

        log.debug(f'response: \ncode: {response.status_code}\nheaders:{response.headers}\ntext:{response.text}')

        if response.json()["credit_card_info"] != get_request.flag:
            return Verdict.CORRUPT('wrong flag')

        return Verdict.OK()
    except Timeout as e:
        log.error(f'{e}, {print_exc()}')
        return down
    except Exception as e:
        log.error(f'{e}, {print_exc()}')
        return Verdict.CORRUPT("service can't give a flag")


@checker.define_check
def check(check_request: CheckRequest) -> Verdict:
    try:
        login1 = str(uuid.uuid4())
        password1 = str(uuid.uuid4())
        password_hash1 = get_sha256(password1)
        flag1 = generate_some_flag()

        login2 = str(uuid.uuid4())
        password2 = str(uuid.uuid4())
        password_hash2 = get_sha256(password2)
        flag2 = generate_some_flag()

        host = nicehost(check_request.hostname)

        uid1, auth1, verdict = check_create_user(host, login1, password1, password_hash1, flag1)
        if verdict is not None:
            return verdict

        uid2, auth2, verdict = check_create_user(host, login2, password2, password_hash2, flag2)
        if verdict is not None:
            return verdict

        verdict = check_auth(host, login1, password_hash1, auth1)
        if verdict is not None:
            return verdict

        verdict = check_auth(host, login2, password_hash2, auth2)
        if verdict is not None:
            return verdict

        verdict = check_users_list(host, auth1, auth2)
        if verdict is not None:
            return verdict

        image1_id, verdict = check_upload_image(host, auth1)
        if verdict is not None:
            return verdict

        image2_id, verdict = check_upload_image(host, auth2)
        if verdict is not None:
            return verdict

        ware_id1, verdict = check_create_ware_and_my_wares_list(host, auth1, uid1, image1_id)
        if verdict is not None:
            return verdict

        ware_id2, verdict = check_create_ware_and_my_wares_list(host, auth2, uid2, image2_id)
        if verdict is not None:
            return verdict

        verdict = check_make_purchase(host, auth1, ware_id2)
        if verdict is not None:
            return verdict

        verdict = check_make_purchase(host, auth2, ware_id1)
        if verdict is not None:
            return verdict

        log.debug("OK")

        return Verdict.OK()
    except Timeout as e:
        log.error(f'{e}, {print_exc()}')
        return down
    except Exception as e:
        log.error(f'{e}, {print_exc()}')
        return down


def check_upload_image(host, auth):
    path, name = select_random_image()
    with open(path, 'rb') as f:
        data = f.read()
    response = get_session_with_retry().post(
        f'http://{host}/api/images',
        headers={'User-Agent': get_random_user_agent()},
        cookies={AUTH_COOKIE_NAME: auth},
        files={name: data}
    )
    log.debug(f'response: \ncode: {response.status_code}\nheaders:{response.headers}\ntext:{response.text}')
    if down_status_code(response):
        return None, down

    if response.status_code != 201:
        return None, mumble

    j = response.json()
    id_, path = j["id"], j["path"]
    if path is None or id_ is None:
        return None, mumble

    response = get_session_with_retry().get(
        f'http://{host}{path}'
    )
    if down_status_code(response):
        return None, down

    if response.status_code != 200:
        return None, mumble

    content = response.content
    if len(data) != len(content):
        return None, mumble

    for i in range(len(data)):
        if data[i] != content[i]:
            log.error('returning files are not equals')
            return None, mumble

    response = get_session_with_retry().get(
        f'http://{host}/api/images/{id_}',
        headers={'User-Agent': get_random_user_agent()},
        cookies={AUTH_COOKIE_NAME: auth}
    )
    log.debug(f'response: \ncode: {response.status_code}\nheaders:{response.headers}\ntext:{response.text}')
    if down_status_code(response):
        return None, down
    if response.status_code != 200:
        return None, mumble

    if str(response.json()["id"]) != str(id_) or response.json()["path"] != path:
        return None, mumble

    response = get_session_with_retry().get(
        f'http://{host}/api/images/my',
        headers={'User-Agent': get_random_user_agent()},
        cookies={AUTH_COOKIE_NAME: auth}
    )
    log.debug(f'response: \ncode: {response.status_code}\nheaders:{response.headers}\ntext:{response.text}')
    if down_status_code(response):
        return None, down
    if response.status_code != 200:
        return None, mumble

    imgs = response.json()["images"]
    if imgs is None or len(imgs) != 1:
        return None, mumble

    if str(imgs[0]["id"]) != str(id_) or str(imgs[0]["path"]) != path:
        return None, mumble

    return id_, None


def check_make_purchase(host, auth, ware_id):
    log.debug("check making purchases")

    purchase_ids = set()

    for i in range(5):
        response = get_session_with_retry().post(
            f'http://{host}/api/purchases?ware_id={ware_id}',
            headers={'User-Agent': get_random_user_agent()},
            cookies={AUTH_COOKIE_NAME: auth}
        )

        if down_status_code(response):
            return down

        if response.status_code != 201:
            log.error("invalid status")
            return mumble

        pid = response.json()["id"]
        if pid is None:
            return mumble

        purchase_ids.add(pid)

    if len(purchase_ids) != 5:
        log.error('invalid purchases size')
        return mumble

    response = get_session_with_retry().get(
        f'http://{host}/api/purchases/my',
        headers={'User-Agent': get_random_user_agent()},
        cookies={AUTH_COOKIE_NAME: auth}
    )
    log.debug(f'response: \ncode: {response.status_code}\nheaders:{response.headers}\ntext:{response.text}')

    if down_status_code(response):
        return down

    if response.status_code != 200:
        return mumble

    log.debug(f"created ids are {purchase_ids}")

    j = response.json()
    for p in j["purchases"]:
        if str(p["id"]) not in purchase_ids:
            log.error("given id not found")
            return mumble
        purchase_ids.remove(str(p["id"]))
        if str(p["ware_id"]) != ware_id:
            log.error('invalid ware id')
            return mumble

    return None


def check_create_ware_and_my_wares_list(host, auth, uid, iid):
    log.debug(f'check create ware for {auth} {uid}')

    title = str(uuid.uuid4())
    description = str(uuid.uuid4())
    price = random.randint(500, 100500 * 2)
    service_fee = str(calc.GetServiceFee(price, title.encode('utf-8'), description.encode('utf-8')))

    response = get_session_with_retry().post(
        f'http://{host}/api/wares',
        headers={'User-Agent': get_random_user_agent()},
        cookies={AUTH_COOKIE_NAME: auth},
        json={
            "title": title,
            "description": description,
            "price": price,
            "image_id": iid
        },
        timeout=3
    )
    log.debug(f'response: \ncode: {response.status_code}\nheaders:{response.headers}\ntext:{response.text}')

    if down_status_code(response):
        return None, down

    if response.status_code != 201:
        return None, mumble

    ware_id = response.json()["id"]
    if ware_id is None:
        return None, mumble

    response = get_session_with_retry().get(
        f'http://{host}/api/wares/my',
        headers={'User-Agent': get_random_user_agent()},
        cookies={AUTH_COOKIE_NAME: auth},
        timeout=3
    )
    log.debug(f'response: \ncode: {response.status_code}\nheaders:{response.headers}\ntext:{response.text}')

    if down_status_code(response):
        return None, down

    if response.status_code != 200 or response.json()["ids"] is None:
        return None, mumble

    ids = response.json()["ids"]

    if len(ids) != 1 or str(ids[0]) != str(ware_id):
        return None, mumble

    response = get_session_with_retry().get(
        f'http://{host}/api/wares/{ids[0]}',
        headers={'User-Agent': get_random_user_agent()},
        cookies={AUTH_COOKIE_NAME: auth},
        timeout=3
    )
    log.debug(f'response: \ncode: {response.status_code}\nheaders:{response.headers}\ntext:{response.text}')

    if down_status_code(response):
        return None, down

    if response.status_code != 200:
        return None, mumble

    j = response.json()
    if str(j["id"]) != str(ware_id) or str(j["seller_id"]) != str(uid) or j["title"] != title or j[
        "description"] != description \
            or str(j["price"]) != str(price) or str(j["service_fee"]) != str(service_fee):
        log.error(f'{j["id"]} {j["seller_id"]} {j["title"]} {j["description"]} {j["price"]} {j["service_fee"]}')
        log.error(f'{ware_id} {uid} {title} {description} {price} {service_fee}')
        return None, mumble

    return ware_id, None


def check_users_list(host, auth1, auth2):
    log.debug('check login')

    s = get_session_with_retry()

    response = s.get(
        f'http://{host}/api/users',
        headers={"User-Agent": get_random_user_agent()},
        cookies={AUTH_COOKIE_NAME: auth1},
        timeout=3
    )

    log.debug(f'response: \ncode: {response.status_code}\nheaders:{response.headers}\ntext:{response.text}')

    if down_status_code(response):
        return down

    if response.status_code != 200 or response.json()["id"] is None:
        return mumble

    id1 = response.json()["id"]

    response = s.get(
        f'http://{host}/api/users',
        headers={"User-Agent": get_random_user_agent()},
        cookies={AUTH_COOKIE_NAME: auth2},
        timeout=3
    )

    log.debug(f'response: \ncode: {response.status_code}\nheaders:{response.headers}\ntext:{response.text}')

    if down_status_code(response):
        return down

    if response.status_code != 200 or response.json()["id"] is None:
        return mumble

    id2 = response.json()["id"]

    id1found = False
    id2found = False
    page = 0

    while True:
        response = s.get(
            f'http://{host}/api/users/list?page_size=100&page_num={page}',
            headers={"User-Agent": get_random_user_agent()},
            cookies={AUTH_COOKIE_NAME: auth1},
            timeout=3
        )

        if down_status_code(response):
            return down

        if response.status_code != 200 or response.json()["users"] is None:
            return mumble

        if len(response.json()["users"]) > 100:
            return mumble

        if len(response.json()["users"]) == 0:
            break

        for u in response.json()["users"]:
            if u['id'] == id1:
                id1found = True
            if u['id'] == id2:
                id2found = True

        page += 1

    if not id1found or not id2found:
        return mumble

    return None


def check_auth(host, login, phash, auth):
    log.debug(f'checking auth {login} with password_hash {phash}')

    response = get_session_with_retry().put(
        f'http://{host}/api/users/auth',
        headers={'User-Agent': get_random_user_agent()},
        json={
            'login': login,
            'password_hash': phash
        }
    )

    if down_status_code(response):
        return down

    if response.status_code != 200 or response.cookies[AUTH_COOKIE_NAME] != auth:
        return mumble

    return None


def check_create_user(host, login, password, password_hash, flag):
    log.debug(f'creating user {login} with password {password} ({password_hash})')

    response = get_session_with_retry().post(
        f'http://{host}/api/users',
        headers={'User-Agent': get_random_user_agent()},
        json={
            'login': login,
            'password_hash': password_hash,
            'credit_card_info': flag
        }
    )

    log.debug(f'response: \ncode: {response.status_code}\nheaders:{response.headers}\ntext:{response.text}')

    if down_status_code(response):
        return None, None, down

    user_id = response.json()["id"]
    if response.status_code != 201 or user_id is None:
        log.error(f'unexpected code - {response.status_code}')
        return None, None, mumble

    auth_token = response.cookies[AUTH_COOKIE_NAME]
    log.debug(f'auth cookie will be {auth_token}')

    if auth_token != response.json()["auth_cookie"]:
        return None, None, mumble

    response = get_session_with_retry().get(
        f'http://{host}/api/users',
        headers={'User-Agent': get_random_user_agent()},
        cookies={AUTH_COOKIE_NAME: auth_token}
    )

    if down_status_code(response):
        return None, None, down

    if response.status_code != 200:
        return None, None, mumble

    j = response.json()

    log.debug(f'response: \ncode: {response.status_code}\nheaders:{response.headers}\ntext:{response.text}')

    if j["login"] != login:
        log.debug('invalid login')
        return None, None, mumble

    # TODO: may be delete this checking ?
    if (datetime.utcnow() - parser.parse(j["created_at"])).seconds >= 20:
        log.debug('invalid created at')
        return None, None, mumble

    if str(j["cashback"]) != str(calc.GetCashback(login.encode('utf-8'))):
        log.debug(f'invalid cashback {j["cashback"]} vs {str(calc.GetCashback(login.encode("utf-8")))}')
        return None, None, mumble

    if j["credit_card_info"] != flag:
        log.debug('invalid flag')
        return None, None, mumble

    return user_id, auth_token, None


if __name__ == "__main__":
    checker.run()
