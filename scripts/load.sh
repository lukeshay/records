#!/bin/bash

set -eou pipefail

IMAGE_ID="$(docker load --quiet --input "./.deployer/artifacts/images/$1.tar" | sed 's/^Loaded image ID: //')"

docker tag $IMAGE_ID $1
