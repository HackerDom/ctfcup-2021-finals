#!/bin/bash
set -e

REGISTRY=cr.yandex/e2lqkrnu4ardtff6uhc4

src=$1
dst="$REGISTRY/$1"

echo "Mirroring $src -> $dst"
docker pull "$src" \
  && docker tag "$src" "$dst" \
  && docker push "$dst"

