package adt

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"sort"
	"strings"
	"sync/atomic"

	tree "github.com/GregoryDosh/org-chart/tree_building"

	log "gopkg.in/Sirupsen/logrus.v0"
	ldap "gopkg.in/ldap.v2"
)

// ActiveDirectoryConfig holds credentials information
// to connect to Active Directory and some internal
// information about it's connection status and top level
// search user for tree traversal
type ActiveDirectoryConfig struct {
	BindAddress              string
	BindPort                 int
	BindDN                   string
	BindID                   string
	BindPassword             string
	BindDomain               string
	SearchDepth              int32
	SearchDisplayName        string
	SearchFieldName          string
	SearchFieldAltNames      []string
	SearchFieldTitle         string
	SearchFieldImage         string
	SearchFieldDirectReports string
	MaxUsers                 int32
	connected                bool
	l                        *ldap.Conn
}

// Connect creates a connection to Active Directory
func (c *ActiveDirectoryConfig) Connect() error {

	if len(c.BindID) == 0 {
		return errors.New("missing bind username")
	}

	if len(c.BindPassword) == 0 {
		return fmt.Errorf("password required for %s", c.BindID)
	}

	log.Debugf("trying to bind as %s", c.BindID)
	l, err := ldap.DialTLS("tcp", fmt.Sprintf("%s:%d", c.BindAddress, c.BindPort), &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		return err
	}

	if err = l.Bind(fmt.Sprintf("%s@%s", c.BindID, c.BindDomain), c.BindPassword); err != nil {
		return err
	}

	log.Debug("bind successful")
	c.l = l
	c.connected = true

	return nil
}

// Close disconnects connection from Active Directory
func (c *ActiveDirectoryConfig) Close() {
	c.l.Close()
}

var (
	foundUser int32
)

func getUserInfo(c *ActiveDirectoryConfig, user string, searchedDepth int32) (*tree.EmployeeTree, error) {
	t := &tree.EmployeeTree{}
	log.Debugf("On search depth %d of %d & user %d of %d.", searchedDepth, c.SearchDepth, foundUser, c.MaxUsers)
	if foundUser > c.MaxUsers {
		return t, fmt.Errorf("users found %d exceeds MaxUsers %d", foundUser, c.MaxUsers)
	}
	if searchedDepth > c.SearchDepth {
		return t, fmt.Errorf("searchedDepth %d exceeds SearchDepth %d", searchedDepth, c.SearchDepth)
	}

	searchFields := []string{c.SearchFieldName, c.SearchFieldTitle, c.SearchFieldDirectReports, c.SearchDisplayName}
	if c.SearchFieldImage != "" {
		searchFields = append(searchFields, c.SearchFieldImage)
	}
	ldapQuery := "(&(objectClass=user)(objectCategory=person)(|"
	for _, field := range append(c.SearchFieldAltNames, c.SearchFieldName) {
		ldapQuery += fmt.Sprintf("(%s=%s)", field, user)
	}
	ldapQuery += "))"

	searchRequest := ldap.NewSearchRequest(
		c.BindDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		ldapQuery,
		searchFields,
		nil,
	)

	sr, err := c.l.Search(searchRequest)
	if err != nil {
		return t, err
	}

	switch n := len(sr.Entries); n {
	case 0:
		return t, fmt.Errorf("%s not found", user)
	case 1:
		log.Debugf("Parsing %s", user)
	default:
		for _, entry := range sr.Entries {
			log.Warnf("User %s - %s found?", entry.GetAttributeValues(c.SearchFieldName), entry.GetAttributeValues(c.SearchDisplayName))
		}
		return t, fmt.Errorf("found %d results but expected 1 for %s using %s", n, user, ldapQuery)
	}

	atomic.AddInt32(&foundUser, 1)

	for _, entry := range sr.Entries {
		directReportsFound := entry.GetAttributeValues(c.SearchFieldDirectReports)
		for _, dr := range directReportsFound {
			logMessage := fmt.Sprintf("Found direct report %s", dr)
			dr = strings.Split(dr, ",")[0]
			dr = strings.Split(dr, "=")[1]
			logMessage = logMessage + fmt.Sprintf(" and cleansed to %s", dr)
			log.Debugf(logMessage)
			drTree, err := getUserInfo(c, dr, searchedDepth+1)
			if err != nil {
				return t, err
			}
			t.DirectReports = append(t.DirectReports, *drTree)
		}

		sort.Slice(t.DirectReports, func(i, j int) bool {
			return t.DirectReports[i].Name <= t.DirectReports[j].Name
		})

		if len(t.DirectReports) > 0 {
			log.Debugf("All DR Info: %s", t.DirectReports)
		}

		t.Name = entry.GetAttributeValue(c.SearchDisplayName)
		t.Title = entry.GetAttributeValue(c.SearchFieldTitle)

		rawUserPicture := entry.GetRawAttributeValue(c.SearchFieldImage)
		if len(rawUserPicture) > 0 {
			userPicture := fmt.Sprintf("images/tmp-%s.jpg", t.Name)
			if err := ioutil.WriteFile(userPicture, rawUserPicture, 0644); err != nil {
				return t, err
			}
			t.Image = userPicture
		}
	}

	return t, nil
}

// TraverseEmployeeTree traverses Active Directory for a given
// search user and returns an EmployeeTree struct with
// all the relevant information for them and their direct reports.
func TraverseEmployeeTree(c *ActiveDirectoryConfig, topUser string) (*tree.EmployeeTree, error) {
	t := &tree.EmployeeTree{}

	// Just leave if use didn't specify and hard limits
	// we don't want to abuse anything here on accident.
	if c.MaxUsers == 0 {
		return t, errors.New("warning: MaxUsers cannot be left at 0")
	}
	if c.SearchDepth == 0 {
		return t, errors.New("warning: SearchDepth cannot be left at 0")
	}

	// Validate AD Mapping
	if c.SearchDisplayName == "" {
		return t, errors.New("warning: SearchDisplayName cannot be left empty")
	}
	if c.SearchFieldName == "" {
		return t, errors.New("warning: SearchFieldName cannot be left empty")
	}
	if c.SearchFieldTitle == "" {
		return t, errors.New("warning: SearchFieldTitle cannot be left empty")
	}
	if c.SearchFieldDirectReports == "" {
		return t, errors.New("warning: SearchFieldDirectReports cannot be left empty")
	}

	// Connect to LDAP if we haven't already
	if !c.connected {
		if err := c.Connect(); err != nil {
			return t, err
		}
		defer c.Close()
	}

	// This call will be recursive
	parent, err := getUserInfo(c, topUser, 1)
	if err != nil {
		return t, err
	}
	return parent, nil
}
