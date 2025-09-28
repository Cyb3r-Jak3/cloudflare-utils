# GitHub Actions

I have created a first-party action to install and run cloudflare-utils in your workflows. You can find it on the [GitHub Marketplace](https://github.com/marketplace/actions/cloudflare-utils)

## Usage

Add a step in your workflow to use the action. Below is an example of using the action to run `dns-cleaner`.

```yaml
- name: Cloudflare Utils - DNS Cleaner
  uses: Cyb3r-Jak3/actions-cloudflare-utils@v1
  with:
     version: 'latest' # Optional, defaults to latest
     args: '--api-token ${{ secrets.CLOUDFLARE_API_TOKEN }} --zone-name example.com dns-purge --confirm'
```

If you only want to install the action and not run a command then you don't need to pass any args.

```yaml
- name: Cloudflare Utils - Install
  uses: Cyb3r-Jak3/actions-cloudflare-utils@v1
  with:
     version: 'latest' # Optional, defaults to latest
```


## Examples

### List Sync

```yaml
name: Sync IPs


on:
  schedule:
    - cron: '0 0 * * *' # Runs every day at midnight
  workflow_dispatch:


jobs:
  sync-ips:
    runs-on: ubuntu-latest
    steps:
      - name: Install Cloudflare Utils
        uses: Cyb3r-Jak3/actions-cloudflare-utils@v1
        with:
          version: 'latest'
          args: sync-list --list-name github_actions_ips --source preset://github
        env:
          CLOUDFLARE_API_TOKEN: ${{ secrets.CLOUDFLARE_API_TOKEN }}
          CLOUDFLARE_ACCOUNT_ID: ${{ secrets.CLOUDFLARE_ACCOUNT_ID }}
```
From [here](https://github.com/Cyb3r-Jak3/cloudflare-util-github-syncer/blob/main/.github/workflows/run.yml)


### Pages Prune

```yaml
name: Clear Pages Deployments on PR Close
on:
  pull_request:
    types: [ closed ]

jobs:
  clear-caches:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      pull-requests: read
      actions: write
    steps:
      - name: Clear Page Deployments
        uses: Cyb3r-Jak3/actions-cloudflare-utils@v1
        if: ${{ github.event.pull_request.merged == true }}
        with:
          args: prune-deployments --project-name cloudflare-utils --branch ${{ github.event.pull_request.head.ref }}
        env:
          CLOUDFLARE_API_TOKEN: ${{ secrets.CLOUDFLARE_CACHE_CLEANER }}
          CLOUDFLARE_ACCOUNT_ID: ${{ secrets.CLOUDFLARE_ACCOUNT_ID }}
```

From [here](https://github.com/Cyb3r-Jak3/cloudflare-util-github-syncer/blob/main/.github/workflows/run.yml)