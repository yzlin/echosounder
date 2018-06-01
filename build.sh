#!/bin/bash

set -e

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

version=$(awk '/const Binary/ {print $NF}' < $DIR/internal/version/binary.go | sed 's/"//g')
image=yzlin/echosounder
tag=$version

echo "... Building image ${image}:$tag"
docker build -t ${image}:$tag .
if [[ ! $tag == *"-"* ]]; then
	echo "... Tagging ${image}:$tag as the latest release."
	docker tag ${image}:$tag ${image}:latest
fi
