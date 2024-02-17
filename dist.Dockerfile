FROM golang:1.21-alpine AS builder

WORKDIR /usr/app
ENV CGO_ENABLED=0
RUN apk update && apk -U --no-cache add git make build-base ca-certificates && git config --global --add safe.directory '*'
COPY . .
RUN make build

FROM scratch
COPY --from=builder /usr/app/cloudflare-utils /
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/cloudflare-utils"]
CMD ["--help"]