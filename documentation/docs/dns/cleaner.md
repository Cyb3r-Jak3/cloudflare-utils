# DNS Cleaner

The purpose of DNS cleaner is to offer a quick way to bulk remove DNS records that you don't want.  
If you are looking to remove all DNS records then check out the [dns-purge feature](purge.md)


## Running

Clear mode operates in two steps, one to download the records and another to make any requested changes.

#### Default Behavior

When you run dns-cleaner command it will download dns records if a DNS file does not exist and apply changes if a DNS file exists.
You can choose either downloading or uploading with the `download` and `upload` sub-commands

### 1. Download

When you download DNS records, a file called `dns-records.yml` will be created with contains your DNS records.  
If you want a different file name then add `--dns-file` with the name of the file you want.

##### Download options

`--no-keep`: Changes the default keep value to false

`--quick-clean`: Looks through all DNS records for ones that are numeric values and sets those to be removed.

???+ note 
    Using `--no-keep` with `--quick-clean` is not supported.

### 2. Edit your records

Open the newly created file and any record that you do not want to keep change `keep:` to false. Do not delete records you want to remove, _only change `keep:` to false_

### 3. Apply your changes

Once you have made all the changes you need to apply the changes. You can do this by either running just `dns-cleaner` command or `dns-cleaner upload` command.

##### Upload options

`--dry-run`: See what would be deleted without actually deleting anything.

`--remove-file`: Remove the DNS file after uploading.

!!! note
  * If you changed the name of the file via the flag then you need to point to the same file
  * Once a DNS record is deleted, then it is gone. This program does not recreate records

#### Required API Permissions

[Token Quick Link](https://dash.cloudflare.com/profile/api-tokens?permissionGroupKeys=%5B%7B%22key%22%3A%22dns%22%2C%22type%22%3A%22edit%22%7D%5D&name=Cloudflare+Utils%3A+DNS+Write)

- _Zone:DNS:Edit_

