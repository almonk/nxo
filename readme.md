# nxo
Bootstrap nix environments in seconds.

### Usage

Setup a project with ruby and go dependencies:

```
$ nxo install ruby go
```

Append a package to an existing `shell.nix` file:

```
$ nxo add sqlite
```

### Installation

You'll need to build `nxo` from source.

```
$ git clone [this repo]
$ go build -o dist/nxo
$ sudo ln -s [path to nxo]/dist/nxo /usr/local/bin # Symlink to make `nxo` available globally
```