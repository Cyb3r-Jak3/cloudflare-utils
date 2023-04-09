FROM scratch
COPY cloudflare-utils /cloudflare-utils
ENTRYPOINT ["/cloudflare-utils"]
CMD ["--help"]