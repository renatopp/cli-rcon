# go-rcon

CLI for RCON.

## Installation

```bash
go install github.com/renatopp/go-rcon@latest
```

## Usage

```bash
Usage: rcon [options] <address> [<commands>]

Options:
  -p, --pass         Server RCON password
  -h, --help         Show help message
  --no-autocomplete  Disable autocomplete

Arguments:
  address           (required) The host and port of the RCON server
  commands          The RCON commands to execute
```

Examples:

- Send an RCON:

  `rcon -p password 127.0.0.1:27015 "say Hello, world!"`

- Open RCON REPL:

  `rcon -p password 127.0.0.1:27015`
