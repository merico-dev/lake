# Licensed to the Apache Software Foundation (ASF) under one or more
# contributor license agreements.  See the NOTICE file distributed with
# this work for additional information regarding copyright ownership.
# The ASF licenses this file to You under the Apache License, Version 2.0
# (the "License"); you may not use this file except in compliance with
# the License.  You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

# current tag: mericodev/lake-builder:0.0.9
FROM alpine:3.15 as builder
# this distro version gives us libgit2 1.3.0 which we want for gitextractor
RUN apk update && apk upgrade && apk add libgit2-dev
RUN mkdir /deps &&\
    cp --parents -r /usr/include /deps &&\
    cp --parents -r /usr/lib/*.so* /deps &&\
    cp --parents -r /lib/*.so* /deps &&\
    cp --parents -r /usr/lib/pkgconfig /deps &&\
    cp --parents -r /usr/bin/pkgconf /deps &&\
    cp --parents -r /usr/bin/pkg-config /deps

FROM golang:1.19.3-alpine3.16

COPY --from=builder /deps /deps
#RUN apk add --update gcc=130.2.1_pre1-r3 g++=10.2.1_pre1-r3
RUN apk update && apk upgrade && apk add --no-cache tzdata gcc g++ make tar git
RUN go install github.com/vektra/mockery/v2@v2.12.3
RUN go install github.com/swaggo/swag/cmd/swag@v1.8.4