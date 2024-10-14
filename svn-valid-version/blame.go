package main

import (
	"encoding/xml"
	"fmt"
)

type Blame struct {
	XMLName xml.Name `xml:"blame"`
	Entries []Entry  `xml:"target>entry"`
}

type Entry struct {
	Commit Commit `xml:"commit"`
}

type Commit struct {
	Revision string `xml:"revision,attr"`
}

func getVersion(blame string) ([]string, error) {
	var blameData Blame

	// Parse the XML data
	if err := xml.Unmarshal([]byte(blame), &blameData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal XML: %w", err)
	}

	// Extract the revision numbers
	var revisions []string
	for _, entry := range blameData.Entries {
		revisions = append(revisions, entry.Commit.Revision)
	}

	return revisions, nil
}
