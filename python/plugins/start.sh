#!/bin/sh

endpoint="$1"

cd "${0%/*}" # make sure we're in the correct dir

for plugin_dir in $(ls -d */*.toml); do
  plugin_dir=$(dirname $plugin_dir)
  cd $plugin_dir &&\
  poetry run python $plugin_dir/main.py startup "$endpoint" &&\
  cd -
  exit_code=$?
  if [ $exit_code != 0 ]; then
    exit $exit_code
  fi
done