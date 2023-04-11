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

## Global Flags

- `--account-id`
You can pass your account ID with the `--account-id` flag or with the environment variable `CLOUDFLARE_ACCOUNT_ID`

- `--api-email`
You can pass your API email with the `--api-email` flag or with the environment variable `CLOUDFLARE_API_EMAIL`. This is only needed if you are using the legacy auth method.

- `--api-key`
You can pass your API key with the `--api-key` flag or with the environment variable `CLOUDFLARE_API_KEY`. This is only needed if you are using the legacy auth method.

- `--api-token`
You can pass your API token with the `--api-token` flag or with the environment variable `CLOUDFLARE_API_TOKEN`. This is the recommended method of authentication.

- `--rate-limit`
You can pass the rate limit in milliseconds with the `--rate-limit` flag. This is useful if you are getting rate limited by Cloudflare or want to speed up the rate of requests.

- `--zone-name`
You can pass your zone name with the `--zone-name` flag or with the environment variable `CLOUDFLARE_ZONE_NAME`. This is useful if you are running a command that only requires a zone name.

- `--zone-id`
You can pass your zone ID with the `--zone-id` flag or with the environment variable `CLOUDFLARE_ZONE_ID`. This is useful if you are running a command that only requires a zone ID.