package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Processor handles the directory comparison logic
type Processor struct {
	UE4Dir     string
	UE5Dir     string
	OutputPath string
}

// NewProcessor creates a new Processor instance
func NewProcessor(ue4Dir, ue5Dir, outputPath string) *Processor {
	return &Processor{
		UE4Dir:     ue4Dir,
		UE5Dir:     ue5Dir,
		OutputPath: outputPath,
	}
}

// Process walks through directories and generates the diff
func (p *Processor) Process() error {
	// Create output directory if it doesn't exist
	outputDir := filepath.Dir(p.OutputPath)
	if outputDir != "." {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("error creating output directory: %v", err)
		}
	}

	// Create or truncate output file
	outputFile, err := os.Create(p.OutputPath)
	if err != nil {
		return fmt.Errorf("error creating output file: %v", err)
	}
	defer outputFile.Close()

	// Walk through UE4 directory
	err = filepath.Walk(p.UE4Dir, func(path string, info os.FileInfo, err error) error {
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

		relPath, err := filepath.Rel(p.UE4Dir, path)
		if err != nil {
			return err
		}

		ue5Path := filepath.Join(p.UE5Dir, relPath)
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

	return err
}
