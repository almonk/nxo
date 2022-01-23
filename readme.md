# nxo
Bootstrap nix environments in seconds.

### Usage

Setup a project with ruby and go dependencies:

```
$ nxo install ruby go
```

Let's add sqlite to this project as well. Simply run `nxo install` again:

```
$ nxo i sqlite # `nxo i` is a shortcut for install
```

`nxo` will generate the `shell.nix` below;

```nix
{ pkgs ? import <nixpkgs> {}
}:

pkgs.mkShell {
  buildInputs = [
    pkgs.ruby
    pkgs.go
    pkgs.sqlite
  ];
}
```

To wipe nix config:

```
$ nxo clean
```

or 

```
$ nxo c
```

### Installation

```bash
curl -L https://github.com/almonk/nxo/releases/download/v0.0.4/nxo_darwin_arm64 --output /tmp/nxo && chmod +x /tmp/nxo && sudo ln -s /tmp/nxo /usr/local/bin
```

### Uninstallation

```
rm -rf /usr/local/bin/nxo
```

### Build from source

You'll need to build `nxo` from source.

```
$ git clone [this repo]
$ go build -o dist/nxo
$ sudo ln -s [path to nxo]/dist/nxo /usr/local/bin # Symlink to make `nxo` available globally
```