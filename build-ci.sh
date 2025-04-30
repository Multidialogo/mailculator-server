#!/bin/sh

script_dir=$(dirname "$(realpath -s "$0")")
root_dir="$script_dir"

mkdir -p .cache/venv

IMAGE_NAME="${ENVIRONMENT}-${SERVICE_NAME}"

build() {
  (
    docker build --target "${IMAGE_NAME}" .
    docker tag "${IMAGE_NAME}:latest" "${DOCKER_REGISTRY}/${IMAGE_NAME}:latest"
    docker tag "${IMAGE_NAME}:latest" "${DOCKER_REGISTRY}/${IMAGE_NAME}:${CUSTOM_TAG}"
  )
}

if ! build; then
  echo "Build failed!"
  exit 1
fi
