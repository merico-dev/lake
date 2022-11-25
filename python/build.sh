#!/bin/sh

cd "${0%/*}" # make sure we're in the correct dir

poetry config virtualenvs.create true

for plugin_dir in $(ls -d plugins/*/*.toml); do
  plugin_dir=$(dirname $plugin_dir)
  echo "installing dependencies of python plugin in: $plugin_dir" &&\
  cd "$plugin_dir" &&\
  poetry install &&\
  cd -
  exit_code=$?
  if [ $exit_code != 0 ]; then
    exit $exit_code
  fi
done