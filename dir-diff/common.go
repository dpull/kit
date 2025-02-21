package main

import (
	"bufio"
	"regexp"
	"strings"
)

// isIncludeLine checks if the line is an include statement
func isIncludeLine(line string) bool {
	line = strings.TrimSpace(line)
	return strings.HasPrefix(line, "#include")
}

var timestampRegex = regexp.MustCompile(`\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\.\d{9} \+\d{4}`)

func filterIncludeLines(diffOutput string) string {
	var result strings.Builder
	var fileHeaders []string
	var currentHunk strings.Builder
	var hasNonIncludeDiff bool
	var currentHunkHasNonIncludeDiff bool
	var inHunk bool
	var validHunks []string

	scanner := bufio.NewScanner(strings.NewReader(diffOutput))
	for scanner.Scan() {
		line := scanner.Text()

		// Store file headers (remove timestamps)
		if strings.HasPrefix(line, "---") || strings.HasPrefix(line, "+++") {
			// Remove timestamp using regex
			line = strings.TrimRight(timestampRegex.ReplaceAllString(line, ""), " \t")
			fileHeaders = append(fileHeaders, line)
			continue
		}

		// Handle hunk header
		if strings.HasPrefix(line, "@@") {
			// Process previous hunk if exists
			if inHunk && currentHunk.Len() > 0 && currentHunkHasNonIncludeDiff {
				validHunks = append(validHunks, currentHunk.String())
			}

			currentHunk.Reset()
			currentHunk.WriteString(line + "\n")
			inHunk = true
			currentHunkHasNonIncludeDiff = false
			continue
		}

		if inHunk {
			currentHunk.WriteString(line + "\n")
			// Check for differences
			if strings.HasPrefix(line, "+") || strings.HasPrefix(line, "-") {
				if !isIncludeLine(line[1:]) {
					hasNonIncludeDiff = true
					currentHunkHasNonIncludeDiff = true
				}
			}
		}
	}

	// Process the last hunk
	if inHunk && currentHunk.Len() > 0 && currentHunkHasNonIncludeDiff {
		validHunks = append(validHunks, currentHunk.String())
	}

	// Only output if there are non-include differences
	if hasNonIncludeDiff {
		// Write file headers
		for _, header := range fileHeaders {
			result.WriteString(header + "\n")
		}
		// Write all valid hunks
		for _, hunk := range validHunks {
			result.WriteString(hunk)
		}
		return result.String()
	}

	return ""
}
