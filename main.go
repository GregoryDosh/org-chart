package main

import (
	"io/ioutil"

	adt "github.com/GregoryDosh/org-chart/active_directory_tree"
	tree "github.com/GregoryDosh/org-chart/tree_building"
)

func main() {
	adConfig := &adt.ActiveDirectoryConfig{
		MaxUsers:                 500,
		BindAddress:              "",
		BindPort:                 636,
		BindDN:                   "",
		BindID:                   "",
		BindPassword:             "",
		BindDomain:               "",
		SearchDepth:              25,
		SearchDisplayName:        "",
		SearchFieldName:          "",
		SearchFieldAltNames:      []string{"cn"},
		SearchFieldTitle:         "",
		SearchFieldImage:         "",
		SearchFieldDirectReports: "",
	}

	parent, err := adt.TraverseEmployeeTree(adConfig, "<USER NAME HERE>")
	if err != nil {
		panic(err)
	}

	graph, err := tree.BuildGraph("Org Chart", parent)
	if err != nil {
		panic(err)
	}

	ioutil.WriteFile("graph.dot", []byte(graph), 0644)
}
