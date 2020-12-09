package transportrequest

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/go-git/go-git/v5/plumbing/object"
	"regexp"
)

// FindLabelsInCommits ...
func FindLabelsInCommits(commits object.CommitIter, label string) ([]string, error) {
	allLabels := []string{}
	labelRegex, err := regexp.Compile(fmt.Sprintf("%s: (.*)", label))
	if err != nil {
		return []string{}, fmt.Errorf("Cannot extract label: %w", err)
	}
	err = commits.ForEach(func(c *object.Commit) error {
		labels, err := findLabels(c.Message, labelRegex)
		if err != nil {
			return fmt.Errorf("Cannot extract label '%s' from commit '%s':%w", label, c.ID(), err)
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

func findLabels(text string, labelRegex *regexp.Regexp) ([]string, error) {
	ids := []string{}
	for _, e := range labelRegex.FindAllStringSubmatch(text, -1) {
		ids = append(ids, e[1])
	}
	return piperutils.UniqueStrings(ids), nil
}
