package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"text/template"

	_ "github.com/gookit/color"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:      "nxo",
		Usage:     "Bootstrap nix environments in seconds",
		UsageText: "nxo [command] [nix packages...]",
		Commands: []*cli.Command{
			{
				Name:    "init",
				Aliases: []string{"i"},
				Usage:   "Initialise a new nix environment",
				Action: func(c *cli.Context) error {
					if passPreflight() != nil {
						// Exit with error
						return cli.Exit(passPreflight(), 1)
					}

					// Check a package is present in args
					if !c.Args().Present() {
						return cli.Exit("Specify at least 1 nix package...", 1)
					}

					// Check a shell.nix file exists...
					if _, err := os.Stat("./shell.nix"); err == nil {
						// If it does, read shell.nix and append new packages
						packages := readPackagesFromShellNix()
						packages = append(packages, c.Args().Slice()...)

						// Setup direnv
						writeDirenvToEnvrc()
						runAllowDirenv()

						writePackagesToShellNix(packages)
					} else if errors.Is(err, os.ErrNotExist) {
						// If it doesn't create from scratch

						// Setup direnv
						writeDirenvToEnvrc()
						runAllowDirenv()

						packages := c.Args().Slice()
						writePackagesToShellNix(packages)
					}

					return nil
				},
			},
			{
				Name:    "clean",
				Aliases: []string{"c"},
				Usage:   "Destroy shell.nix and .envrc",
				Action: func(c *cli.Context) error {
					// Declare managed files
					managedFiles := []string{"./shell.nix", "./.envrc"}

					// For each managed file, attempt to remove...
					for _, file := range managedFiles {
						e := os.Remove(file)
						if e != nil {
							log.Fatal(e)
						}
					}

					return nil
				},
			},
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))
	app.EnableBashCompletion = true

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

// Runs all preflight checks for pre-requisites to make sure nxo can run
// Returns: error
func passPreflight() error {
	if !doesDependencyExist("nix-shell") {
		return errors.New("nix-shell is not installed. Install nix first from https://nixos.org/download.html#nix-install-macos")
	}

	if !doesDependencyExist("direnv") {
		return errors.New("direnv is not installed. Install it with `brew install direnv`")
	}

	return nil
}

// Check whether a user binary e.g. `nix-shell`` exists
// Returns: bool
func doesDependencyExist(name string) bool {
	_, err := exec.LookPath(name)

	if err != nil {
		return false
	} else {
		return true
	}
}

// Creates the obligatory .envrc for direnv access
func writeDirenvToEnvrc() {
	var file = "use nix"
	os.WriteFile("./.envrc", []byte(file), 0644)
}

// Runs the `direnv allow` command
func runAllowDirenv() {
	cmd := exec.Command("direnv", "allow")
	_, err := cmd.Output()

	if err != nil {
		fmt.Println(err.Error())
		return
	}
}

// Writes an array of package names to shell.nix
func writePackagesToShellNix(packages []string) {
	const tpl = `{ pkgs ? import <nixpkgs> {}
}:

pkgs.mkShell {
	buildInputs = [{{range .Packages}}
		pkgs.{{.}}{{end}}
	];
}`

	data := struct {
		Packages []string
	}{
		Packages: packages,
	}

	t, err := template.New("shell").Parse(tpl)
	if err != nil {
		println(err.Error())
	}

	var tplOutput bytes.Buffer

	// Execute the template
	if err := t.Execute(&tplOutput, data); err != nil {
		println(err.Error())
	}

	// Write buffer
	os.WriteFile("./shell.nix", tplOutput.Bytes(), 0644)
}

// Reads the existing packages from shell.nix
// Returns: []string
func readPackagesFromShellNix() []string {
	shellNixContents, err := ioutil.ReadFile("./shell.nix")
	if err != nil {
		log.Fatal(err)
	}

	re := regexp.MustCompile(`(?s)buildInputs = \[\n(.*?)\]`)
	match := re.FindStringSubmatch(string(shellNixContents))

	if len(match) == 0 {
		// We can't find any packages in this file
		cli.Exit("Can't parse shell.nix. Have you manually edited it?", 1)
	}

	packages := strings.Split(match[1], "\n")
	packages = packages[:len(packages)-1] // Remove empty last element

	// Trim whitespace from the file
	for i := range packages {
		packages[i] = strings.TrimSpace(packages[i])

		// Strip out `pkgs.` because we add it again in the template
		packages[i] = strings.Replace(packages[i], "pkgs.", "", 1)
	}

	return packages
}
