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