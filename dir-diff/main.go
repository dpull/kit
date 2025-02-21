package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	// Define command line flags
	ue4Dir := flag.String("ue4", "", "Path to UE4 directory")
	ue5Dir := flag.String("ue5", "", "Path to UE5 directory")
	diffFile := flag.String("diff", "", "Input diff file path")
	outputPath := flag.String("output", "", "Output file path")
	flag.Parse()

	// Validate parameters based on mode
	if *diffFile != "" {
		// Diff file mode
		if *outputPath == "" {
			fmt.Println("Usage: program -diff <diff_file> -output <output_file>")
			os.Exit(1)
		}
		processor := NewDiffProcessor(*diffFile, *outputPath)
		if err := processor.ProcessDiffFile(); err != nil {
			fmt.Printf("Error processing diff file: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Directory comparison mode
		if *ue4Dir == "" || *ue5Dir == "" || *outputPath == "" {
			fmt.Println("Usage: program -ue4 <ue4_dir> -ue5 <ue5_dir> -output <output_file>")
			os.Exit(1)
		}
		processor := NewProcessor(*ue4Dir, *ue5Dir, *outputPath)
		if err := processor.Process(); err != nil {
			fmt.Printf("Error processing files: %v\n", err)
			os.Exit(1)
		}
	}
}
