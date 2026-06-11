FROM --platform=$BUILDPLATFORM library/golang:1.26-alpine@sha256:bd14630652464086289693533d25b791aa9ae7481e784d7eac5d4c948e9736ea AS builder

WORKDIR /usr/app
ENV CGO_ENABLED=0
RUN  --mount=type=cache,target=/var/cache/apk,sharing=locked apk update && apk -U --no-cache add git make build-base ca-certificates && git config --global --add safe.directory '*'
COPY . .

ENV GOCACHE=/root/.cache/go-build
ARG TARGETOS
ARG TARGETARCH
RUN --mount=type=cache,target="/root/.cache/go-build" --mount=type=cache,target=/go/pkg GOOS=${TARGETOS} GOARCH=${TARGETARCH} make build

FROM scratch
COPY --from=builder /usr/app/cloudflare-utils /
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/cloudflare-utils"]
CMD ["--help"]