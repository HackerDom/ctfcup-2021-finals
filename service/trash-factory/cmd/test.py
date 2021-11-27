#!/usr/bin/env python3
from socket import *

TOKEN_KEY = "9360828cf3f9772c"
TOKEN_KEY_BYTES = int(TOKEN_KEY, base=16).to_bytes(8, byteorder="big")

TOKEN = "7b9a84434d6d2516"
TOKEN_BYTES = int(TOKEN, base=16).to_bytes(8, byteorder="big")

ADDR = ("127.0.0.1", 9090)
MAGIC = b"\x03\x13\x37"


def get_payload(command:bytes):
    return MAGIC + command

def encrypt(payload:bytes):
    res = bytearray()
    for i in range(len(payload)):
        res.append(payload[i] ^ TOKEN_BYTES[i % len(TOKEN_BYTES)])
    return bytes(res)

def get_message(command:bytes):
    return TOKEN_KEY_BYTES + encrypt(get_payload(command))

def create_container(s:socket):
    command = b"\x09\x03hello"
    msg = get_message(command)
    s.send(msg)
    return s.recv(1)

def main():
    s = socket()
    s.connect(ADDR)
    print(f"Hello bytes from server: {s.recv(3)}")
    print(f"STAUS CODE IS {create_container(s)}")
    s.close()

if __name__ == "__main__":
    main()
