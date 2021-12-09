#!/bin/bash
set -e

REGISTRY=cr.yandex/crp649c8570akro5vmp6

src=$1
dst="$REGISTRY/$1"

echo "Mirroring $src -> $dst"
docker pull "$src" \
  && docker tag "$src" "$dst" \
  && docker push "$dst"

