# Sync List

The purpose of this command is to offer a way to sync a list of ips fetched from a URL with a Cloudflare list. Which can then be used in firewall rules, rate limiting rules, or other places that support lists.

There are currently 3 supported sources:

- file
- url
- preset

You can supply either the list name or list id to identify the list you want to sync. If you supply a list name then it will look for an existing list with that name and use it. If no list exists with that name then it will create a new list with that name.

If you supply a list id and not a list name and the list does not exist then it will return an error.

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

- `cloudflare`
    List of ips that Cloudflare uses for their services.
- `cloudflare-china`
    List of ips that Cloudflare uses for their services including China.
- `github`
    List of ips that GitHub uses for their services. Includes, web, hooks, api, git, packages, packages, actions, actions_macos.
- `uptime-robot`
    List of ips that Uptime Robot uses for their services.

```shell
cloudflare-utils --api-token <API Token with Account:Rule Lists:Edit> --account-id <account id> sync-list --list-name <list name> preset://cloudflare
```

If you want to see a new preset added, please open an issue or a PR.

#### Required API Permissions

[Token Quick Link](https://dash.cloudflare.com/profile/api-tokens?permissionGroupKeys=%5B%7B%22key%22%3A%22account_rule_lists%22%2C%22type%22%3A%22edit%22%7D%5D&name=Cloudflare+Utils%3A+List+Sync&accountId=*&zoneId=all)

- _Account:DNS:Edit_