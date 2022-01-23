{ pkgs ? import <nixpkgs> {}
}:

pkgs.mkShell {
	buildInputs = [
		pkgs.go_1_17
	];
}