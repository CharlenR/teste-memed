FROM golang:1.25-alpine AS base
WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go install github.com/air-verse/air@latest
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./processor ./cmd/processor


FROM alpine:3.20 AS api
WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY --from=base /app/api .
CMD ["./api"]

FROM alpine:3.20 AS processor
WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY --from=base /app/processor .
CMD ["./processor"]