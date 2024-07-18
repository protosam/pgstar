# PGStar - Hello World
The steps below are a quick intro to help you get started.
There are many things you can accomplish with the [modules](Modules.md) available.

First you need a configuration file to setup routes, here is an example `config.star` file.
```starlark
# variables can be configured globally for all routes to consume
setGlobal("foo", "bar")

# environment variables prefixed with "PGSTAR_ENV" can be read
some_environment_var = getEnv("MYVAR", None)


# routes must be configured with allowed request types, url path, and script
addRoute([ "GET" ], "/ping", "ping.star")

# routes support reading variables from path
# this route has variables "name" and "misc"
addRoute([ "GET" ], "/hellodb/{name}/{misc:.*}", "path_reader.star")
```

Here is the contents of `ping.star`.
```starlark
# this module is used for http response related tasks
load("pgstar/http", http="exports")

# simply return a 200 status and write the hello world string
# all output is json encoded, once this is completed execution terminates
http.write(200, "Hello, World!")
```

Here is the contents of `path_reader.star`.
```starlark
# this module is used for http response related tasks
load("pgstar/http", http="exports")

# get the value of name and path
vars = http.vars()
name = vars["name"]
path = vars["misc"]

# simply return a 200 status and write the name
http.write(200, name)
```

For testing, `pgstar exec` performs evaluation for a path in your CLI.
```shell
pgstar exec ./config.star GET /hellodb/bob/extra/pathing/here
```

When ready to deploy, you will want to run the server to process HTTP requests.
```shell
pgstar server ./config.star
```
