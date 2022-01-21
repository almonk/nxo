# nxo
Bootstrap nix environments in seconds.

### Usage

Setup a project with ruby and go dependencies:

```
$ nxo install ruby go
```

To add other nix packages, you can use `nxo install` again:

```
$ nxo install sqlite
```

To wipe nix config setup by `nxo`:

```
$ nxo clean
```

### Installation

You'll need to build `nxo` from source.

```
$ git clone [this repo]
$ go build -o dist/nxo
$ sudo ln -s [path to nxo]/dist/nxo /usr/local/bin # Symlink to make `nxo` available globally
```