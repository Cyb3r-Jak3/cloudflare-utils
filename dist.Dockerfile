FROM --platform=$BUILDPLATFORM library/golang:1.27-alpine@sha256:4c9fe60190a2a3350ddc51de80d0224b8a6698d12bdfc999fee45ea9d6c46dbc AS builder

WORKDIR /usr/app
ENV CGO_ENABLED=0
RUN  --mount=type=cache,target=/var/cache/apk,sharing=locked apk update && apk -U --no-cache add git curl build-base ca-certificates && git config --global --add safe.directory '*'
RUN sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b /usr/local/bin
COPY . .
ENV GOCACHE=/root/.cache/go-build
ARG TARGETOS
ARG TARGETARCH
RUN --mount=type=cache,target="/root/.cache/go-build" --mount=type=cache,target=/go/pkg GOOS=${TARGETOS} GOARCH=${TARGETARCH} task build

FROM scratch
COPY --from=builder /usr/app/cloudflare-utils /
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/cloudflare-utils"]
CMD ["--help"]