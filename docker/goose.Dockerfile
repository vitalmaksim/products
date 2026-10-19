FROM golang:1.27-alpine AS build

RUN CGO_ENABLED=0 go install github.com/pressly/goose/v3/cmd/goose@v3.24.1

FROM alpine:3.24

COPY --from=build /go/bin/goose /usr/local/bin/goose

WORKDIR /migrations

ENTRYPOINT ["goose"]

