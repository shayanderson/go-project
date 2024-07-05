FROM golang:1.22 AS builder

ARG PROJECT

ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0

WORKDIR /app
COPY . .

RUN go mod download
RUN --mount=type=cache,target="/root/.cache/go-build" \
  go build -o /$PROJECT /app/cmd/app

FROM alpine:latest

ARG PROJECT

COPY --from=builder /$PROJECT /$PROJECT

CMD ["/$PROJECT"]