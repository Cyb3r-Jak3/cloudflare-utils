FROM gcr.io/distroless/static-debian11:nonroot
COPY cloudflare-utils /
ENTRYPOINT ["/cloudflare-utils"]
CMD ["--help"]