from os import system as r

for i in range(1, 11):
    r('wget -O {} "https://cataas.com/cat?type=sq"'.format(100+i))
