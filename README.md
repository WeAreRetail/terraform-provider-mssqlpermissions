# Terraform Provider Microsoft SLQ Server Permissions Provider

MSSQL Permissions manages logins, users, roles, and permissions in a Microsoft SQL Server.
It can handle both on premise servers and Azure Databases.

## Requirements

1. Docker
2. Docker Compose
3. All other requirements are bundled in the Dev Container.

Note: the Dev Container uses SQL Server 2022 for Linux. On Apple Silicon, use [Rosetta for x86/AMD64 emulation]([url](https://docs.docker.com/desktop/settings/mac/#use-rosetta-for-x86amd64-emulation-on-apple-silicon)).

## Dev Container

Open the repository with VS Code. It should show a popup window to start the container.

If it is not the case, hit ctrl+shift+p, type `Dev Containers: ` and choose the option to reopen in the container.

## Building The Provider

1. Clone the repository
1. Start the Dev Container
1. Build the provider using the Go `install` command:

```shell
go install .
```

## Using the provider

Fill this in for each provider

## Developing the Provider

If you wish to work on the provider, you'll first need to start the Dev Container (see [Requirements](#requirements) above).

To compile the provider, run `go install .`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate`.

```shell
go generate
```

## Tasks

The code comes with a list tasks to help testing. The tasks use [Task](https://taskfile.dev/).

You can run the task using command line or the Visual Studio Code addon bundled in the dev container.

```shell
task terraform-up
```
