#!/usr/bin/env python3
import random
import os

user_agents_list = []


def get_random_user_agent():
    global user_agents_list

    if len(user_agents_list) == 0:
        error = None
        for i in range(2):
            try:
                error = None
                user_agents_list = __get()
            except Exception as e:
                error = e
        if error is not None:
            raise OSError(str(error))

    return random.choice(user_agents_list)


def __get():
    with open(f'{os.path.dirname(os.path.realpath(__file__))}/useragents') as fin:
        return [line.strip() for line in fin]
