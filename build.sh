#!/usr/bin/env bash

# Do a multistage build
BUILDER_INSTANCE_NAME="synology-csi-multiarch"

export DOCKER_BUILDKIT=1
export DOCKER_CLI_EXPERIMENTAL=enabled

if docker buildx inspect ${BUILDER_INSTANCE_NAME} > /dev/null; then
    # clean up any stale instance 
    docker buildx rm --name ${BUILDER_INSTANCE_NAME}
fi

docker buildx create --name ${BUILDER_INSTANCE_NAME}
docker buildx use ${BUILDER_INSTANCE_NAME}

if [[ "$*" == *--push* ]]; then
    if [[ -n "$DOCKER_USERNAME" ]] && [[ -n "$DOCKER_PASSWORD" ]]; then
        echo "Logging into docker registry $DOCKER_REGISTRY_URL...."
        echo "$DOCKER_PASSWORD" | docker login --username $DOCKER_USERNAME --password-stdin $DOCKER_REGISTRY_URL
    fi
fi

if [[ -z "$PLATFORMS" ]]; then
    PLATFORMS="linux/amd64,linux/arm64,linux/arm/v7"
fi

docker buildx build --platform $PLATFORMS . $*
