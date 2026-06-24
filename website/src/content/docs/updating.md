---
title: Updating
description: How to update the Craftybase CLI to the latest version.
---

How you update depends on how you installed the CLI.

## Self-update (script / manual installs)

If you installed with the `install` script or a downloaded binary, update in place:

```sh
craftybase update
```

This downloads the latest release for your platform, verifies its checksum, and
replaces the running binary. To see whether a newer version exists without
installing it:

```sh
craftybase update --check
```

## Homebrew

If you installed via Homebrew, use brew so its bookkeeping stays in sync:

```sh
brew upgrade craftybase
```

`craftybase update` detects a Homebrew install and will point you here.

## Windows

Download the latest release `.zip` from the
[releases page](https://github.com/craftybase/craftybase-cli/releases) and replace
the binary on your `PATH`.

## go install

```sh
go install github.com/craftybase/craftybase-cli/cmd/craftybase@latest
```
