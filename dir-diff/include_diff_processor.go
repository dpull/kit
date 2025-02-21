package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type IncludeDiffProcessor struct {
	UE4Dir     string
	UE5Dir     string
	OutputPath string
}

func NewIncludeDiffProcessor(ue4Dir, ue5Dir, outputPath string) *IncludeDiffProcessor {
	return &IncludeDiffProcessor{
		UE4Dir:     ue4Dir,
		UE5Dir:     ue5Dir,
		OutputPath: outputPath,
	}
}

func (p *IncludeDiffProcessor) Process() error {
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
		return diffFile(err, info, path, p.UE4Dir, p.UE5Dir, outputFile, filterNonIncludeLines)
	})

	return err
}

func diffFile(err error, info os.FileInfo, path, ue4Dir, ue5Dir string, outputFile *os.File, filterFunc func(string) string) error {
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

	relPath, err := filepath.Rel(ue4Dir, path)
	if err != nil {
		return err
	}

	ue5Path := filepath.Join(ue5Dir, relPath)
	if _, err := os.Stat(ue5Path); os.IsNotExist(err) {
		return nil // Skip if file doesn't exist in UE5
	}

	// Run diff command
	cmd := exec.Command("diff", "-u", "--strip-trailing-cr", path, ue5Path)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			// Files are different, process the output
			filteredOutput := filterFunc(string(output))
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
}

// filterNonIncludeLines filters the diff output to retain only the lines that indicate included files.
func filterNonIncludeLines(diffOutput string) string {
	var file string
	var current strings.Builder

	scanner := bufio.NewScanner(strings.NewReader(diffOutput))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "+++") {
			file = line[4:]
			file = strings.TrimRight(timestampRegex.ReplaceAllString(file, ""), " \t")
			continue // Remove the '+' prefix
		}
		// Check if the line starts with a '+' indicating an addition in the UE5 file
		if strings.HasPrefix(line, "+") || strings.HasPrefix(line, "-") {
			if isIncludeLine(line[1:]) {
				current.WriteString(line)
				current.WriteString("\n")
				continue
			}
		}
	}
	if current.Len() == 0 {
		return ""
	}
	return file + "\n" + current.String()
}
