
# MultiDialogo - MailCulator server

## Provisioning

This Dockerfile is designed to build and deploy the `mailculator server` application using three distinct stages:
1. Builder Stage
2. Development Stage
3. Production Stage

Each stage serves a specific purpose, and you can use them based on your needs.

### Stage 1: Builder

Purpose:
This stage is responsible for building the Go application, running tests, and preparing it for further stages (Development or Production).

Description:
- The base image used is `golang:1.23`.
- The `mailculator server` project is copied into the container and the necessary dependencies are downloaded using `go mod tidy` and `go mod download`.
- The tests are run with `go test ./...` to ensure everything is correct.
- The application is built with `go build` and the resulting binary is copied to `/usr/local/bin/mailculator-server`.
- Finally, the binary is made executable with `chmod +x`.

To build the image:
```bash
 docker build --no-cache -t mailculators-builder --target mailculators-builder .
 ```

To introspect the builder image:

```bash
docker run -ti --rm mailculators-builder bash
```

### Stage 2: Development

Purpose: This stage is used for local development. It allows you to run the mailculator application with live-reload using air.

Description:

The base image used is golang:1.23.
The binary generated in the builder stage is copied into this container.
The air tool is installed, which enables live-reload for Go projects during development.
The container exposes port 8080, which can be used for development and debugging.
To build the image:
```bash
docker build -t mailculators-dev --target mailculators-dev .
```

To run the development container:
```bash
docker run -v$(pwd)/data:/var/lib/mailculator-server -p 8080:8080 mailculators-dev
```

### Stage 3: Production

Purpose: This stage is optimized for production deployment. It creates a minimal container to run the mailculator server binary in a secure and efficient environment.

Description:

The base image used is gcr.io/distroless/base-debian12, which is a minimal image without unnecessary tools or packages.
The binary from the builder stage is copied into the container.
Port 8080 is exposed for production use.
The container is configured to run the mailculator binary.
To build the image:
```bash
docker build -t mailculators-prod --target mailculators-prod .
```

### API clients

Generate clients from open api:

```bash
docker run --rm -v ${PWD}:/local openapitools/openapi-generator-cli generate \
    -i /local/openapi.yaml \
    -g <language> \
    -o /local/client/<language>
```

For example for php:

```bash
docker run --rm -v ${PWD}:/local openapitools/openapi-generator-cli generate \
    -i /local/openapi.yaml \
    -g php \
    -o /local/client/php
```

Then you will find the generated client in the directory client/php in the root path of this repository.

