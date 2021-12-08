#!/usr/bin/env bash

set -uex

clang++ -Wall -fPIC -shared -static-libgcc -static-libstdc++ -o libcalc.so calc.cpp

