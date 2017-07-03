// Copyright (C) 2017 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

func generatePackageSources() error {

	data := templateParams{
		"PackageName":        "Test",
		"PackageDescription": "Description",
		"Copyright":          "Copyright",
		"License":            "License",
		"SOURCES":            []string{"source1.cc", "source2.cc"}}

	err := generateBuildFilesFromProjectTemplate(
		"templates/asdf/..//./application",
		"output", data)

	if err != nil {
		return err
	}

	return nil
}
