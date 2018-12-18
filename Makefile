#
#  Copyright 2018 Ji-Young Park(jiyoung.park.dev@gmail.com)
#
#  Licensed under the Apache License, Version 2.0 (the "License");
#  you may not use this file except in compliance with the License.
#  You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
#  Unless required by applicable law or agreed to in writing, software
#  distributed under the License is distributed on an "AS IS" BASIS,
#  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#  See the License for the specific language governing permissions and
#  limitations under the License.
#

REV=v0.1.0

all: synology-csi-driver

synology-csi-driver:
	mkdir -p bin
	GO111MODULE=on GOOS=linux go build -a -ldflags '-X main.version=$(REV) -extldflags "-static"' -o ./bin/synology-csi-driver ./cmd/syno-csi-plugin

test:
	GO111MODULE=on go test ./...