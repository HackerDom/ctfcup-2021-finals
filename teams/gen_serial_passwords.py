#!/usr/bin/python
from os import system as r
import random
import string


for i in range(101, 105):
    p = ''.join(random.sample(string.ascii_letters + string.digits, 14))
    r('echo "{}" > {}/serial_passwd'.format(p, i))
    r('echo "{}" | mkpasswd --method=SHA-512 --rounds=4096 --stdin > {}/serial_passwd_hash'.format(p,i))

