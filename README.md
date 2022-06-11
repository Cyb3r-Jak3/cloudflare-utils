# Cloudflare Utilities

[![Go Checks](https://github.com/Cyb3r-Jak3/cloudflare-utils/actions/workflows/golang.yml/badge.svg)](https://github.com/Cyb3r-Jak3/cloudflare-utils/actions/workflows/golang.yml) [![Golanglint CI](https://github.com/Cyb3r-Jak3/cloudflare-utils/actions/workflows/golangci-lint.yml/badge.svg)](https://github.com/Cyb3r-Jak3/cloudflare-utils/actions/workflows/golangci-lint.yml)

[![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/Cyb3r-Jak3/cloudflare-utils)](https://github.com/Cyb3r-Jak3/cloudflare-utils/releases/latest)

## Tools

### DNS Cleaner

**DNS Cleaner** is a tool that downloads DNS records to a YAML file then will you will edit and your records
 
Basic Steps:

1. Download your DNS records `./cloudflare-utils dns-cleaner`
2. There will be a file called `dns-records.yml`
3. For any record you do not want to keep change `keep: true` to false
4. Rerun `./cloudflare-utils dns-cleaner` and all records not marked to keep will be removed. **This tool does not recreate records if they are missing**

[Read More](https://cloudflare-utils.cyberjake.xyz/dns-cleaner/)
