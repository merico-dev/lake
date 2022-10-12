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

#===================================== constants =======================================

time_format='
  {
    "type": "time.Time",
    "imports": ["time"]
  }
'

#======================================================================================

json_path=$1 # e.g. "./config/singer/github.json"
tap_stream=$2 # e.g. "issues"
plugin_path=$3 # e.g. "./plugins/github_singer"

if [ $# != 3 ]; then
  printf "not enough args. Usage: <json_path> <tap_stream> <output_path>: e.g.\n    \"./config/singer/github.json\" \"issues\" \"./plugins/github_singer\"\n"
  exit 1
fi

package="generated"
file_name="$tap_stream".go
output_path="$plugin_path/models/generated/$file_name"

tmp_dir=$(mktemp -d -t schema-XXXXX)

json_schema_path="$tmp_dir"/"$tap_stream"

# add, as necessary, more elif blocks for additional transformations
modified_schema=$(cat "$json_path" |  jq --argjson tf "$time_format" '
    .streams[] |
    select(.stream=="'"$tap_stream"'").schema |
      . += { "$schema": "http://json-schema.org/draft-07/schema#" } |
      walk(
        if type == "object" and .format == "date-time" then
          . += { "goJSONSchema": ($tf) }
        elif "place_holder" == "" then
          empty
        else . end
      )
')

# additional cleanup
modified_schema=$(echo "$modified_schema" | sed -r "/\"null\",/d")
modified_schema=$(echo "$modified_schema" | sed -r "/.*additionalProperties.*/d")

echo "$modified_schema" > "$json_schema_path" &&\

# see output
cat "$json_schema_path" | jq -r

exitcode=$?
if [ $exitcode != 0 ]; then
  exit $exitcode
fi

gojsonschema -v -p "$package" "$json_schema_path" -o "$output_path"