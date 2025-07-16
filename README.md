# PaperMC Manager TUI

A TUI application to manage [PaperMC](https://papermc.io/downloads/paper) versions.


It allows you to check the latest stable version and build as well as downloading the latest stable build.

When downloading a new version, it will prompt for a backup so that you can easily revert back to the previous version.

Logs are created to keep track of the latest downloaded version and to follow the actions being taken.

### Usage

Download the latest version from releases;

Binaries are built and released automatically for Windows (amd64 & arm64), Linux (amd64 & arm64) and MacOS (amd64).

Simply run the binary and follow along the TUI.

### Building from source

Using the Makefile is the easiest way:

```bash
make build
```

It will create a binary called `paper-mc-tui` in project root folder.

There is nothing complicated in the build command so feel free to tweak it as you like.
