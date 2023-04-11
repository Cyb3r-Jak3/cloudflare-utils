# Purge Pages Deployments

The purpose of this program is to offer a quick way to bulk remove all Cloudflare Pages deployments for a project.

**Required API Permissions**:
- _Account:Cloudflare Pages:Edit_

## Running

You need to pass the following flags to run this program:
- `--api-token`: Your API token with the required permissions.
- `--account-id`: Your account ID where the pages project is located.
- `--project-name`: The name of the pages project.

Optional flags:
- `--dry-run`: If you want to see what would be deleted without actually deleting anything.
- `--delete-project`: If you want to delete the Pages project after deleting all deployments.

Example: 
```shell
cloudflare-utils --api-token <API Token with Pages:Edit> --account-id <account ID> purge-deployments --project-name <project name>
```

**Note**:

I have not tested with 1000+ deployments and am not sure how the rate limit will take effect.
You should use the `--lots-of-deployments` flag if you have more than 1000 deployments as slow down the rate of listing deployments.