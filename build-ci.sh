#!/bin/sh

IMAGE_NAME="${ENVIRONMENT}-${SERVICE_NAME}"

build() {
  (
    docker build -t "${IMAGE_NAME}:latest" -f Dockerfile .
    docker tag "${IMAGE_NAME}:latest" "${DOCKER_REGISTRY}/${IMAGE_NAME}:latest"
    docker tag "${IMAGE_NAME}:latest" "${DOCKER_REGISTRY}/${IMAGE_NAME}:${GITHUB_SHA}"
    docker push "${DOCKER_REGISTRY}/${IMAGE_NAME}" --all-tags
  )
}

if ! build; then
  echo "Build failed!"
  exit 1
fi
