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

## Building
Build with Go locally.
```shell
go build -o pgstar github.com/protosam/pgstar/cli
```

Building a container.
```shell
docker build . -t <TAGNAME_HERE>
```

## Docker Containers
Docker containers are available and tagged by version.
```shell
docker pull ghcr.io/protosam/pgstar:v0.0.2-alpha
```
