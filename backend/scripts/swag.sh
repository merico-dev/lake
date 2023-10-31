#!/bin/sh
#
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

set -e

ROOT_DIR=$(dirname $(dirname "$0"))

# generate all docs by default, set the working dir (-d .) to root and general api info file (-g ./server/api/api.go).
DOC_DIRS=$ROOT_DIR
GENERAL_API_INFO_PATH=$ROOT_DIR/server/api/api.go


if [ -n "$DEVLAKE_PLUGINS" ]; then
  # change doc dir to avoid generating all docs (-d ./server/api)
  DOC_DIRS=$ROOT_DIR/server/api
  GENERAL_API_INFO_PATH=api.go
  if ! [ "$DEVLAKE_PLUGINS" = "none" ]; then
    # append plugin dirs to the doc dirs
    for plugin in $(echo $DEVLAKE_PLUGINS | tr "," "\n"); do
      DOC_DIRS="$DOC_DIRS,$ROOT_DIR/plugins/$plugin"
    done
  fi
fi

swag -v
ls -l
go get ./...
echo swag init --parseDependency --parseInternal -o $ROOT_DIR/server/api/docs -g $GENERAL_API_INFO_PATH -d $DOC_DIRS
swag init --parseDependency --parseInternal -o $ROOT_DIR/server/api/docs -g $GENERAL_API_INFO_PATH -d $DOC_DIRS
echo "visit the swagger document on http://localhost:8080/swagger/index.html";