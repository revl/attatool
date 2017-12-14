// Copyright (C) 2017 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

func generateAndBootstrapPackage(pkgSelection []string) error {
	packageIndex, err := buildPackageIndex()
	if err != nil {
		return err
	}

	pkgRootDir := filepath.Join(getWorkspaceDir(), "packages")

	type packageAndGenerator struct {
		pd         *packageDefinition
		packageDir string
		generator  func() error
	}

	var packagesAndGenerators []packageAndGenerator

	for _, packageName := range pkgSelection {
		pd, ok := packageIndex.packageByName[packageName]
		if !ok {
			return errors.New("no such package: " + packageName)
		}

		packageDir := filepath.Join(pkgRootDir, pd.packageName)

		generator, err := pd.getPackageGeneratorFunc(packageDir)
		if err != nil {
			return err
		}

		packagesAndGenerators = append(packagesAndGenerators,
			packageAndGenerator{pd, packageDir, generator})
	}

	optRegexp := regexp.MustCompile(`^--(enable|disable|with|without)` +
		`-([^\s\[=]+)([^\s]*)\s*(.*)$`)

	autoconfOptions := map[string]bool{
		"FEATURE":             true,
		"PACKAGE":             true,
		"aix-soname":          true,
		"dependency-tracking": true,
		"fast-install":        true,
		"gnu-ld":              true,
		"libtool-lock":        true,
		"option-checking":     true,
		"pic":                 true,
		"pkgconfigdir":        true,
		"shared":              true,
		"silent-rules":        true,
		"static":              true,
		"sysroot":             true,
	}

	for _, pg := range packagesAndGenerators {
		// Generate autoconf and automake sources for the package.
		if err = pg.generator(); err != nil {
			return err
		}

		configurePathname := filepath.Join(pg.packageDir, "configure")

		// Bootstrap the package if 'configure' does not exist.
		_, err = os.Lstat(configurePathname)
		if os.IsNotExist(err) {
			bootstrapCmd := exec.Command("./autogen.sh")
			bootstrapCmd.Dir = pg.packageDir
			if err = bootstrapCmd.Run(); err != nil {
				return errors.New(filepath.Join(pg.packageDir,
					"autogen.sh") + ": " + err.Error())
			}
		}

		configureHelpCmd := exec.Command(configurePathname, "--help")
		configureHelpStdout, err := configureHelpCmd.StdoutPipe()
		if err != nil {
			return err
		}
		if err = configureHelpCmd.Start(); err != nil {
			return err
		}
		helpScanner := bufio.NewScanner(configureHelpStdout)
		type optionOrFeature = struct {
			keyword     string
			option      string
			arg         string
			description string
		}
		var options []optionOrFeature
		var currentOption *optionOrFeature

		for helpScanner.Scan() {
			helpLine := strings.TrimRight(helpScanner.Text(), " ")

			if helpLine == "" ||
				!strings.HasPrefix(helpLine, " ") {
				if currentOption != nil {
					options = append(options, *currentOption)
					currentOption = nil
				}
				continue
			}

			helpLine = strings.TrimLeft(helpLine, " ")

			if strings.HasPrefix(helpLine, "--") {
				if currentOption != nil {
					options = append(options, *currentOption)
					currentOption = nil
				}
			} else {
				if currentOption != nil {
					if currentOption.description != "" {
						currentOption.description += " "
					}
					currentOption.description += helpLine
				}
				continue
			}

			matches := optRegexp.FindStringSubmatch(helpLine)

			if len(matches) > 4 {
				option := matches[2]
				if _, present := autoconfOptions[option]; present {
					continue
				}
				currentOption = &optionOrFeature{
					matches[1],
					option,
					matches[3],
					matches[4]}
			}
		}
		if err := helpScanner.Err(); err != nil {
			return err
		}
		if err = configureHelpCmd.Wait(); err != nil {
			return err
		}
		for _, opt := range options {
			fmt.Println(opt.keyword)
			fmt.Println(opt.option)
			fmt.Println(opt.arg)
			fmt.Println(opt.description)
		}
	}

	return nil
}

// SelectCmd represents the select command
var selectCmd = &cobra.Command{
	Use:   "select package_range...",
	Short: "Choose one or more packages to work on",
	Args:  cobra.MinimumNArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		if err := generateAndBootstrapPackage(args); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(selectCmd)

	selectCmd.Flags().SortFlags = false
	addQuietFlag(selectCmd)
	addWorkspaceDirFlag(selectCmd)
	addPkgPathFlag(selectCmd)
}
