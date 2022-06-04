# Getting Started

## Download

Cloudflare-utils is build for Windows, Mac and Linux and the latest release is available to download from [GitHub](https://github.com/Cyb3r-Jak3/cloudflare-utils/releases). Download the tar/zip file for your operating system and then extract the progra

To see all the options and commands run `cloudflare-utils --help`

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