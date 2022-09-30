#!/bin/sh

json_path=$1 # e.g. "./github.json"
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

echo $json_schema_path

cat "$json_path" |  jq '
    .streams[] |
    select(.stream=="'"$tap_stream"'").schema |
    . += { "$schema": "http://json-schema.org/draft-07/schema#" }
  ' > "$json_schema_path" &&\
sed -i -r "/\"null\",/d" "$json_schema_path"
sed -i -r "/.*additionalProperties.*/d" "$json_schema_path"

# see output
cat "$json_schema_path" | jq -r

exitcode=$?
if [ $exitcode != 0 ]; then
  exit $exitcode
fi

gojsonschema -v -p "$package" "$json_schema_path" -o "$output_path"