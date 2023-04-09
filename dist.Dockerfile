FROM busybox:1.36
COPY cloudflare-utils /cloudflare-utils
ENTRYPOINT ["/cloudflare-utils"]
CMD ["--help"]