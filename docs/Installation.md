# Installing PGStar
You will need a Postgres database and a built copy of `pgstard` (build instructions are below).

If you have a [kind](https://kind.sigs.k8s.io/) with at least 3 workers running, you can deploy the yugabyte helm chart and forward the service port.
```shell
# Install or upgrade the yugabyte chart
helm upgrade --install --namespace yugabyte yugabyte yugabyte --repo https://charts.yugabyte.com

# Expose the port locally for development
kubectl port-forward -n yugabyte svc/yb-tserver-service 5433:5433
```

An environment variable must be set for `pgstard`.
```shell
export PGSTAR_POSTGRES_CONFIG="host=localhost port=5433 user=yugabyte password=yugabyte database=mydatabase sslmode=disable"
```

Now you should be prepared to check out the hello world example [here](HelloWorld.md).

## Environment Variables
Here is the full list of environment variables available for the `pgstar` command.
- `PGSTAR_BIND_ADDR` - Network interface and port to listen on.
- `PGSTAR_POSTGRES_CONFIG` - Database connection arguments.
- `PGSTAR_SSL_CERTIFICATE` - Optional, provides SSL certificate path to enable SSL.
- `PGSTAR_SSL_PRIVATE_KEY` - Optional, provides SSL key path to enable SSL.



## Building
This project requires Go to be built.
```
go build -o pgstar github.com/protosam/pgstar/server
```

## Docker Containers
TODO: Release `amd64` and `aarch64` containers for `v0.1.0-alpha`.
