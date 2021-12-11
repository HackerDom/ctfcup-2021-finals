#!/usr/bin/env python3.9

import multiprocessing as mp
import threading as th
import os

def do():
  if os.system('./checker.py check localhost:4040 1>/dev/null 2>/dev/null') != 0x6500:
    print('pref test failed')

def fef(x):
  print(f'\b\b\b\b\b\b\b\b\b{x // 1000 * 100}%')
  threads = []
  for i in range(3):
    t = th.Thread(target=do, args=())
    t.start()
    threads.append(t)
  for t in threads:
    t.join()

with mp.Pool() as mp:
  mp.map(fef, [i for i in range(1000)])

