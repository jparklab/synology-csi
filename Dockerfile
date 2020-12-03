#
# Copyright 2018 Ji-Young Park(jiyoung.park.dev@gmail.com)
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

# default values
ARG BUILDPLATFORM="linux/amd64"
ARG TARGETPLATFORM="linux/amd64"
#
# Use 2-stage builds to reduce size of the final docker image
#

# Cache modules. The cache will be then used by all build stages
FROM --platform=$BUILDPLATFORM golang:1.13.6-alpine as builder
ARG BUILDPLATFORM
RUN apk add --no-cache alpine-sdk
WORKDIR /go/src/github.com/jparklab/synology-csi
COPY go.mod .
RUN go mod download

# Cross-platform build. Build everything on AMD64 and cross compile
# as this is much faster than buildin with QEMU on ARMv7
FROM --platform=$BUILDPLATFORM builder as compiler

COPY Makefile .

# Copy over a dummy main file to populate build cache. We do this so
# that next builds will be faster. The file does exactly nothing more
# and nothing less than just pull in all imports.
COPY dummy/main.go ./

ARG TARGETPLATFORM
RUN env \
        CGO_ENABLED=0 \
        GOARM=$(echo "$TARGETPLATFORM" | cut -f3 -d/ | cut -c2-) \
        GOARCH=$(echo "$TARGETPLATFORM" | cut -f2 -d/) \
    make dummy && rm bin/dummy main.go
COPY cmd ./cmd
COPY pkg ./pkg
RUN env \
        CGO_ENABLED=0 \
        GOARM=$(echo "$TARGETPLATFORM" | cut -f3 -d/ | cut -c2-) \
        GOARCH=$(echo "$TARGETPLATFORM" | cut -f2 -d/) \
    make

# Alpine is provided for different architectures, amd64, arm32 and arm64
FROM alpine:latest

LABEL maintainers="Kubernetes Authors"
LABEL description="Synology CSI Plugin"

RUN apk add --no-cache e2fsprogs e2fsprogs-extra xfsprogs xfsprogs-extra util-linux iproute2 blkid
COPY --from=compiler /go/src/github.com/jparklab/synology-csi/bin/synology-csi-driver synology-csi-driver

ENTRYPOINT ["/synology-csi-driver"]
