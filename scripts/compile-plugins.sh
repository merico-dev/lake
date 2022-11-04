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
#

# compile all plugins and fire up api server:
#   make dev
#
# compile specific plugin and fire up api server:
#   PLUGIN=<PLUGIN_NAME> make dev
#   PLUGIN=<PLUGIN_NAME> PLUGIN2=<PLUGIN_NAME2> make dev
#
# compile all plugins and fire up api server in DEBUG MODE with `delve`:
#   make debug
#
# compile specific plugin and fire up api server in DEBUG MODE with `delve`:
#   PLUGIN=<PLUGIN_NAME> make dev
#   PLUGIN=<PLUGIN_NAME> PLUGIN2=<PLUGIN_NAME> make dev

set -e

echo "Usage: "
echo "  build all plugins:              $0 [golang build flags...]"
echo "  build and keep one plugin only: PLUGIN=jira $0 [golang build flags...]"
echo "  build and keep two plugin only: PLUGIN=jira PLUGIN2=github $0 [golang build flags...]"

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"
PLUGIN_SRC_DIR=$SCRIPT_DIR/../plugins
PLUGIN_OUTPUT_DIR=$SCRIPT_DIR/../bin/plugins

if [ -z "$PLUGIN" ]; then
    PLUGINS=$(find $PLUGIN_SRC_DIR/* -maxdepth 0 -type d -not -name core -not -name helper -not -empty)
else
    PLUGINS=$PLUGIN_SRC_DIR/$PLUGIN
fi

if [ $PLUGIN ] && [ $PLUGIN2 ]; then
    PLUGINS="$PLUGINS $PLUGIN_SRC_DIR/$PLUGIN2"
fi

rm -rf $PLUGIN_OUTPUT_DIR/*

PIDS=""
for PLUG in $PLUGINS; do
    NAME=$(basename $PLUG)
    echo "Building plugin $NAME to bin/plugins/$NAME/$NAME.so"
    go build -buildmode=plugin "$@" -o $PLUGIN_OUTPUT_DIR/$NAME/$NAME.so $PLUG/*.go &
    PIDS="$PIDS $!"
done

for PID in $PIDS; do
    wait $PID
done
