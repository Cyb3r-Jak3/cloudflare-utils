# Sync List

The purpose of this command is to offer a way to sync a list of ips fetched from a URL with a Cloudflare list. Which can then be used in firewall rules, rate limiting rules, or other places that support lists.

There are currently 3 supported sources:

- file
- url
- preset

All items of the list will be replaced with the new items.

#### File

Use the `file://` prefix to read a list from a local file. The file should contain one ip or cidr per line.

```shell
cloudflare-utils --api-token <API Token with Account:Rule Lists:Edit> --account-id <account id> sync-list --list-name <list name> file://path/to/file.txt
```

#### URL

Use the `http://` or `https://` prefix to read a list from a URL. The URL should return a plain text response with one ip or cidr per line.

```shell
cloudflare-utils --api-token <API Token with Account:Rule Lists:Edit> --account-id <account id> sync-list --list-name <list name> https://example.com/list.txt
```

#### Preset

There are the following presets available:

#### - `cloudflare`

IPs that Cloudflare uses for their services. To include china ips add `?include=china` to the end of the preset. For example: `preset://cloudflare?include=china`

#### - `github`

IPs that GitHub uses for their services. Includes, web, hooks, api, git, packages, packages, actions, actions_macos. You can exclude any of these by adding `?exclude=` followed by a comma separated list of items to exclude. For example, to exclude actions and actions_macos: `preset://github?exclude=actions,actions_macos`

#### - `uptime-robot`

IPs that Uptime Robot uses for their services.

Example usage:
```shell
cloudflare-utils --api-token <API Token with Account:Rule Lists:Edit> --account-id <account id> sync-list --list-name <list name> preset://cloudflare
```

If you want to see a new preset added, please open an issue or a PR.

### Options

One of the source options must be provided and either `--list-id` or `--list-name` must be provided.

- `--list-id`: ID of the list you want to sync. If you supply a list id and not a list name and the list does not exist then it will return an error.
- `--list-name`: Name of the list you want to sync. If no list exists with that name then it will create a new list with that name.
- `--list-description`: Description of the list you want to create. Only used if the list does not exist and is being created.
- `--item-comment`: Comment to add to each item in the list. Default is "Added by cloudflare-utils"
- `--dry-run`: Output what would be changed without actually making any changes.
- `--no-comment`: Don't add a comment to each item in the list. Overrides `--comment`
- `--comment`: Comment to add to each item in the list. Default is "Added by cloudflare-utils"
- `--no-wait`: Do not wait for the list to be updated. By default, the command will wait for the list to be updated before exiting.
- `--source`: Source of the list. Can be `file://`, `http://`, `https://`, or `preset://`. It can also be supplied as the last argument without the `--source` flag.

#### Required API Permissions

[Token Quick Link](https://dash.cloudflare.com/profile/api-tokens?permissionGroupKeys=%5B%7B%22key%22%3A%22account_rule_lists%22%2C%22type%22%3A%22edit%22%7D%5D&name=Cloudflare+Utils%3A+List+Sync&accountId=*&zoneId=all)

- _Account:DNS:Edit_