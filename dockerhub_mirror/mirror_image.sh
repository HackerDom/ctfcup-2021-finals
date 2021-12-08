#!/bin/bash
set -e

REGISTRY=cr.yandex/crpogk7287k25tqmrogj

src=$1
dst="$REGISTRY/$1"

echo "Mirroring $src -> $dst"
docker pull "$src" \
  && docker tag "$src" "$dst" \
  && docker push "$dst"

