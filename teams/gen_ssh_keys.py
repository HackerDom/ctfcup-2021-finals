#!/usr/bin/python
from os import system as r
import random
import string

r("ssh-keygen -t rsa -N '' -f for_devs.ssh_key")

for i in range(101, 111):
    r('mkdir -p {}'.format(i))
    r("ssh-keygen -t rsa -N '' -f {}/ssh_key".format(i))

