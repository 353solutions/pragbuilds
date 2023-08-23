package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_parseHTML(t *testing.T) {
	file, err := os.Open("testdata/builds.html")
	require.NoError(t, err, "open")
	defer file.Close()

	builds, err := parseHTML(file)
	require.NoError(t, err, "parse")
	require.Equal(t, 43, len(builds), "count")

	for i, b := range builds {
		require.NotEqualf(t, "", b.Name, "%d name", i)
		require.NotEqualf(t, 0, b.ID, "%d id", i)
	}
}
