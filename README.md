# Cloudflare Utilities

[![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/Cyb3r-Jak3/cloudflare-utils)](https://github.com/Cyb3r-Jak3/cloudflare-utils/releases/latest)

[![Go Checks](https://github.com/Cyb3r-Jak3/cloudflare-utils/actions/workflows/golang.yml/badge.svg)](https://github.com/Cyb3r-Jak3/cloudflare-utils/actions/workflows/golang.yml) [![GolangLint CI](https://github.com/Cyb3r-Jak3/cloudflare-utils/actions/workflows/golangci-lint.yml/badge.svg)](https://github.com/Cyb3r-Jak3/cloudflare-utils/actions/workflows/golangci-lint.yml) [![codecov](https://codecov.io/gh/Cyb3r-Jak3/cloudflare-utils/graph/badge.svg?token=p1NsbLftFq)](https://codecov.io/gh/Cyb3r-Jak3/cloudflare-utils)

![GitHub all releases](https://img.shields.io/github/downloads/Cyb3r-Jak3/cloudflare-utils/total?label=GitHub%20Total%20Downloads) ![Chocolatey](https://img.shields.io/chocolatey/dt/cloudflare-utils?label=Chocolatey%20Downloads)


## About

This is a collection of utilities for Cloudflare. The utilities are written in Go and are cross-platform. The utilities are:

* [DNS Cleaner](https://cloudflare-utils.cyberjake.xyz/dns/cleaner/)
* [DNS Purge](https://cloudflare-utils.cyberjake.xyz/dns/purge/)
* [Deployment Purge](https://cloudflare-utils.cyberjake.xyz/pages/purge-deployments/)
* [Deployment Prune](https://cloudflare-utils.cyberjake.xyz/pages/prune-deployments/)
* [List Tunnel Version](https://cloudflare-utils.cyberjake.xyz/tunnels/list-versions/)
* [Sync IP List](https://cloudflare-utils.cyberjake.xyz/lists/sync-list/)

## Installation

### Chocolatey

```powershell
choco install cloudflare-utils
```

### GitHub

Download the latest release from the [releases page](https://github.com/Cyb3r-Jak3/cloudflare-utils/releases/latest)

### Docker

```bash
docker pull cyb3rjak3/cloudflare-utils
```

### Go

```bash
go install github.com/Cyb3r-Jak3/cloudflare-utils/cmd/cloudflare-utils@latest
```

### Homebrew

```bash
brew install cyb3r-jak3/cyberjake/cloudflare-utils
```


### GitHub Actions

```yaml
- name: Cloudflare Utils - DNS Purge
  uses: Cyb3r-Jak3/actions-cloudflare-utils@v1
  with:
     version: 'latest' # Optional, defaults to latest
     args: '--zone-name example.com dns-purge --confirm'
  env:
    CLOUDFLARE_API_TOKEN: ${{ secrets.CLOUDFLARE_API_TOKEN }}
```

More:
 - [Repo](https://github.com/Cyb3r-Jak3/actions-cloudflare-utils)
 - [Marketplace](https://github.com/marketplace/actions/cloudflare-utils)

## Usage

Check the [docs](https://cloudflare-utils.cyberjake.xyz/) for more information on how to use the utilities.

## Note

This project is not affiliated with Cloudflare in any way.
