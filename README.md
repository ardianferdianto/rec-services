# Reconciliation Service
**Description:** Responsible in reconciliate system generated transaction with bank statements
# Seq diagramsequenceDiagram
# Architecture
## Layering
This is the overview of this repository architecture layer

`cmd/server/main` ↔ `internal/presenter` ↔ `internal/usecase` and `internal/domain` ↔ `internal/infrastructure/` ↔ `Database/Message Broker/APIs`

### Main
This folder is the entrypoint of this repository. Main initialize these things listed below:
* config
* logger
* migrator
* server

### Presenter
This section manages the presentation layer of this repository. We put console, migrator, and server code here. We also put API communication mechanism codes here (example: REST, RPC, GraphQL).

### Usecase and Model
Usecase and model contain business logic of this repository.

### Infrastructure
Infrasturcture layer contains infrastructure code needed for business logic purposes. This may include code such as initiating external API calls, SQL store, PubSub, etc.

# Development Environment Setup
This section will guide on how to setup this repository on your local machine. You can also add how to solve some problems during setup process on the troubleshooting section.

## Initial Setup
### Initialization
Run this command to download external dependencies and download required package.
```
make init
```

### Docker
We are using docker to initialize our dependencies such as postgres, redis, etc. To run docker, use this command.
```
make deps-up
``` 

To turn off the dependencies from docker, use the command below.
```
make deps-down
```

## Running the Repo on Local
We have 2 ways to run the repo, using the manual reload and hot reload approach. **Please note that transaction-management-service require PUBSUB to run, so either you need to run "make deps-up" or have your own built-in PUBSUB image**.

To do manual approach, use this command.
```
make run
```
This will automatically trigger the `make compile` command to compile the binary.

To do hot reload approach, use this command.
```
make air-http
```
The config to run air is located in `.dev/http.air.toml`. You can change the config if needed on that toml file.

## Migrate
For local database, we can migrate the schema by using this.
```
make migrate
```
It will run `make compile` and execute the migrate command. **Make sure that the database connection on `config.yaml` is correct and can connect to your local database**.

### Generate Mock
For unit test purposes, we need to generate mock. To generate all required mock, use this.
```
make generate-mock
```

## Unit Test
To run all unit tests, you can use this command below.
```
make test
```

## Troubleshooting
Note that sometimes when you already install some of the dependencies, the binary is not located on the $PATH variable yet which makes it not accesible globally. To fix this, do these steps:

1. Find out your default shell configuration file (`.bashrc`, `.zshrc`, etc)
2. Add this command
```
export GO111MODULE = on
export PATH=$PATH:/home/[your_username]/[your_go_path]/bin
export GOPRIVATE = gitlab.com
```
3. Restart your terminal

`export GO111MODULE=on` is a command that sets the GO111MODULE environment variable to the value `on`. This variable controls how Go manages dependencies in your projects.

`export PATH=$PATH:/home/[your_username]/[your_go_path]/bin` is used to add your custom go binary into the PATH variable to make it accessible globally. Please adjust the PATH accordingly with your own Go folder location.

`export GOPRIVATE = gitlab.com` is to make imports using SSH instead of regular HTTP from gitlab.com.


