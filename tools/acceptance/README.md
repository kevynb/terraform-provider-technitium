# Acceptance Tests (Local)

This directory contains a Docker Compose setup for a local Technitium DNS server
to run Terraform acceptance tests.

## Start the server

```sh
make acc-up
```

The UI will be available at:

```
http://localhost:5380
```

Note: the acceptance docker-compose only exposes port 5380; DNS port 53 is not
published to avoid conflicts in CI.

Default admin password (from docker-compose):

```
changeme
```

## Create an API token

In the Technitium UI, create an API token and export it:

```sh
export TECHNITIUM_API_URL="http://localhost:5380"
export TECHNITIUM_API_TOKEN="your-token"
export TF_ACC=1
```

If `TECHNITIUM_API_TOKEN` is not set, the acceptance tests will try to log in
using the admin credentials and fetch a token. You can override the defaults:

```sh
export TECHNITIUM_ADMIN_USER="admin"
export TECHNITIUM_ADMIN_PASSWORD="changeme"
```

You can also let `make acc-up` handle the login step. It will fetch a token and
store it in `tools/acceptance/token.env` for subsequent `make acc-test` runs.

If you are using a self-signed cert, also set:

```sh
export TECHNITIUM_SKIP_TLS_VERIFY=1
```

## Run acceptance tests

```sh
make acc-test
```

## Tear down

```sh
make acc-down
```
