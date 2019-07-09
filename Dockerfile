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

#
# Use 2-stage builds to reduce size of the final docker image
#

# build stage
FROM golang:1.12.6-stretch as builder
WORKDIR /go/src/github.com/jparklab/synology-csi
ADD . .
RUN make 

FROM centos:7.4.1708
LABEL maintainers="Kubernetes Authors"
LABEL description="Synology CSI Plugin"

# NOTE(jparklab):
#   Install open-iscsi instead of iscsi-initiator-utils if base image is ubuntu or debian
RUN yum install ca-certificates e2fsprogs util-linux iscsi-initiator-utils -y

# Install 'ip' tools
RUN yum install iproute -y
COPY --from=builder /go/src/github.com/jparklab/synology-csi/bin/synology-csi-driver /bin/synology-csi-driver

ENTRYPOINT ["/bin/synology-csi-driver"]
