# Craftybase CLI

The official command-line interface for the [Craftybase](https://craftybase.com) Public API — manage your inventory from the terminal.

Full docs: **https://cli.craftybase.dev**

## Install

```sh
curl -fsSL https://cli.craftybase.dev/install | bash
```

Or with Homebrew:

```sh
brew install craftybase/tap/craftybase
```

Or with Go:

```sh
go install github.com/craftybase/craftybase-cli/cmd/craftybase@latest
```

## Quickstart

```sh
craftybase auth login        # paste your API key
craftybase materials list    # list materials
craftybase materials list --json | jq '.materials[].name'
```

See [Getting Started](https://cli.craftybase.dev/getting-started/) for more.

## License

[MIT](LICENSE)
