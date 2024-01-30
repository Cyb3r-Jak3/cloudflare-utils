# DNS Purge

The purpose of DNS purge is to offer a quick way to bulk remove all DNS records.

## Running

Once you have it downloaded run `cloudflare-utils --api-token <API Token with DNS:Edit> --zone-name <your.domain> dns-purge`. It will prompt you to confirm deleting records.

Optional flags:

- `--dry-run`: See what would be deleted without actually deleting anything.
- `--confirm`: Skip the confirmation prompt.


#### Required API Permissions

- _Zone:DNS:Edit_

[Token Quick Link](https://dash.cloudflare.com/profile/api-tokens?permissionGroupKeys=%5B%7B%22key%22%3A%22dns%22%2C%22type%22%3A%22edit%22%7D%5D&name=Cloudflare+Utils%3A+DNS+Write)
