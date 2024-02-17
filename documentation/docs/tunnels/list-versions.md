# Tunnel Version

The purpose of tunnel version is to offer a quick way to check which tunnels are using outdated versions of cloudflared. This works by checking the version of cloudflared that all connectors are running and comparing it to the latest version available.

## Running

Once you have it downloaded run `cloudflare-utils --api-token <API Token with Cloudflare Tunnel:Read> --account-id <account id> tunnel-versions`. It will print a list of all out of date tunnels and their versions.

Optional flags:

- `all-tunnels`: If you want to see all tunnels, not just the out of date ones.
- `include-deleted`: If you want to include deleted tunnels in the list.
- `healthy-only`: If you want to only see healthy tunnels in the list.


#### Required API Permissions

[Token Quick Link](https://dash.cloudflare.com/profile/api-tokens?permissionGroupKeys=%5B%7B%22key%22%3A%22argotunnel%22%2C%22type%22%3A%22read%22%7D%5D&name=Cloudflare+Utils%3A+Tunnels+Read)

- _Account:Cloudflare Tunnel:Read_