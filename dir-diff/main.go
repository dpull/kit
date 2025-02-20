package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// isIncludeLine checks if the line is an include statement
func isIncludeLine(line string) bool {
	line = strings.TrimSpace(line)
	return strings.HasPrefix(line, "#include")
}

func main() {
	// Define command line flags
	ue4Dir := flag.String("ue4", "", "Path to UE4 directory")
	ue5Dir := flag.String("ue5", "", "Path to UE5 directory")
	outputPath := flag.String("output", "", "Output diff file path")
	flag.Parse()

	// Validate required parameters
	if *ue4Dir == "" || *ue5Dir == "" || *outputPath == "" {
		fmt.Println("Usage: program -ue4 <ue4_dir> -ue5 <ue5_dir> -output <output_file>")
		os.Exit(1)
	}

	// Create output directory if it doesn't exist
	outputDir := filepath.Dir(*outputPath)
	if outputDir != "." {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			fmt.Printf("Error creating output directory: %v\n", err)
			os.Exit(1)
		}
	}

	// Create or truncate output file
	outputFile, err := os.Create(*outputPath)
	if err != nil {
		fmt.Printf("Error creating output file: %v\n", err)
		os.Exit(1)
	}
	defer outputFile.Close()

	// Walk through UE4 directory
	err = filepath.Walk(*ue4Dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		// Only process C++ files
		ext := strings.ToLower(filepath.Ext(path))
		if !strings.HasSuffix(ext, ".h") && !strings.HasSuffix(ext, ".cpp") &&
			!strings.HasSuffix(ext, ".hpp") && !strings.HasSuffix(ext, ".hxx") &&
			!strings.HasSuffix(ext, ".cc") && !strings.HasSuffix(ext, ".c++") {
			return nil
		}

		relPath, err := filepath.Rel(*ue4Dir, path)
		if err != nil {
			return err
		}

		ue5Path := filepath.Join(*ue5Dir, relPath)
		if _, err := os.Stat(ue5Path); os.IsNotExist(err) {
			return nil // Skip if file doesn't exist in UE5
		}

		// Run diff command
		cmd := exec.Command("diff", "-u", "--strip-trailing-cr", path, ue5Path)
		output, err := cmd.Output()
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
				// Files are different, process the output
				filteredOutput := filterIncludeLines(string(output))
				if filteredOutput != "" {
					_, err := outputFile.WriteString(filteredOutput)
					if err != nil {
						return fmt.Errorf("error writing to output file: %v", err)
					}
				}
			}
			// Ignore other errors
			return nil
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error processing files: %v\n", err)
		os.Exit(1)
	}
}

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

		// Store file headers
		if strings.HasPrefix(line, "---") || strings.HasPrefix(line, "+++") {
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
