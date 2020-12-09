package main

import (
	"strings"
	"regexp"
	"fmt"
	pipergit "github.com/SAP/jenkins-library/pkg/git"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func main() {
	fmt.Printf("Hello Marcus\n")
	r, err := git.PlainOpen(".")
	if err != nil {
		fmt.Printf("ERR PlainOpen: %v\n", err)
		return
	}
	cIter, err := pipergit.LogRange(r, "github/master", "HEAD")
	if err != nil {
		fmt.Printf("ERR Log: %v\n", err)
		return
	}
	xids, err := FindLabelsInCommits(cIter, "TransportRequest")
	if err != nil {
		fmt.Printf("ERR FindLabelsInCommit: %v\n", err)
		return
	}
	fmt.Printf("XXXX xids: %v\n", xids)

/*
	err = cIter.ForEach(func(c *object.Commit) error {
		fmt.Printf("Commit ID: '%s'\nCommit Message: '%s'\n", c.ID(), c.Message)		
		return nil
	})
	if err != nil {
		fmt.Printf("ERR Iter: %v\n", err)
		return
	}
*/
/*
	cm := "TransportRequest: 12345678\nmore text\nTransportRequest: 987654321\nTransportRequest: 12345678"
	id, err := FindLabels(cm, "TransportRequest")
	if err != nil {
		fmt.Printf("ERR FindLabel: %v\n", err)
		return
	}
	fmt.Printf("ID: %s", id)
*/
/*
	cm := "TransportRequest: 12345678\nmore text\nTransportRequest: 987654321\nTransportRequest: 12345678"
	re, err := regexp.Compile(`TransportRequest: (.*)`)
	if err != nil {
		fmt.Printf("ERR Regex: %v\n", err)
		return
	}

	ids := []string{}
	result := re.FindAllStringSubmatch(cm, -1)
	fmt.Printf("Result: '%s' (%v)\n", result[0][1], result)
	for _, e := range result {
		ids = append(ids, e[1])
	}
	fmt.Printf("ids: %v\n", ids)
	uniqueIds := piperutils.UniqueStrings(ids)
	fmt.Printf("uds: %v\n", uniqueIds)
*/
}

func FindLabelsInCommits(commits object.CommitIter, label string) ([]string, error) {
	allLabels := []string{}
	labelRegex, err := regexp.Compile(fmt.Sprintf("%s: (.*)", label))
        if err != nil {
                return []string{}, fmt.Errorf("Cannot extract label: %w", err)
        }
	err = commits.ForEach(func(c *object.Commit) error {
                fmt.Printf("[MH] Commit ID: '%s'\n[MH] Commit Message: '%s'\n", c.ID(), strings.TrimSpace(c.Message))
		labels, err := FindLabels(c.Message, labelRegex)
		if err != nil {
			return fmt.Errorf("Cannot extract label '%s' from commit '%s':%w", label, c.ID, err)
		}
		if len(labels) > 1 {
			return fmt.Errorf("Found more than one labels (%s) in commit '%s': %s", label, c.ID(), labels)
		}
		allLabels = append(allLabels, labels...)
                return nil
        })
        if err != nil {
                return []string{}, fmt.Errorf("Cannot extract label: %w", err)
        }

	return piperutils.UniqueStrings(allLabels), nil
}

func FindLabels(text string, labelRegex *regexp.Regexp) ([]string, error) {
	ids := []string{}
	for _, e := range labelRegex.FindAllStringSubmatch(text, -1) {
		ids = append(ids, e[1])
	}
	return piperutils.UniqueStrings(ids), nil
}
