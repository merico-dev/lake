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

# current tag: mericodev/alpine-dbt:0.0.2
FROM --platform=linux/amd64 alpine:3.17.0

RUN apk add --no-cache musl-dev libffi-dev mysql-dev \
    && apk add --no-cache gcc

ENV PYTHONUNBUFFERED=1
RUN apk add --update --no-cache tar curl python3-dev && ln -sf python3 /usr/bin/python
RUN python3 -m ensurepip
RUN pip3 install --no-cache --upgrade pip setuptools
RUN pip3 install dbt-mysql
RUN pip3 install dbt-postgres

RUN curl -sSL https://install.python-poetry.org | python3 -
RUN ln -sf /root/.local/bin/poetry /usr/local/bin
