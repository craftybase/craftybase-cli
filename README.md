# Stocksmith CLI

The official command-line interface for the [Stocksmith](https://stocksmith.com) Public API — manage your inventory from the terminal.

Full docs: **https://cli.stocksmith.dev**

## Install

```sh
curl -fsSL https://cli.stocksmith.dev/install | bash
```

Or with Homebrew:

```sh
brew install craftybase/tap/stocksmith
```

Or with Go:

```sh
go install github.com/craftybase/stocksmith-cli/cmd/stocksmith@latest
```

## Quickstart

```sh
stocksmith auth login        # paste your API key
stocksmith materials list    # list materials
stocksmith materials list --json | jq '.materials[].name'
```

See [Getting Started](https://cli.stocksmith.dev/getting-started/) for more.

## License

[MIT](LICENSE)
