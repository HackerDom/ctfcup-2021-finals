#!/bin/bash

set -e

SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" &> /dev/null && pwd)

cd $SCRIPT_DIR

for image in $(cat "./image_list.txt"); do
    echo "ko: $image"
    ./mirror_image.sh "$image"
done
