FROM golang:1.23
RUN mkdir -p /usr/local/go/src/mailculator-server
WORKDIR /usr/local/go/src/mailculator-server
COPY ../../../. .
RUN go mod tidy
RUN go mod download
RUN go install github.com/air-verse/air@latest
EXPOSE 8080