# Prune Pages Deployments

**Required API Permissions**:

- _Account:Cloudflare Pages:Edit_


The purpose of this program is to offer a quick way to bulk remove Cloudflare Pages deployments.

There are three ways to remove deployments:

- Deleting all deployments for a branch.
- Deleting all deployments before a certain time.
- Deleting all deployments after a certain time.

If you want to delete all deployments for a project, check out [purge deployments](purge-deployments.md) command.

## Running

You need to pass the following flags to run this program:

- `--api-token`: Your API token with the required permissions.
- `--account-id`: Your account ID where the pages project is located.
- `--project-name`: The name of the pages project.

You need to pass one of the following flags:

- `--branch`: The alias you want to remove deployments from.
- `--before`: The date you want to remove deployments before. Format: `YYYY-MM-DDTHH:mm:ss`.
- `--after`: The date you want to remove deployments after. Format: `YYYY-MM-DDTHH:mm:ss`.

Optional flags:

- `--dry-run`: If you want to see what would be deleted without actually deleting anything.
- `--lots-of-deployments`: If you have more than 1000 deployments, this will slow down the rate of listing deployments.

Example:

```shell
cloudflare-utils --api-token <API Token with Pages:Edit> --account-id <account ID> prune-deployments --project-name <project name> --branch <branch>
```

???+ warning

    I have only tested this with a project with 20,000 deployments. While doing so, it was able to delete all deployments even though some throw errors.
    It will take a while to run with a lot of deployments so be patient.