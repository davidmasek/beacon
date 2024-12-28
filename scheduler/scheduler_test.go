package scheduler

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/davidmasek/beacon/monitor"
	"github.com/davidmasek/beacon/storage"
	"github.com/stretchr/testify/require"
)

func TestRunSingle(t *testing.T) {
	db := storage.NewTestDb(t)
	config, err := monitor.ExampleConfig()
	require.NoError(t, err)

	tmp, err := os.CreateTemp("", "beacon-test-report-*.html")
	require.NoError(t, err)
	tmp_file := tmp.Name()
	defer os.Remove(tmp_file)
	t.Logf("Using tmp file: %q\n", tmp_file)

	config.Set("report-name", tmp_file)
	err = RunSingle(db, config)
	require.NoError(t, err)
	require.FileExists(t, tmp_file)
}

func TestStart(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	calledCounter := 0
	t.Log("Starting...")
	startFunction(ctx, time.Microsecond, func(time.Time) error {
		t.Log("Called")
		calledCounter += 1
		cancel()
		return nil
	})
	time.Sleep(1 * time.Millisecond)
	t.Log("ctx.Err():", ctx.Err())
	// context should be done (canceled)
	require.ErrorContains(t, ctx.Err(), "canceled")
	require.Equal(t, 1, calledCounter)
}

func TestStartCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	calledCounter := 0
	t.Log("Starting...")
	startFunction(ctx, time.Microsecond, func(time.Time) error {
		t.Log("Called")
		calledCounter += 1
		return nil
	})
	time.Sleep(1 * time.Millisecond)
	require.Equal(t, 0, calledCounter)
}
