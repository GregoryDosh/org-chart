package tree

import (
	"errors"
	"fmt"
	"html"
	"strings"

	"github.com/awalterschulze/gographviz"
	wordwrap "github.com/mitchellh/go-wordwrap"
	log "gopkg.in/Sirupsen/logrus.v0"
)

var (
	// EdgeAttributes holds the default edge attributes
	EdgeAttributes = map[string]string{
		"penwidth": "3",
	}
	// RootAttributes holds the default root graph attributes
	RootAttributes = map[string]string{
		"bgcolor":  "transparent",
		"labelloc": "t",
		"fontsize": "50",
		"rankdir":  "LR",
		"overlap":  "false",
		"splines":  "ortho",
		"nodesep":  "2",
		"ranksep":  "4",
	}
)

func sanitizeName(name string) string {
	if string(name[0]) != "\"" {
		return fmt.Sprintf("\"%s\"", name)
	}
	return name
}

func sanitizeHTML(title string) string {
	return html.EscapeString(title)
}

func individualContributersAttrs(name string, title string, imagePath string) map[string]string {
	name = strings.Replace(wordwrap.WrapString(sanitizeHTML(name), 16), "\n", "<BR />", -1)
	title = strings.Replace(wordwrap.WrapString(sanitizeHTML(title), 16), "\n", "<BR />", -1)
	if imagePath != "" {
		imagePath = fmt.Sprintf(`<IMG SRC="%s" SCALE="TRUE"/>`, imagePath)
	}

	formattedLabel := fmt.Sprintf(`
    <<TABLE border="0" cellborder="0">
    <TR><TD WIDTH="144px" HEIGHT="144px">%s</TD></TR>
    <TR><TD><B><FONT POINT-SIZE="45">%s</FONT></B><BR /><FONT POINT-SIZE="30">%s</FONT></TD></TR>
    </TABLE>>
    `, imagePath, name, title)
	return map[string]string{
		"shape":     "box",
		"label":     formattedLabel,
		"fixedsize": "shape",
		"width":     "8",
		"height":    "4",
		"penwidth":  "3",
		"fillcolor": "#CCCCCC88",
		"style":     "filled",
	}
}

func directReportsAttrs(name string, title string, imagePath string) map[string]string {
	name = strings.Replace(wordwrap.WrapString(sanitizeHTML(name), 16), "\n", "<BR />", -1)
	title = strings.Replace(wordwrap.WrapString(sanitizeHTML(title), 16), "\n", "<BR />", -1)
	if imagePath != "" {
		imagePath = fmt.Sprintf(`<IMG SRC="%s" SCALE="TRUE"/>`, imagePath)
	}

	formattedLabel := fmt.Sprintf(`
    <<TABLE border="0" cellborder="0">
    <TR><TD WIDTH="144px" HEIGHT="144px">%s</TD></TR>
    <TR><TD><B><FONT POINT-SIZE="45">%s</FONT></B><BR /><FONT POINT-SIZE="30">%s</FONT></TD></TR>
    </TABLE>>
    `, imagePath, name, title)
	return map[string]string{
		"shape":     "circle",
		"label":     formattedLabel,
		"fixedsize": "shape",
		"width":     "5",
		"height":    "5",
		"penwidth":  "3",
		"fillcolor": "#88888888",
		"style":     "filled",
	}
}

// EmployeeTree contains a nested structure
// of employees with titles, names, and pictures.
type EmployeeTree struct {
	Name          string
	Title         string
	Image         string
	DirectReports []EmployeeTree
}

func traverseGraph(g *gographviz.Escape, parent string, employees *[]EmployeeTree) {
	log.Debugf("Parent %s %s", parent, employees)
	for _, employee := range *employees {
		log.Debugf("Checking out %s", employee.Name)
		if len(employee.DirectReports) > 0 {
			log.Debugf("Has DR")
			log.Debugf("Adding Node %s", employee.Name)
			err := g.AddNode(sanitizeName(parent), sanitizeName(employee.Name), directReportsAttrs(employee.Name, employee.Title, employee.Image))
			if err != nil {
				panic(err)
			}
		} else {
			log.Debugf("Has No DR")
			log.Debugf("Adding Node %s", employee.Name)
			err := g.AddNode(sanitizeName(parent), sanitizeName(employee.Name), individualContributersAttrs(employee.Name, employee.Title, employee.Image))
			if err != nil {
				panic(err)
			}
		}
		err := g.AddEdge(sanitizeName(parent), sanitizeName(employee.Name), true, EdgeAttributes)
		if err != nil {
			panic(err)
		}

		for _, directReport := range employee.DirectReports {
			log.Debugf("Adding Edge %s %s", employee.Name, directReport.Name)
			err := g.AddEdge(sanitizeName(employee.Name), sanitizeName(directReport.Name), true, EdgeAttributes)
			if err != nil {
				panic(err)
			}
			if len(directReport.DirectReports) > 0 {
				log.Debugf("Recursive Checking out %s %s", directReport.Name, directReport.DirectReports)
				log.Debugf("Adding Node %s", directReport.Name)
				err := g.AddNode(sanitizeName(employee.Name), sanitizeName(directReport.Name), directReportsAttrs(directReport.Name, directReport.Title, directReport.Image))
				if err != nil {
					panic(err)
				}
				traverseGraph(g, directReport.Name, &directReport.DirectReports)
			} else {
				log.Debugf("Adding Node %s", directReport.Name)
				err := g.AddNode(sanitizeName(employee.Name), sanitizeName(directReport.Name), individualContributersAttrs(directReport.Name, directReport.Title, directReport.Image))
				if err != nil {
					panic(err)
				}
			}
		}
	}
}

//BuildGraph builds the employee tree nodes & edges and returns a string representation
// of the graph as a .dot file for graphviz
func BuildGraph(title string, parent *EmployeeTree) (string, error) {
	g := gographviz.NewEscape()
	if parent.Name == "" {
		return "", errors.New("parent.Name cannot be empty")
	}
	if parent.Title == "" {
		parent.Title = " "
		log.Warnf("empty title for user %s", parent.Name)
	}
	if title == "" {
		title = fmt.Sprintf("Org Chart - %s", parent.Name)
	}
	g.SetName(title)
	g.SetDir(true)

	g.AddAttr(title, "label", title)
	for field, value := range RootAttributes {
		g.AddAttr(title, field, value)
	}

	if len(parent.DirectReports) > 0 {
		err := g.AddNode(sanitizeName(title), sanitizeName(parent.Name), directReportsAttrs(parent.Name, parent.Title, parent.Image))
		if err != nil {
			panic(err)
		}
	} else {
		err := g.AddNode(sanitizeName(title), sanitizeName(parent.Name), individualContributersAttrs(parent.Name, parent.Title, parent.Image))
		if err != nil {
			panic(err)
		}
	}

	traverseGraph(g, parent.Name, &parent.DirectReports)
	return g.String(), nil
}

// TestLDAP shows an example Employee Tree
var TestLDAP = EmployeeTree{
	Name:  "Top O the World",
	Title: "Everything",
	DirectReports: []EmployeeTree{
		{
			Name:  "IC1",
			Title: "Director Something",
			Image: "images/ic.png",
		},
		{
			Name:  "Manager 1",
			Title: "Director Something Else",
			Image: "images/dr.png",
			DirectReports: []EmployeeTree{
				{
					Name:  "IC2",
					Title: "Individual Contributer",
					Image: "images/ic.png",
				},
				{
					Name:  "IC3",
					Title: "Individual Contributer",
					Image: "images/ic.png",
				},
				{
					Name:  "IC4",
					Title: "Individual Contributer",
					Image: "images/ic.png",
				},
				{
					Name:  "Manager 2",
					Title: "That Manager!?",
					Image: "images/dr.png",
					DirectReports: []EmployeeTree{
						{
							Name:  "IC5",
							Title: "Individual Contributer",
							Image: "images/ic.png",
						},
						{
							Name:  "Manager3",
							Title: "Which manager?",
							Image: "images/dr.png",
							DirectReports: []EmployeeTree{
								{Name: "IC6", Title: "Some Title", Image: "images/ic.png"},
								{Name: "IC7", Title: "Some Title Again", Image: "images/ic.png"},
							},
						},
					},
				},
			},
		},
		{
			Name:  "IC8",
			Title: "Director In Title but not Function or Something in General",
			Image: "images/ic.png",
		},
		{
			Name:  "Manager 4",
			Title: "Director Something in Specific",
			Image: "images/dr.png",
			DirectReports: []EmployeeTree{
				{Name: "IC9", Title: "It'll be fine?", Image: "images/ic.png"},
			},
		},
	},
}
