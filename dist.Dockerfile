FROM golang:1.20-alpine AS builder

WORKDIR /go/src/app
ENV CGO_ENABLED=0
COPY . .
RUN make build

FROM alpine:latest
COPY --from=builder /go/src/app/cloudflare-utils /
ENTRYPOINT ["/cloudflare-utils"]
CMD ["--help"]