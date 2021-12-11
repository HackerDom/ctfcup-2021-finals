from os import system as r
import os
import re

PASSWORD='r86ue5laz88f9p4asn'

teams = [t for t in os.listdir('.') if t.isdigit()]
teams.sort()

r('rm -r archives')
os.makedirs('archives')


with open('../ansible/group_vars/cs') as f:
    lines = f.read()
    tokens = re.findall("token => '([0-9a-z]+)'", lines)
    names = re.findall("name => '([^']+)'", lines)
    print(tokens)
    print(names)


for i,t in enumerate(teams):
    with open('{}/checksystem_token'.format(t), 'w') as f:
        f.write(tokens[i] + "\n")

    n = names[i].replace(' ', '_')
    r('7z a -p{p} "archives/{t}-{n}.7z" {t}/main_ssh_host {t}/ssh_key {t}/ssh_key.pub {t}/client.ovpn {t}/checksystem_token'.format(
        t=t, p=PASSWORD, n=n))
