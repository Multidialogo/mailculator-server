
# MultiDialogo - MailCulator

## Provisioning

This Dockerfile is designed to build and deploy the `mailcalculator` application using three distinct stages:
1. Builder Stage
2. Development Stage
3. Production Stage

Each stage serves a specific purpose, and you can use them based on your needs.

### Stage 1: Builder

Purpose:
This stage is responsible for building the Go application, running tests, and preparing it for further stages (Development or Production).

Description:
- The base image used is `golang:1.23`.
- The `mailcalculator` project is copied into the container and the necessary dependencies are downloaded using `go mod tidy` and `go mod download`.
- The tests are run with `go test ./...` to ensure everything is correct.
- The application is built with `go build` and the resulting binary is copied to `/usr/local/bin/mailcalculator`.
- Finally, the binary is made executable with `chmod +x`.

To build the image:
```bash
docker build -t mailcalculator-builder --target builder .
```

### Stage 2: Development

Purpose: This stage is used for local development. It allows you to run the mailcalculator application with live-reload using air.

Description:

The base image used is golang:1.23.
The binary generated in the builder stage is copied into this container.
The air tool is installed, which enables live-reload for Go projects during development.
The container exposes port 8080, which can be used for development and debugging.
To build the image:
```bash
docker build -t mailcalculator-dev --target dev .
```

To run the development container:
```bash
docker run -p 8080:8080 mailcalculator-dev
```

### Stage 3: Production

Purpose: This stage is optimized for production deployment. It creates a minimal container to run the mailcalculator binary in a secure and efficient environment.

Description:

The base image used is gcr.io/distroless/base-debian12, which is a minimal image without unnecessary tools or packages.
The binary from the builder stage is copied into the container.
Port 8080 is exposed for production use.
The container is configured to run the mailcalculator binary.
To build the image:
```bash
docker build -t mailcalculator-prod --target prod .
```

To run the production container:
```bash
docker run -p 8080:8080 mailcalculator-prod
```

### General Instructions

To build and use the images, you can use the following Docker commands:

To build the builder image:
```bash
docker build -t mailcalculator-builder --target builder .
```

To build the development image:
```bash
docker build -t mailcalculator-dev --target dev .
```

To build the production image:
```bash
docker build -t mailcalculator-prod --target prod .
```
To run the container in development mode (with live-reload using air):

```bash
docker run -p 8080:8080 mailcalculator-dev
```

To run the container in production mode:
```bash
docker run -p 8080:8080 mailcalculator-prod
```
