# Troubleshooting

## Getting Permission Errors

If you are getting permission errors for your api token then you can run with the `--debug` flag as it will check the permissions of the token and print them out.
Please note that the API token needs to have `API Token:Read` in order to be able to read the token permissions.

## Unable to run commands in docker containers

If you are running the commands in a docker container then you either need to run the container with interactive mode or pass the `--confirm` flag to skip the confirmation prompt.