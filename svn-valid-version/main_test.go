package main

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestProc(t *testing.T) {
	file := "Source/Runtime/Core/Private/Stats/StatsFile.cpp"
	versions := validVersion{
		versions: map[string]map[string]bool{},
	}

	err := os.Chdir("")
	require.NoError(t, err)

	err = proc(file, &versions)
	require.NoError(t, err)

	require.Nil(t, versions.versions["565286"])

	fmt.Println("versions", versions.versions)
}
