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
#Apache DevLake is an effort undergoing incubation at The Apache Software
#Foundation (ASF), sponsored by the Apache Incubator PMC.
#
#Incubation is required of all newly accepted projects until a further review
#indicates that the infrastructure, communications, and decision making process
#have stabilized in a manner consistent with other successful ASF projects.
#
#While incubation status is not necessarily a reflection of the completeness or stability of the code,
#it does indicate that the project has yet to be fully endorsed by the ASF.

FROM mericodev/lake-builder:latest as builder

# docker build --build-arg GOPROXY=https://goproxy.io,direct -t mericodev/lake .
ARG GOPROXY=
# docker build --build-arg HTTPS_PROXY=http://localhost:4780 -t mericodev/lake .
ARG HTTP_PROXY=
ARG HTTPS_PROXY=
ARG TAG=
ARG SHA=

WORKDIR /app
COPY . /app
ENV GOBIN=/app/bin

RUN make clean && make all

FROM --platform=linux/amd64 mericodev/alpine-dbt:0.0.1

EXPOSE 8080

WORKDIR /app

COPY --from=builder /app/bin /app/bin
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /app/requirements.txt /app/requirements.txt
COPY --from=builder /app/config/tap /app/config/tap

# Setup Python
RUN python -m venv /app/.venv
RUN echo "source /app/.venv/bin/activate" >> ~/.profile
RUN source ~/.profile
RUN pip install --upgrade pip -r requirements.txt
RUN apk add --no-cache curl

ENV PATH="/app/bin:${PATH}"

CMD ["lake"]

# Notes: Docker for Mac(M1) sets up qemu emulation, you can try to use the amd64 image by adding the --platform=linux/amd64 flag. 
# Such as: FROM --platform=linux/amd64 alpine:3.15
