# Delete Branch Deployments

The purpose of this program is to offer a quick way to bulk remove Cloudflare Pages branch deployments that you don't want typically once your have merged a pull request.

**Required API Permissions**:

 - _Account:Cloudflare Pages:Edit_

## Running

You need to pass the following flags to run this program:

- `--api-token`: Your API token with the required permissions.
- `--account-id`: Your account ID where the pages project is located.
- `--project-name`: The name of the pages project.
- `--branch`: The alias you want to remove deployments from.

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