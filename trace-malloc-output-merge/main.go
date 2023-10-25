package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

type info struct {
	stackTrace string
	size       int
}

const (
	prefix = "malloc size:"
	suffix = ", stack:"
)

func readFile(path string) ([]info, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	merge := make(map[string]int, 1024*1024)

	scanner := bufio.NewScanner(file)
	var mallocSize int
	var stackTrace strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		if start := strings.Index(line, prefix); start != -1 {
			start += len(prefix)
			end := strings.Index(line[start:], suffix)
			if end >= 0 {
				numStr := line[start : start+end]
				size, err := strconv.Atoi(numStr)
				if err == nil {
					mallocSize = size
				}
			}
		} else if len(line) > 0 {
			stackTrace.WriteString(line)
			stackTrace.WriteString("\n")
		} else {
			if mallocSize > 0 {
				merge[stackTrace.String()] += mallocSize
			}
			mallocSize = 0
			stackTrace.Reset()
		}
	}
	if mallocSize > 0 {
		merge[stackTrace.String()] += mallocSize
	}

	sorted := make([]info, 0, len(merge))
	for k, v := range merge {
		sorted = append(sorted, info{stackTrace: k, size: v})
	}

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].size > sorted[j].size
	})

	return sorted, nil
}

func main() {
	if len(os.Args) <= 1 {
		fmt.Println("usage: trace-malloc-output-merge file")
		return
	}

	sorted, err := readFile(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}

	file, err := os.Create(os.Args[2])
	if err != nil {
		fmt.Println("Failed to create file!")
		return
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	for _, kv := range sorted {
		fmt.Fprintf(writer, "%s%d%s\n%s\n\n", prefix, kv.size, suffix, kv.stackTrace)
	}
	writer.Flush()
}
