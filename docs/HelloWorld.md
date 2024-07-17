# PGStar - Hello World
The steps below are a quick intro to help you get started.
There are many things you can accomplish with the [modules](Modules.md) available.

First you need a configuration file to setup routes, here is an example `config.star` file.
```starlark
# variables can be configured globally for all routes to consume
setglobal("foo", "bar")

# environment variables prefixed with "PGSTAR_ENV" can be read
some_environment_var = getenv("MYVAR", None)


# routes must be configured with allowed request types, url path, and script
addroute([ "GET" ], "/ping", "ping.star")

# routes support reading variables from path
# this route has variables "name" and "misc"
addroute([ "GET" ], "/hellodb/{name}/{misc:.*}", "path_reader.star")
```

Here is the contents of `ping.star`.
```starlark
# this module is used for http response related tasks
load("pgstar/http", http="exports")

# simply return a 201 status and write the hello world string
# all output is json encoded, once this is completed execution terminates
print(http.write(201, "Hello, World!"))
```

Here is the contents of `path_reader.star`.
```starlark
# this module is used for http response related tasks
load("pgstar/http", http="exports")

# get the value of name and path
vars = http.vars()
name = vars["name"]
path = vars["path"]


# simply return a 201 status and write the name
print(http.write(201, name)
```

Finally start PGStar:
```shell
pgstar ./config.star
```
