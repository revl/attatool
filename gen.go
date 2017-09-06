// Copyright (C) 2017 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

func generatePackageSources() error {

	sources := []string{"source1.cc", "source2.cc"}

	configurePatch := "AC_MSG_NOTICE([Hello from package definition file])"

	type externalLib struct {
		Name      string
		Function  string
		OtherLibs string
	}

	//requires := []string{"liba", "libb"}
	requires := []string{"b"}

	externalLibs := []externalLib{}
	/*
		{Name: "a", Function: "afunc"},
		{"b", "bfunc", "-ldependency"}}
	*/

	pd := packageDefinition{
		packageName: "Test",
		packageType: "application",
		params: templateParams{
			"name":          "Test",
			"description":   "Description",
			"version":       "1.0.0",
			"type":          "application",
			"copyright":     "Copyright",
			"requires":      requires,
			"license":       "License",
			"sources":       sources,
			"configure_ac":  configurePatch,
			"external_libs": externalLibs}}

	/*
		err := generateBuildFilesFromProjectTemplate(
			"templates/asdf/..//./application",
			"output", pd.params)

		if err != nil {
			return err
		}
	*/

	err := generateBuildFilesFromEmbeddedTemplate(&appTemplate,
		"output-app", pd.params)

	if err != nil {
		return err
	}

	return nil
}
