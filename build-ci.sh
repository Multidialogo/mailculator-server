#!/bin/sh

script_dir=$(dirname "$(realpath -s "$0")")
root_dir="$script_dir"

mkdir -p .cache/venv

build() {
  (
    docker compose run --rm --build build
  )
}

if ! build; then
  echo "Build failed!"
  exit 1
fi
