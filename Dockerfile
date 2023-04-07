FROM scratch
COPY cloudflare-utils /
ENTRYPOINT ["/cloudflare-utils"]
CMD ["--help"]