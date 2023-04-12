# DNS Purge

The purpose of DNS purge is to offer a quick way to bulk remove all DNS records.

**Required API Permissions**:
- _Zone:DNS:Edit_

## Running

Once you have it downloaded run `cloudflare-utils --api-token <API Token with DNS:Edit> --zone-name <your.domain> dns-purge`. It will prompt you to confirm deleting records.

Optional flags:

- `--dry-run`: See what would be deleted without actually deleting anything.
- `--confirm`: Skip the confirmation prompt.