package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"sync"
)

type validVersion struct {
	versions map[string]map[string]bool
	mutex    sync.Mutex
}

func (v *validVersion) add(file string, versions []string) {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	for _, version := range versions {
		m := v.versions[version]
		if m == nil {
			m = map[string]bool{}
			v.versions[version] = m
		}
		m[file] = true
	}
}

func (v *validVersion) output(file string) error {
	csvFile, err := os.Create(file)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer csvFile.Close()

	writer := csv.NewWriter(csvFile)
	defer writer.Flush()

	header := []string{"version", "file1", "file2"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %v", err)
	}

	recordCache := make([]string, 0, 1024)

	for key, innerMap := range v.versions {
		record := recordCache
		record = append(record, key)

		for file, _ := range innerMap {
			record = append(record, file)
		}

		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write record: %v", err)
		}
	}
	return nil
}
