# Prune Pages Deployments

The purpose of this command is to offer a quick way to bulk remove Cloudflare Pages deployments.

There are three ways to remove deployments:

- Deleting all deployments for a branch.
- Deleting all deployments before a certain time.
- Deleting all deployments after a certain time.

If you want to delete all deployments for a project, check out the [purge deployments](purge-deployments.md) command.

## Running

You need to pass the following flags to run this program:

- `--api-token`: Your API token with the required permissions.
- `--account-id`: Your account ID where the pages project is located.
- `--project`: Name of the pages project.

You need to pass only one of the following flags:

- `--branch`: Alias you want to remove deployments from.
- `--before`: Date you want to remove deployments before. Format: `YYYY-MM-DDTHH:mm:ss`. Example: `2021-01-01T00:00:00` = January 1st, 2021 at 12:00:00 AM.
- `--after`: Date you want to remove deployments after. Format: `YYYY-MM-DDTHH:mm:ss`. Example: `2021-01-01T00:00:00` = January 1st, 2021 at 12:00:00 AM.

[//]: # (- `--time`: Shortcut for deleting based on time. Use the format of `1<unit>` where unit is one of y &#40;year&#41;, M &#40;month&#41;, w &#40;week&#41;, d &#40;day&#41;, h &#40;hour&#41;, m &#40;minute&#41;, s &#40;second&#41;. To delete all deployments older than an hours use `1h`. For more into refer to [time-shortcut]&#40;#time-shortcut&#41;.)

Optional flags:

- `--dry-run`: See what would be deleted without actually deleting anything.
- `--lots-of-deployments`: Useful if there are more than 1000 deployments, this will slow down the rate of listing deployments.
- `--force`: Forces the deletes of deployments.

Example:

```shell
cloudflare-utils --api-token <API Token with Pages:Edit> --account-id <account ID> prune-deployments --project-name <project name> --branch <branch>
```

[//]: # (### Time Shortcut)

[//]: # ()
[//]: # (By using the `--time` flag, you can delete deployments based on time from when they were created. This is useful if you want to delete all deployments older than a certain amount of time.)

[//]: # ()
[//]: # (Example:  )

[//]: # ()
[//]: # (To delete all deployments older than 1 month, use the following command:)

[//]: # (```shell)

[//]: # (cloudflare-utils --api-token <API Token with Pages:Edit> --account-id <account ID> prune-deployments --project-name <project name> --time 1M)

[//]: # (```)

???+ warning

    I have only tested this with a project with 20,000 deployments. While doing so, it was able to delete all deployments even though some throw errors.
    It will take a while to run with a lot of deployments so be patient.

#### Required API Permissions

- _Account:Cloudflare Pages:Edit_

[Token Quick Link](https://dash.cloudflare.com/profile/api-tokens?permissionGroupKeys=%5B%7B%22key%22%3A%22page%22%2C%22type%22%3A%22edit%22%7D%5D&name=Cloudflare+Utils%3A+Page+Write)
