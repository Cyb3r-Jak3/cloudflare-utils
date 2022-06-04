# DNS Cleaner

The purpose of DNS cleaner is to offer a quick way to bulk remove DNS records that you don't want. 

**Required API Permissions**: _DNS:Edit_


It operates in two steps, one to download the records and another to make any requested changes.

### 1. Download

To download you current records run `cloudflare-utils <auth> --zone-name <ZONE> dns-cleaner` where `<ZONE>` is your domain ie `example.com`.  
This will create a file called `dns-records.yml` with contains your DNS records.  
If you want a different file name then add `--dns-file` with the name of the file you want. It needs to end in either `.yml` or `.yaml`

### 2. Edit your records

Open the newly created file and any record that you do not want to keep change `keep:` to false.

### 3. Apply your changes

Once you have made all the changes you need re-run the command you used to download. 

Notes:
  * If you changed the name of the file then you need to point to the same file
  * Once a DNS record is deleted, then it is gone. This program does not recreate records