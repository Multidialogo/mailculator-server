# Stage 1: Builder
FROM golang:1.23 AS mailculators-builder
RUN mkdir -p /usr/local/go/src/mailculator
WORKDIR /usr/local/go/src/mailculator
COPY . .
COPY .env.test /usr/local/go/src/mailculator/.env
RUN go mod tidy
RUN go mod download
RUN go test ./...
RUN go build -o /usr/local/bin/mailculator .
RUN chmod +x /usr/local/bin/mailculator

# Stage 2: Development
FROM golang:1.23 AS mailculators-dev
WORKDIR /usr/local/go/src/mailculator
COPY . .
COPY .env.dev /usr/local/go/src/mailculator/.env
RUN go mod tidy
RUN go mod download
RUN go install github.com/air-verse/air@latest
EXPOSE 8080
CMD ["air"]

# Stage 3: Production
FROM gcr.io/distroless/base-debian12 AS mailculators-prod
WORKDIR /usr/local/bin/mailculator/server
COPY --from=mailculators-builder /usr/local/bin/mailculator /usr/local/bin/mailculator/server
COPY .env.prod /usr/local/bin/mailculator/.env
EXPOSE 8080
CMD ["server"]
