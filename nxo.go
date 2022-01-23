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

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

func main() {

	app := &cli.App{
		Name:      "nxo",
		Usage:     "Bootstrap nix environments in seconds",
		UsageText: "nxo [command] [nix packages...]",
		Commands: []*cli.Command{
			{
				Name:      "install",
				Aliases:   []string{"i"},
				Usage:     "Install nix packages for the working directory",
				UsageText: "nxo install [package]",
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

					color.Green("✓ Installed %d packages:", c.Args().Len())

					for _, arg := range c.Args().Slice() {
						color.Green("    - %s", arg)
					}

					return nil
				},
			},
			{
				Name:      "clean",
				Aliases:   []string{"c"},
				Usage:     "Destroy shell.nix and .envrc",
				UsageText: "nxo clean",
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
			{
				Name:      "search",
				Aliases:   []string{"s"},
				Usage:     "Search the nix package registry",
				UsageText: "nxo search [package]",
				Action: func(c *cli.Context) error {
					// Check a package is present in args
					if !c.Args().Present() {
						return cli.Exit("Specify at least 1 nix package...", 1)
					}

					// Declare managed files
					query := c.Args().First()
					url := fmt.Sprintf("https://search.nixos.org/packages?channel=21.11&from=0&size=50&sort=relevance&type=packages&query=%s", query)

					// Open the URL in the web browser
					cmd := exec.Command("open", url)
					_, err := cmd.Output()

					if err != nil {
						fmt.Println(err.Error())
					}

					return nil
				},
			},
			{
				Name:      "replace",
				Aliases:   []string{"r"},
				Usage:     "Replace a nix package in shell.nix with another",
				UsageText: "nxo replace [original package] [target package]",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "invert",
						Aliases: []string{"i"},
						Value:   false,
						Usage:   "Inverts the replace order of packages: [target package] ← [original package]",
					},
				},
				Action: func(c *cli.Context) error {
					// Check preflight...
					if passPreflight() != nil {
						// Exit with error
						return cli.Exit(passPreflight(), 1)
					}

					// Package to be replaced
					originalPackage := c.Args().First()

					// Package to replace with...
					newPackage := c.Args().Get(1)

					// If invert is used, then swap the packages
					if c.Bool("invert") {
						originalPackage, newPackage = newPackage, originalPackage
					}

					// Read existing packages in
					packages := readPackagesFromShellNix()

					anyMatches := false

					for i := range packages {
						if packages[i] == originalPackage {
							packages[i] = newPackage
							anyMatches = true
						}
					}

					// We looped through the whole shell.nix but didn't find any packages to replace...
					if !anyMatches {
						return cli.Exit(fmt.Sprintf("`%s` not found in shell.nix", originalPackage), 1)
					}

					// Write new shell.nix
					writePackagesToShellNix(packages)

					color.Green(fmt.Sprintf("✓ Replacing %s → %s", originalPackage, newPackage))

					return nil
				},
			},
			{
				Name:      "remove",
				Aliases:   []string{"rm"},
				Usage:     "Remove a package from shell.nix",
				UsageText: "nxo remove [package]",
				Action: func(c *cli.Context) error {
					// Check preflight...
					if passPreflight() != nil {
						// Exit with error
						return cli.Exit(passPreflight(), 1)
					}

					// Check a package is present in args
					if !c.Args().Present() {
						return cli.Exit("Specify at least 1 nix package...", 1)
					}

					// Read existing packages in
					packages := readPackagesFromShellNix()

					// Remove each package from arguments
					for _, item := range c.Args().Slice() {
						packages = removeItemFromPackages(packages, item)
						color.Green(fmt.Sprintf("✓ Removing %s", item))
					}

					// Write new shell.nix
					writePackagesToShellNix(packages)

					return nil
				},
			},
		},
	}
	app.UseShortOptionHandling = true
	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

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

// Remove an element from a packages array
func removeItemFromPackages(packages []string, item string) []string {
	newitems := []string{}

	for _, i := range packages {
		if i != item {
			newitems = append(newitems, i)
		}
	}

	return newitems
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
