FROM golang:1.24-alpine

WORKDIR /tmp/app

COPY . /tmp/app

RUN go mod tidy && go build -o bin/main .

WORKDIR /app/bin

RUN cp /tmp/app/bin/main /app/bin/main

CMD ["./main"]
