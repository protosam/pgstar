# PGStar
A server for building transactional HTTP APIs with Postgres and Starlark.

## About PGStar
PGStar is a simple toolkit that can enable HTTP API development.

As a primary goal PGStar is intended to abstract your application away from Postgres operations.
As a configuration, Starlark seemed like a great language choice.
The language is easily embeddable with `starlark-go`, provides no hidden means of IO, and was indended to be a configuration language.

The workflow for PGStar is simple.
- Create a starlark script to do something.
- Update config.star with a route to the new script.

The safeties guaranteed are as follows.
- All database interactions happen in a transaction.
- All failures result in transaction rollback.
- Your data stays consistent.

## Getting Started
Instructions can be found in the `docs/` directory.
- [Installation](docs/Installation.md)
- [Hello World Example](docs/HelloWorld.md)
- [Module Details](docs/Modules.md)

## Design Choices
- Starlark was chosen because it is a configuration language and resembles Python.
- Functions shall panic only when a system issue occurs and pass all other errors back to be handled in a tuple.
- Built-in modules should only support

## Development Plan Notes
These are things slated to be added to PGStar.

- Implement CLI tooling to simply scripts without web service
- Implement package management system (using git repos)
- - Versioning structure
- - Caching mechanisms
- Implement schema management tooling (with plain sql files)
- Implement logging module
- - Default log levels: DEBUG, INFO, ERROR, WARNING, DEPRECATED
- - Custom log level support
- - Override print's default level from INFO
- - Output to STDOUT or log file.
- Identify and implement strategy for test coverage
