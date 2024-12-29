package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func setupDbPathEnv(t *testing.T) {
	dir, err := os.MkdirTemp("", "test-beacon")
	require.Nil(t, err)
	t.Log("tmp dir:", dir)
	tmp_file := filepath.Join(dir, "test.db")
	err = os.Setenv("BEACON_DB", tmp_file)
	require.Nil(t, err)
}

func TestStartServer(t *testing.T) {
	// be sure to setup DB path before interacting with Beacon
	setupDbPathEnv(t)

	serverPort := "9100"
	var outputBuffer bytes.Buffer
	rootCmd.SetOut(&outputBuffer)
	rootCmd.SetErr(&outputBuffer)
	outputBuffer.Reset()
	rootCmd.SetArgs([]string{"start", "--port", serverPort, "--stop"})
	err := rootCmd.Execute()
	require.NoError(t, err)
	output := outputBuffer.String()
	require.Contains(t, output, SERVER_SUCCESS_MESSAGE)
}

func TestListServices(t *testing.T) {
	// be sure to setup DB path before interacting with Beacon
	setupDbPathEnv(t)

	var outputBuffer bytes.Buffer
	rootCmd.SetOut(&outputBuffer)
	rootCmd.SetErr(&outputBuffer)

	// list services
	rootCmd.SetArgs([]string{"list"})
	err := rootCmd.Execute()
	require.NoError(t, err)

	// verify no services known
	// we should be running with empty DB
	// if there is something already - abort
	require.Empty(t, outputBuffer)

}
