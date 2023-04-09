# Getting Started

## Download

Cloudflare-utils is build for Windows, Mac and Linux and the latest release is available to download from [GitHub](https://github.com/Cyb3r-Jak3/cloudflare-utils/releases). Download the tar/zip file for your operating system and then extract the executable.

### Additional Installation Methods

1. Docker 
   There is also a docker image available at `cyb3rjak3/cloudflare-utils` or `ghcr.io/cyb3r-jak3/cloudflare-utils`
2. Chocolatey
   There is also a chocolatey package available at `cloudflare-utils`

## Authentication

### API Token

The recommended method to authenticate is with an [API Token](https://developers.cloudflare.com/api/tokens/create/). Each command will list the API permissions needed for it to run.

**To Use**

`cloudflare-utils --api-token <Token Here>`

You can also pass your API Token via an environment variable of `CLOUDFLARE_API_TOKEN`

### API Key

The legacy [API Key](https://developers.cloudflare.com/api/keys/) method is also supported

**To Use**

`cloudflare-utils --api-email <Email Here> --api-key <API Key Here>`

You can pass your API email and key with environment variables of `CLOUDFLARE_API_EMAIL` and `CLOUDFLARE_API_KEY`