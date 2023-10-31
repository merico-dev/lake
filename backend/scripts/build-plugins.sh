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
#   PLUGIN=<PLUGIN_NAME[,PLUGIN_NAME2]> make dev
#
# compile all plugins and fire up api server in DEBUG MODE with `delve`:
#   make debug
#
# compile specific plugin and fire up api server in DEBUG MODE with `delve`:
#   PLUGIN=<PLUGIN_NAME[,PLUGIN_NAME2]> make debug

set -e

echo "Usage: "
echo "  build all plugins:              $0 [golang build flags...]"
echo "  build and keep specified plugins only: DEVLAKE_PLUGINS=github,jira $0 [golang build flags...]"

ROOT_DIR=$(dirname $(dirname "$0"))
EXTRA=""
PLUGIN_SRC_DIR=$ROOT_DIR/plugins
PLUGIN_OUTPUT_DIR=${PLUGIN_DIR:-$ROOT_DIR/bin/plugins}

if [ -n "$DEVLAKE_DEBUG" ]; then
    EXTRA="-gcflags='all=-N -l'"
fi

if [ -z "$DEVLAKE_PLUGINS" ]; then
    echo "Building all plugins"
    PLUGINS=$(find $PLUGIN_SRC_DIR/* -maxdepth 0 -type d -not -name core -not -name helper -not -name logs -not -empty)
else
    echo "Building the following plugins: $PLUGIN"
    PLUGINS=
    for p in $(echo "$DEVLAKE_PLUGINS" | tr "," "\n"); do
        PLUGINS="$PLUGINS $PLUGIN_SRC_DIR/$p"
    done
fi


rm -rf $PLUGIN_OUTPUT_DIR/*

PIDS=""
for PLUG in $PLUGINS; do
    NAME=$(basename $PLUG)
    echo "Building plugin $NAME to bin/plugins/$NAME/$NAME.so with args: $*"
    go build -buildmode=plugin $EXTRA -o $PLUGIN_OUTPUT_DIR/$NAME/$NAME.so $PLUG/*.go &
    PIDS="$PIDS $!"
    # avoid too many processes causing signal killed
    COUNT=$(echo "$PIDS" | wc -w)
    PARALLELISM=4
    if command -v nproc >/dev/null 2>&1; then
        PARALLELISM=$(nproc)
    elif command -v sysctl >/dev/null 2>&1; then
        PARALLELISM=$(sysctl -n hw.ncpu)
    fi
    if [ "$COUNT" -ge "$PARALLELISM" ]; then
        for PID in $PIDS; do
            wait $PID
        done
        PIDS=""
    fi
done

for PID in $PIDS; do
    wait $PID
done