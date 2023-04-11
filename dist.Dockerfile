FROM golang:1.20-alpine AS builder

WORKDIR /usr/app
ENV CGO_ENABLED=0
RUN apk update && apk add git make build-base
COPY . .
RUN make build

FROM alpine:latest
COPY --from=builder /usr/app/cloudflare-utils /
ENTRYPOINT ["/cloudflare-utils"]
CMD ["--help"]