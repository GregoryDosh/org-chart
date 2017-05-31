package main

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/awalterschulze/gographviz"
	wordwrap "github.com/mitchellh/go-wordwrap"
)

var edgeAttributes = map[string]string{
	"penwidth": "3",
}

func individualContributersAttrs(name string, title string, imagePath string) map[string]string {
	name = strings.Replace(wordwrap.WrapString(name, 16), "\n", "<BR />", -1)
	title = strings.Replace(wordwrap.WrapString(title, 16), "\n", "<BR />", -1)
	formattedLabel := fmt.Sprintf(`
    <<TABLE border="0" cellborder="0">
    <TR><TD><IMG SRC="%s" SCALE="TRUE"/></TD></TR>
    <TR><TD><B><FONT POINT-SIZE="30">%s</FONT></B><BR /><FONT POINT-SIZE="20">%s</FONT></TD></TR>
    </TABLE>>
    `, imagePath, name, title)
	return map[string]string{
		"shape":     "box",
		"label":     formattedLabel,
		"fixedsize": "shape",
		"width":     "3",
		"height":    "2",
		"penwidth":  "3",
	}
}

func directReportsAttrs(name string, title string, imagePath string) map[string]string {
	name = strings.Replace(wordwrap.WrapString(name, 16), "\n", "<BR />", -1)
	title = strings.Replace(wordwrap.WrapString(title, 16), "\n", "<BR />", -1)
	formattedLabel := fmt.Sprintf(`
    <<TABLE border="0" cellborder="0">
    <TR><TD><IMG SRC="%s" SCALE="TRUE"/></TD></TR>
    <TR><TD><B><FONT POINT-SIZE="30">%s</FONT></B><BR /><FONT POINT-SIZE="20">%s</FONT></TD></TR>
    </TABLE>>
    `, imagePath, name, title)
	return map[string]string{
		"shape":     "invhouse",
		"label":     formattedLabel,
		"fixedsize": "shape",
		"width":     "3",
		"height":    "3",
		"penwidth":  "3",
	}
}

type employee struct {
	name          string
	title         string
	image         string
	directReports []employee
}

var testLDAP = []employee{
	{
		name:  "IC1",
		title: "Director Something",
		image: "images/ic.png",
	},
	{
		name:  "Manager 1",
		title: "Director Something Else",
		image: "images/dr.png",
		directReports: []employee{
			{
				name:  "IC2",
				title: "Individual Contributer",
				image: "images/ic.png",
			},
			{
				name:  "IC3",
				title: "Individual Contributer",
				image: "images/ic.png",
			},
			{
				name:  "IC4",
				title: "Individual Contributer",
				image: "images/ic.png",
			},
			{
				name:  "Manager 2",
				title: "That Manager!?",
				image: "images/dr.png",
				directReports: []employee{
					{
						name:  "IC5",
						title: "Individual Contributer",
						image: "images/ic.png",
					},
					{
						name:  "Manager3",
						title: "Which manager?",
						image: "images/dr.png",
						directReports: []employee{
							{name: "IC6", title: "Some Title", image: "images/ic.png"},
							{name: "IC7", title: "Some Title Again", image: "images/ic.png"},
						},
					},
				},
			},
		},
	},
	{
		name:  "IC8",
		title: "BestBuy Director Something in General",
		image: "images/ic.png",
	},
	{
		name:  "Manager 4",
		title: "Director Something in Specific",
		image: "images/dr.png",
		directReports: []employee{
			{name: "IC9", title: "No-one Cares", image: "images/ic.png"},
		},
	},
}

func traverseGraph(g *gographviz.Escape, parent string, employees *[]employee) {
	// fmt.Println("Parent", parent, employees)
	for _, employee := range *employees {
		// fmt.Println("Checking out", employee.name)
		if len(employee.directReports) > 0 {
			// fmt.Println("Has DR")
			// fmt.Println("Adding Node", employee.name)
			g.AddNode(parent, employee.name, directReportsAttrs(employee.name, employee.title, employee.image))
		} else {
			// fmt.Println("Has No DR")
			// fmt.Println("Adding Node", employee.name)
			g.AddNode(parent, employee.name, individualContributersAttrs(employee.name, employee.title, employee.image))
		}
		g.AddEdge(parent, employee.name, true, edgeAttributes)
		for _, directReport := range employee.directReports {
			// fmt.Println("Adding Edge", employee.name, directReport.name)
			g.AddEdge(employee.name, directReport.name, true, edgeAttributes)
			if len(directReport.directReports) > 0 {
				// fmt.Println("Recursive Checking out", directReport.name, directReport.directReports)
				// fmt.Println("Adding Node", directReport.name)
				g.AddNode(employee.name, directReport.name, directReportsAttrs(directReport.name, directReport.title, directReport.image))
				traverseGraph(g, directReport.name, &directReport.directReports)
			} else {
				// fmt.Println("Adding Node", directReport.name)
				g.AddNode(employee.name, directReport.name, individualContributersAttrs(directReport.name, directReport.title, directReport.image))
			}
		}
	}
}

func main() {
	g := gographviz.NewEscape()
	g.SetName("Org Chart")
	g.SetDir(true)

	g.AddNode("Top", "MvO", directReportsAttrs("MvO", "Sr. Director", "images/user.jpg"))
	traverseGraph(g, "MvO", &testLDAP)
	s := g.String()

	ioutil.WriteFile("graph.dot", []byte(s), 0644)
}
