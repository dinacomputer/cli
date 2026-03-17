# Dina CLI

Go CLI for the Dina platform (`github.com/dinacomputer/cli`).

## Build & Run

```sh
task build        # builds ./dina
task install      # installs to $GOPATH/bin
go vet ./...      # lint
```

## Keeping the CLI up to date

Before running any `dina` commands, check if the CLI is current:

```sh
dina check-update
```

If an update is available, install it with:

```sh
dina update
```

Always ensure you are running the latest version before deploying or managing apps.
