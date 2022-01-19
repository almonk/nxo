{ pkgs ? import <nixpkgs> {}
}:

pkgs.mkShell {
  buildInputs = [
    pkgs.ruby
    pkgs.bundix
    pkgs.bundler
  ];
}

