# Dina CLI

Command line tool for the Dina platform — deploy apps, manage environments, and more.

## Installation

### Homebrew (macOS and Linux)

```sh
brew install dinacomputer/tap/dina
```

### Shell script (macOS and Linux)

```sh
curl -sSL https://raw.githubusercontent.com/dinacomputer/cli/main/install.sh | sh
```

To install to a custom directory:

```sh
INSTALL_DIR=~/.local/bin curl -sSL https://raw.githubusercontent.com/dinacomputer/cli/main/install.sh | sh
```

To install a specific version:

```sh
VERSION=0.1.0 curl -sSL https://raw.githubusercontent.com/dinacomputer/cli/main/install.sh | sh
```

### Go

```sh
go install github.com/dinacomputer/cli/cmd/dina@latest
```

### Manual download

Download the latest release for your platform from
[GitHub Releases](https://github.com/dinacomputer/cli/releases), extract it,
and place the `dina` binary somewhere on your `PATH`.

### Linux packages

`.deb` and `.rpm` packages are available on each
[GitHub Release](https://github.com/dinacomputer/cli/releases).

```sh
# Debian / Ubuntu
sudo dpkg -i dina_*.deb

# Fedora / RHEL
sudo rpm -i dina_*.rpm
```

## Verify installation

```sh
dina version
```

## Usage

```sh
dina --help
```
