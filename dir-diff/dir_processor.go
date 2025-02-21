package main

import (
	"fmt"
	"os"
	"path/filepath"
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
		return diffFile(err, info, path, p.UE4Dir, p.UE5Dir, outputFile, filterIncludeLines)
	})

	return err
}
