# how to connect via serial ssh:

last resort if network is down

```
ssh -o ControlPath=none -o IdentitiesOnly=yes -o CheckHostIP=no -o UserKnownHostsFile=./serialssh-knownhosts -p 9600 -i ~/.ssh/id_rsa epdo9egut
unl4o6bfsk8.yc-serialssh@serialssh.cloud.yandex.net
```

This also works:
```
ssh -p 9600 epdo9egutunl4o6bfsk8.yc-serialssh@serialssh.cloud.yandex.net
```
