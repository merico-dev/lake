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

FROM node:16 as builder
WORKDIR /home/node/code
COPY package.json /home/node/code
COPY package-lock.json /home/node/code
RUN npm ci
COPY . /home/node/code
RUN npm run build-production

FROM nginxinc/nginx-unprivileged:1.21
USER 0
#ADD https://github.com/ufoscout/docker-compose-wait/releases/download/2.9.0/wait /wait
#RUN chmod +x /wait
RUN rm /etc/nginx/conf.d/default.conf
COPY ./nginx.conf /etc/nginx/conf.d/default.conf.tpl
WORKDIR /usr/share/nginx/html
RUN rm -rf ./*
COPY --from=builder /home/node/code/dist/. ./
EXPOSE 80 443
RUN apt update && apt install -y apache2-utils
COPY ./nginx.sh /usr/bin/nginx.sh
RUN chmod +x /usr/bin/nginx.sh
USER 101
CMD nginx.sh
