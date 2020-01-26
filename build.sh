#!/usr/bin/env bash

# Do a multistage build

export DOCKER_BUILDKIT=1
export DOCKER_CLI_EXPERIMENTAL=enabled

if ! docker buildx inspect multiarch > /dev/null; then
    docker buildx create --name multiarch
fi
docker buildx use multiarch
docker buildx build --platform linux/amd64,linux/arm64,linux/arm/v7 --push -t $1 .

