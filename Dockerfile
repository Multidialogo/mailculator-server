# Stage 1: Builder
FROM golang:1.23 AS builder
RUN mkdir -p /usr/local/go/src/mailculator
WORKDIR /usr/local/go/src/mailculator
COPY . .
RUN go mod tidy
RUN go mod download
RUN go test ./...
RUN go build -o /usr/local/bin/mailculator .
RUN chmod +x /usr/local/bin/mailculator

# Stage 2: Development
FROM golang:1.23 AS dev
WORKDIR /usr/local/go/src/mailculator
COPY --from=builder /usr/local/bin/mailculator /usr/local/bin/mailculator
RUN go install github.com/air-verse/air@latest
EXPOSE 8080
CMD ["air"]

# Stage 3: Production
FROM gcr.io/distroless/base-debian12 AS prod
WORKDIR /usr/local/go/src/mailculator
COPY --from=builder /usr/local/bin/mailculator /usr/local/bin/mailculator
EXPOSE 8080
CMD ["mailcalculator"]
