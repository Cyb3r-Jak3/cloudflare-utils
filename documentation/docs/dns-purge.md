# DNS Purge

The purpose of DNS purge is to offer a quick way to bulk remove all DNS records.

**Required API Permissions**: _DNS:Edit_

## Running

It is very easy to run. Once you have it downloaded just run `cloudflare-utils --api-token <API Token with DNS:Edit> --zone-name <your.domain> dns-purge`. It will prompt you to confirm deleting records. If you want to auto remove then add `--confirm`