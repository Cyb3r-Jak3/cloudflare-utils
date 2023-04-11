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

Example:

```shell
cloudflare-utils --api-token <API Token with Pages:Edit> --account-id <account ID> prune-deployments --project-name <project name> --branch <branch>
```

**Note**:

I have not tested with 1000+ deployments and am not sure how the rate limit will take effect.
You should use the `--lots-of-deployments` flag if you have more than 1000 deployments as slow down the rate of listing deployments.