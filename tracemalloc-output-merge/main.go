package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

func readFile(path string) (map[string]int, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	result := make(map[string]int)

	scanner := bufio.NewScanner(file)
	var mallocSize int
	var stackTrace strings.Builder

	prefix := "malloc size:"
	suffix := ", stack:"

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
			// 处理下一个 malloc
			if mallocSize > 0 {
				result[stackTrace.String()] += mallocSize
			}
			mallocSize = 0
			stackTrace.Reset()
		}
	}
	// 处理最后一个 malloc
	if mallocSize > 0 {
		result[stackTrace.String()] += mallocSize
	}

	return result, nil
}

func main() {
	if len(os.Args) <= 1 {
		fmt.Println("usage: exe file")
		return
	}

	result, err := readFile(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}

	var sorted []struct {
		key   string
		value int
	}
	for k, v := range result {
		sorted = append(sorted, struct {
			key   string
			value int
		}{k, v})
	}

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].value > sorted[j].value
	})

	file, err := os.Create(os.Args[2])
	if err != nil {
		fmt.Println("Failed to create file!")
		return
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	for _, kv := range sorted {
		fmt.Fprintf(writer, "malloc size:%d\n%s\n\n", kv.value, kv.key)
	}
	writer.Flush()
}
