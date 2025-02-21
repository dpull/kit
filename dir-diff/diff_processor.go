package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DiffProcessor handles processing of existing diff files
type DiffProcessor struct {
	InputPath  string
	OutputPath string
}

// NewDiffProcessor creates a new DiffProcessor instance
func NewDiffProcessor(inputPath, outputPath string) *DiffProcessor {
	return &DiffProcessor{
		InputPath:  inputPath,
		OutputPath: outputPath,
	}
}

// ProcessDiffFile processes an existing diff file and outputs a simplified version
func (p *DiffProcessor) ProcessDiffFile() error {
	// Read input file
	input, err := os.ReadFile(p.InputPath)
	if err != nil {
		return fmt.Errorf("error reading input file: %v", err)
	}

	// Process the diff content
	simplified := simplifyAndFilterDiff(string(input))

	// Create output directory if it doesn't exist
	outputDir := filepath.Dir(p.OutputPath)
	if outputDir != "." {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("error creating output directory: %v", err)
		}
	}

	// Write output
	if err := os.WriteFile(p.OutputPath, []byte(simplified), 0644); err != nil {
		return fmt.Errorf("error writing output file: %v", err)
	}

	return nil
}

// simplifyAndFilterDiff processes a diff and returns only file names and non-include changed lines
func simplifyAndFilterDiff(diffContent string) string {
	var result strings.Builder
	var currentFile string
	var inHunk bool

	scanner := bufio.NewScanner(strings.NewReader(diffContent))
	for scanner.Scan() {
		line := scanner.Text()

		// Handle file headers
		if strings.HasPrefix(line, "---") {
			continue // Skip --- line
		}
		if strings.HasPrefix(line, "+++") {
			// Extract and store filename
			parts := strings.Fields(line)
			if len(parts) > 1 {
				currentFile = parts[1]
				result.WriteString(fmt.Sprintf("File: %s\n", currentFile))
			}
			continue
		}

		// Skip hunk headers
		if strings.HasPrefix(line, "@@") {
			inHunk = true
			continue
		}

		// Handle content lines
		if inHunk {
			if strings.HasPrefix(line, "+") || strings.HasPrefix(line, "-") {
				// Only write non-include lines
				if !isIncludeLine(line[1:]) {
					result.WriteString(fmt.Sprintf("%s\n", line))
				}
			}
		}
	}

	return result.String()
}
