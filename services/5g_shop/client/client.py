import curses
import hashlib
import os.path
from curses.textpad import rectangle
from datetime import datetime

import requests

AUTH_COOKIE_NAME = '5GAuth'


def _valid_string(s):
    for c in str(s.decode('utf-8')):
        if c == '_' or c == '-':
            continue
        if not c.isalnum():
            return False

    return True


def _valid_num(s):
    for c in str(s.decode('utf-8')):
        if not c.isdigit():
            return False

    return True


def get_sha256(s):
    return hashlib.sha256(s).hexdigest()


class Client:
    def __init__(self, stdscr, host):
        self.__host = host
        self.__stdscr = stdscr
        self.__login = None, None, None

    def __print_ware(self, ware, q, yorno, title):
        with requests.get(f'http://{self.__host}/api/images/{ware["image_id"]}',
                          cookies={AUTH_COOKIE_NAME: self.__login[1]}) as r:
            if r.status_code != 200:
                raise ValueError(f'invalid server status code {r.status_code}')

            image_path = f'http://{self.__host}{r.json()["path"]}'

        while True:
            self.__draw_windows(title)

            max_y, max_x = self.__stdscr.getmaxyx()
            cy, cx = max_y // 2, max_x // 2

            self.__stdscr.addstr(cy - 2, cx - 15, 'Description: ' + ware['description'])
            self.__stdscr.addstr(cy, cx - 15, 'Price: ' + str(ware['price']))
            self.__stdscr.addstr(cy + 2, cx - 15, 'Image: ' + image_path)
            if ware["service_fee"] is not None:
                self.__stdscr.addstr(cy + 4, cx - 15, 'Service fee: ' + str(ware["service_fee"]))
            m = q
            self.__stdscr.addstr(cy + 8, cx - len(m) // 2, m, curses.A_BOLD | curses.A_BLINK)

            self.__refresh_windows()

            ch = self.__stdscr.getch()

            if yorno:
                if ch == 27 or ch == ord('n'):
                    return None
                elif ch == ord('y'):
                    break
                else:
                    continue
            else:
                return None

    def __ware_buy(self, ware):
        if self.__print_ware(ware, '[y/n]?', True, f'Do you want to buy {ware["title"]}?') is None:
            return

        max_y, max_x = self.__stdscr.getmaxyx()
        cy, cx = max_y // 2, max_x // 2

        with requests.post(f'http://{self.__host}/api/purchases?ware_id={ware["id"]}',
                           cookies={AUTH_COOKIE_NAME: self.__login[1]}) as r:
            if r.status_code != 201:
                raise ValueError(f'invalid server response code {r.status_code}')

            m = '!!! SUCCESS !!!'
            self.__stdscr.addstr(cy, cx - len(m) // 2, m, curses.A_BOLD | curses.A_BLINK)
            self.__refresh_windows()
            self.__stdscr.getch()

    def __select_ware_of(self, uid):
        with requests.get(f'http://{self.__host}/api/wares/of_user/{uid}',
                          cookies={AUTH_COOKIE_NAME: self.__login[1]}) as r:
            if r.status_code != 200:
                raise ValueError(f'invalid server response with code {r.status_code}')
            wares = []
            for ware_id in r.json()["ids"]:
                with requests.get(f'http://{self.__host}/api/wares/{ware_id}',
                                  cookies={AUTH_COOKIE_NAME: self.__login[1]}) as rr:
                    if rr.status_code != 200:
                        raise ValueError(f'invalid server response with code {r.status_code}')
                    wares.append(rr.json())
        if len(wares) == 0:
            return None

        ware_names = [w["title"] for w in wares]
        page = 0
        while True:
            p = ware_names[page * 5: page * 5 + 5]
            choices = ['previous page'] + p + ['next page']
            ch = self.__choose_option(f'Select ware. Page {page}', choices)

            if ch == 'previous page':
                if page > 0:
                    page -= 1
            elif ch == 'next page':
                if len(p) > 0:
                    page += 1
            elif ch is None:
                return None
            else:
                for u in wares:
                    if u["title"] == ch:
                        return u

    def __select_user(self):
        page = 0
        while True:
            with requests.get(
                    f'http://{self.__host}/api/users/list?page_num={page}&page_size=5',
                    cookies={AUTH_COOKIE_NAME: self.__login[1]}
            ) as r:
                if r.status_code != 200:
                    raise ValueError('invalid status code from server')
                users = r.json()['users']

            choices = ['previous page'] + [u["login"] for u in users] + ['next page']

            ch = self.__choose_option(f'Select user to find wares. Page {page}', choices)

            if ch == 'previous page':
                if page > 0:
                    page -= 1
            elif ch == 'next page':
                if len(users) > 0:
                    page += 1
            elif ch is None:
                return None, None
            else:
                for u in users:
                    if u["login"] == ch:
                        return u["id"], u["login"]

    def __buy(self):
        while True:
            uid, ulogin = self.__select_user()

            if uid is None or ulogin is None:
                return

            ware = self.__select_ware_of(uid)

            if ware is None:
                continue

            self.__ware_buy(ware)

    def __display_user_info(self):
        with requests.get(f'http://{self.__host}/api/users', cookies={AUTH_COOKIE_NAME: self.__login[1]}) as r:
            if r.status_code != 200:
                raise ValueError(f'invalid server response {r.status_code}')

            user = r.json()

        self.__draw_windows(f'This is you, {user["login"]}')

        max_y, max_x = self.__stdscr.getmaxyx()
        cy, cx = max_y // 2, max_x // 2

        self.__stdscr.addstr(cy - 2, cx - 15, 'Registration date: ' + user['created_at'])
        self.__stdscr.addstr(cy, cx - 15, 'Credit card info: ' + user['credit_card_info'])
        self.__stdscr.addstr(cy + 2, cx - 15, 'Cashback: ' + str(user['cashback']))
        m = 'Press any key to continue'
        self.__stdscr.addstr(cy + 6, cx - len(m) // 2, m, curses.A_BOLD | curses.A_BLINK)

        self.__refresh_windows()

        self.__stdscr.getch()

    def __my_purchases(self):
        with requests.get(f'http://{self.__host}/api/purchases/my', cookies={AUTH_COOKIE_NAME: self.__login[1]}) as r:
            if r.status_code != 200:
                raise ValueError(f'invalid server response {r.status_code}')

            wares = []

            for j in r.json()["purchases"]:
                with requests.get(f'http://{self.__host}/api/wares/{j["ware_id"]}',
                                  cookies={AUTH_COOKIE_NAME: self.__login[1]}) as r:
                    if r.status_code != 200:
                        raise ValueError(f'invalid server response {r.status_code}')
                    wares.append(r.json())

        ware_names = [w["title"] + " of price " + str(w["price"]) for w in wares]
        page = 0
        while True:
            p = ware_names[page * 5: page * 5 + 5]
            choices = ['previous page'] + p + ['next page']
            ch = self.__choose_option(f'This is list of wares you bought, page {page}', choices)

            if ch == 'previous page':
                if page > 0:
                    page -= 1
            elif ch == 'next page':
                if len(p) > 0:
                    page += 1
            elif ch is None:
                return

    def __my_wares(self):
        with requests.get(f'http://{self.__host}/api/users', cookies={AUTH_COOKIE_NAME: self.__login[1]}) as r:
            if r.status_code != 200:
                raise ValueError(f'invalid server response code {r.status_code}')
            myid = str(r.json()["id"])

        ware = self.__select_ware_of(myid)
        if ware is None:
            return

        self.__print_ware(ware, 'Press any key', False, f'Your ware {ware["title"]}')

    def __account_info(self):
        while True:
            ch = self.__choose_option('What kind of information do you want to watch?',
                                      ['user info', 'my purchases', 'my wares', 'back'])

            if ch is None or ch == 'back':
                return

            if ch == 'user info':
                self.__display_user_info()

            if ch == 'my purchases':
                self.__my_purchases()

            if ch == 'my wares':
                self.__my_wares()

    def __sell_some(self):
        while True:
            self.__draw_windows('Sell new ware')

            curses.echo()
            curses.curs_set(True)

            max_y, max_x = self.__stdscr.getmaxyx()
            cy, cx = max_y // 2, max_x // 2

            self.__stdscr.addstr(cy - 2, cx - 15, 'Title: ', curses.A_BOLD)
            self.__stdscr.addstr(cy + 2, cx - 15, 'Description: ', curses.A_BOLD)
            self.__stdscr.addstr(cy + 4, cx - 15, 'Price: ', curses.A_BOLD)
            self.__stdscr.addstr(cy + 6, cx - 15, 'Image path: ', curses.A_BOLD)

            self.__refresh_windows()

            title = self.__stdscr.getstr(cy - 2, cx - 15 + len('Title: '), 10)
            description = self.__stdscr.getstr(cy + 2, cx - 15 + len('Description: '), 20)
            price = self.__stdscr.getstr(cy + 4, cx - 15 + len('Price: '), 6)
            filename = (self.__stdscr.getstr(cy + 6, cx - 15 + len('Image path: '), 20)).decode('utf-8')

            if not _valid_string(title) or not _valid_string(description) or not _valid_num(price):
                m = '!!! PLEASE ENTER VALID DATA !!!'
                self.__stdscr.addstr(cy, cx - len(m) // 2, m, curses.A_BOLD | curses.A_BLINK)
                self.__refresh_windows()
                self.__stdscr.getch()
                continue

            if not os.path.exists(filename):
                m = '!!! PLEASE ENTER VALID FILENAME !!!'
                self.__stdscr.addstr(cy, cx - len(m) // 2, m, curses.A_BOLD | curses.A_BLINK)
                self.__refresh_windows()
                self.__stdscr.getch()
                continue

            with open(filename, 'rb') as f:
                data = f.read()
            with requests.post(f'http://{self.__host}/api/images',
                               cookies={AUTH_COOKIE_NAME: self.__login[1]},
                               files={filename: data}
                               ) as r:
                if r.status_code != 201:
                    raise ValueError(f'invalid server status code {r.status_code}')
                imgid = str(r.json()["id"])

            with requests.post(f'http://{self.__host}/api/wares',
                               cookies={AUTH_COOKIE_NAME: self.__login[1]},
                               json={
                                   "title": title,
                                   "description": description,
                                   "price": price,
                                   "image_id": imgid
                               }) as r:
                if r.status_code != 201:
                    raise ValueError(f'invalid server response code {r.status_code}')

            m = '!!! SUCCESS !!!'
            self.__stdscr.addstr(cy, cx - len(m) // 2, m, curses.A_BOLD | curses.A_BLINK)
            self.__refresh_windows()
            self.__stdscr.getch()

            curses.noecho()
            curses.curs_set(False)

            break

    def __user_menu(self):
        while True:
            choice = self.__choose_option('What to do?',
                                          ['buy some staff', 'sell some staff', 'watch account info', 'exit'])

            if choice is None or choice == 'exit':
                break

            if choice == 'buy some staff':
                self.__buy()

            if choice == 'watch account info':
                self.__account_info()

            if choice == 'sell some staff':
                self.__sell_some()

    def __create_user(self):
        while True:
            self.__draw_windows('New account creating')

            curses.echo()
            curses.curs_set(True)

            max_y, max_x = self.__stdscr.getmaxyx()
            cy, cx = max_y // 2, max_x // 2

            rectangle(self.__stdscr, cy - 4, cx - 20, cy + 4, cx + 20)

            self.__stdscr.addstr(cy - 2, cx - 15, 'Login: ', curses.A_BOLD)
            self.__stdscr.addstr(cy, cx - 15, 'Password: ', curses.A_BOLD)
            self.__stdscr.addstr(cy + 2, cx - 15, 'Card: ', curses.A_BOLD)

            self.__refresh_windows()

            login = self.__stdscr.getstr(cy - 2, cx - 15 + len('Login: '), 10)
            password = self.__stdscr.getstr(cy, cx - 15 + len('Password: '), 10)
            credit_card_info = self.__stdscr.getstr(cy + 2, cx - 15 + len('Card: '), 20)

            curses.noecho()
            curses.curs_set(False)

            if not _valid_string(login) or not _valid_string(password) or not _valid_string(credit_card_info):
                m = '!!! PLEASE ENTER VALID DATA !!!'
                self.__stdscr.addstr(cy, cx - len(m) // 2, m, curses.A_BOLD | curses.A_BLINK)
                self.__refresh_windows()
                self.__stdscr.getch()
                continue

            try:
                with requests.post(f'http://{self.__host}/api/users',
                                   json={
                                       'login': login,
                                       'password_hash': get_sha256(password),
                                       'credit_card_info': credit_card_info
                                   }) as r:
                    if r.status_code == 409:
                        m = '!!! CONFLICT !!!'
                        self.__stdscr.addstr(cy, cx - len(m) // 2, m, curses.A_BOLD | curses.A_BLINK)
                        self.__refresh_windows()
                        self.__stdscr.getch()
                        continue

                    if r.status_code == 201:
                        m = '!!! SUCCESS !!!'
                        self.__stdscr.addstr(cy, cx - len(m) // 2, m, curses.A_BOLD | curses.A_BLINK)
                        self.__refresh_windows()
                        self.__stdscr.getch()
                        self.__login = login, r.json()["auth_cookie"], r.json()["cashback"]
                        self.__user_menu()
                        return

                    raise ValueError(r.status_code)
            except Exception as e:
                m = '!!! SOME ERROR CHECK HOST TO CONNECT !!!'
                self.__stdscr.addstr(cy, cx - len(m) // 2, m, curses.A_BOLD | curses.A_BLINK)
                self.__refresh_windows()
                self.__stdscr.getch()
                raise

    def __do_login(self):
        while True:
            self.__draw_windows('Log in')

            curses.echo()
            curses.curs_set(True)

            max_y, max_x = self.__stdscr.getmaxyx()
            cy, cx = max_y // 2, max_x // 2

            rectangle(self.__stdscr, cy - 4, cx - 20, cy + 4, cx + 20)

            self.__stdscr.addstr(cy - 2, cx - 15, 'Login: ', curses.A_BOLD)
            self.__stdscr.addstr(cy + 2, cx - 15, 'Password: ', curses.A_BOLD)

            self.__refresh_windows()

            login = self.__stdscr.getstr(cy - 2, cx - 15 + len('Login: '), 10)
            password = self.__stdscr.getstr(cy + 2, cx - 15 + len('Password: '), 10)

            curses.noecho()
            curses.curs_set(False)

            if not _valid_string(login) or not _valid_string(password):
                m = '!!! PLEASE ENTER VALID DATA !!!'
                self.__stdscr.addstr(cy, cx - len(m) // 2, m, curses.A_BOLD | curses.A_BLINK)
                self.__refresh_windows()
                self.__stdscr.getch()
                continue

            try:
                with requests.put(f'http://{self.__host}/api/users/auth',
                                  json={
                                      'login': login,
                                      'password_hash': get_sha256(password)
                                  }) as r:
                    if r.status_code == 401:
                        m = '!!! INVALID LOGIN OR PASSWORD !!!'
                        self.__stdscr.addstr(cy, cx - len(m) // 2, m, curses.A_BOLD | curses.A_BLINK)
                        self.__refresh_windows()
                        self.__stdscr.getch()
                        continue

                    if r.status_code == 200:
                        m = '!!! SUCCESS !!!'
                        self.__stdscr.addstr(cy, cx - len(m) // 2, m, curses.A_BOLD | curses.A_BLINK)
                        self.__refresh_windows()
                        self.__stdscr.getch()
                        self.__login = login, r.cookies[AUTH_COOKIE_NAME], r.json()["cashback"]
                        self.__user_menu()
                        return

                    raise ValueError(r.status_code)
            except Exception as e:
                m = '!!! SOME ERROR CHECK HOST TO CONNECT !!!'
                self.__stdscr.addstr(cy, cx - len(m) // 2, m, curses.A_BOLD | curses.A_BLINK)
                self.__refresh_windows()
                self.__stdscr.getch()
                raise

    def run(self):
        curses.cbreak()
        curses.noecho()
        self.__stdscr.keypad(True)

        curses.curs_set(False)

        self.__begin_scr()

    def __begin_scr(self):
        choice = self.__choose_option("Let's begin", ["i have an account", "i am new user", "exit"])

        if choice is None or choice == 'exit':
            return 'exit'

        if choice == 'i have an account':
            self.__do_login()

        if choice == 'i am new user':
            self.__create_user()

    def __draw_windows(self, title=None):
        self.__stdscr.clear()

        max_y, max_x = self.__stdscr.getmaxyx()

        l = """
 .________ ________    _________.__
 |   ____//  _____/   /   _____/|  |__   ____ ______
 |____  \/   \  ___   \_____  \ |  |  \ /  _ \\\\____\\
 /       \    \_\  \  /        \|   Y  (  <_> )  |_> >
/______  /\______  / /_______  /|___|  /\____/|   __/
       \/        \/          \/      \/       |__|
"""

        self.__stdscr.addstr(0, 0, l, curses.A_BOLD)

        rectangle(self.__stdscr, 0, 0, 10, 54)
        self.__stdscr.border('|', '|', '=', '=', '+', '+', '+', '+')

        login, _, cashback = self.__login

        if login is not None:
            self.__stdscr.addstr(8, 54 // 2 - len(login) // 2, login)
            c = 'cashback = ' + str(cashback)
            self.__stdscr.addstr(9, 54 // 2 - len(c) // 2, c)
        else:
            self.__stdscr.addstr(8, 54 // 2 - len('UNAUTHORIZED') // 2, 'UNAUTHORIZED', curses.A_BOLD)

        curr_time = datetime.now().strftime("%m/%d/%Y, %H:%M:%S")

        if title is not None:
            self.__stdscr.addstr(3, max_x // 2 - len(title) // 2, title, curses.A_BOLD)

        self.__stdscr.addstr(max_y - 3, max_x // 2 - len(curr_time) // 2, curr_time)

        connected_str = 'connected to ' + self.__host

        self.__stdscr.addstr(max_y - 2, max_x // 2 - len(connected_str) // 2, connected_str)

    def __refresh_windows(self):
        self.__stdscr.refresh()

    def __choose_option(self, title, options: list):
        choice = None
        idx = 0

        while choice is None:
            self.__draw_windows(title)

            max_y, max_x = self.__stdscr.getmaxyx()
            center_y, center_x = max_y // 2, max_x // 2

            for i, option in enumerate(options):
                y = center_y + (i - len(options) // 2) * 3
                s = '-> ' + option + ' <-' if i == idx else option
                x = center_x - len(s) // 2

                if i != idx:
                    self.__stdscr.addstr(y, x, s)
                else:
                    self.__stdscr.addstr(y, x, s, curses.A_BOLD)

            self.__refresh_windows()

            key = self.__stdscr.getch()

            if key == 27:
                break
            elif key == curses.KEY_UP:
                idx -= 1
                if idx < 0:
                    idx = len(options) - 1
            elif key == curses.KEY_DOWN:
                idx += 1
                if idx >= len(options):
                    idx = 0
            elif key == 10 or key == 32:
                choice = options[idx]

        return choice
