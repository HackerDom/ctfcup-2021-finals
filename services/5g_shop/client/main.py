import sys
import os
import curses

from client import Client


def main(stdscr, host):
    client = Client(stdscr, host)

    client.run()


if __name__ == '__main__':
    if len(sys.argv) < 2:
        print('please, provide one argument - host of 5g shop')
        exit()

    os.environ.setdefault('ESCDELAY', '25')
    curses.wrapper(main, sys.argv[1])
