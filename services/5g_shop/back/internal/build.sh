#!/usr/bin/env bash

set -uex

make --jobs=9
cp 5g_shop ../../docker/back/5g_shop
