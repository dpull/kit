package main

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Blame struct {
	XMLName xml.Name `xml:"blame"`
	Entries []Entry  `xml:"target>entry"`
}

type Entry struct {
	LineNumber int    `xml:"line-number,attr"`
	Commit     Commit `xml:"commit"`
}

type Commit struct {
	Revision string `xml:"revision,attr"`
}

func getVersion(blame string) (map[int]string, error) {
	var blameData Blame

	// Parse the XML data
	if err := xml.Unmarshal([]byte(blame), &blameData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal XML: %w", err)
	}

	// Extract the revision numbers
	revisions := make(map[int]string)
	for _, entry := range blameData.Entries {
		revisions[entry.LineNumber] = entry.Commit.Revision
	}

	return revisions, nil
}

func removeEmptyLine(revisions map[int]string, fileName string) error {
	ext := filepath.Ext(fileName)
	if ext != ".h" && ext != ".cpp" {
		return nil
	}

	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNumber := 1

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			// 如果是空行，从 revisions 映射中移除对应的行号
			delete(revisions, lineNumber)
		}
		lineNumber++
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}
