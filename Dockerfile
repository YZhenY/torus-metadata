# Builder image, produces a statically linked binary
FROM golang:1.13.4-alpine3.10 as app-builder
RUN apk update && apk add libstdc++ g++

WORKDIR /src
COPY . ./
RUN go build -mod=vendor

FROM alpine:3.10
RUN apk update && apk add ca-certificates --no-cache

RUN mkdir -p /app
COPY --from=app-builder /src/torus-metadata /app/torus-metadata

EXPOSE 22 80 443 3000 5051 5001
VOLUME ["/app"]
CMD ["/app/torus-metadata"]