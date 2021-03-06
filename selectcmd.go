// Copyright (C) 2017, 2018 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"log"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"
)

func getRequired(pd *packageDefinition) packageDefinitionList {
	return pd.required
}

func getDependent(pd *packageDefinition) packageDefinitionList {
	return pd.dependent
}

func applyToSubtree(action func(*packageDefinition),
	root *packageDefinition,
	direction func(*packageDefinition) packageDefinitionList) {

	queue := packageDefinitionList{root}

	for {
		pd := queue[0]
		queue = queue[1:]

		action(pd)

		queue = append(queue, direction(pd)...)

		if len(queue) == 0 {
			break
		}
	}
}

func packageRangesToFlatSelection(pi *packageIndex, args []string) (
	packageDefinitionList, error) {
	selected := make(map[string]bool)
	marked := make(map[string]int)

	inclusion := true

	selectPackage := func(pd *packageDefinition) {
		selected[pd.PackageName] = inclusion
	}

	mark := 0

	markPackage := func(pd *packageDefinition) {
		marked[pd.PackageName] = mark
	}

	selectIfMarked := func(pd *packageDefinition) {
		if marked[pd.PackageName] == mark {
			selected[pd.PackageName] = inclusion
		}
	}

	for _, arg := range args {
		if arg == "+" {
			inclusion = true
			continue
		}

		if arg == "-" {
			inclusion = false
			continue
		}

		var pkgRange packageDefinitionList

		emptyRange := true

		for _, pkgName := range strings.SplitN(arg, ":", 2) {
			var pd *packageDefinition
			if pkgName != "" {
				var err error
				pd, err = pi.getPackageByName(pkgName)
				if err != nil {
					return nil, err
				}
				emptyRange = false
			}
			pkgRange = append(pkgRange, pd)
		}

		if emptyRange {
			continue
		}

		if len(pkgRange) == 1 {
			selected[arg] = inclusion
			continue
		}

		from, to := pkgRange[0], pkgRange[1]

		if from == nil {
			applyToSubtree(selectPackage, to, getRequired)
		} else if to == nil {
			applyToSubtree(selectPackage, from, getDependent)
		} else {
			mark++

			applyToSubtree(markPackage, to, getRequired)

			applyToSubtree(selectIfMarked, from, getDependent)
		}
	}

	var selection packageDefinitionList

	for _, pd := range pi.orderedPackages {
		if selected[pd.PackageName] {
			selection = append(selection, pd)
		}
	}

	return selection, nil
}

func selectPackages(args []string) error {
	ws, err := loadWorkspace()
	if err != nil {
		return err
	}

	pi, err := readPackageDefinitions(ws.wp)
	if err != nil {
		return err
	}

	selection, err := packageRangesToFlatSelection(pi, args)
	if err != nil {
		return err
	}

	conftab, err := readConftab(path.Join(ws.absPrivateDir,
		conftabFilename))
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		conftab = newConftab()
	}

	return generateAndBootstrapPackages(ws, pi, selection, conftab)
}

// selectCmd represents the select command
var selectCmd = &cobra.Command{
	Use:   "select package_range...",
	Short: "Choose one or more packages to work on",
	Args:  cobra.MinimumNArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		if err := selectPackages(args); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(selectCmd)

	selectCmd.Flags().SortFlags = false
	addQuietFlag(selectCmd)
	addPkgPathFlag(selectCmd)
	addWorkspaceDirFlag(selectCmd)
	addNoBootstrapFlag(selectCmd)
}
