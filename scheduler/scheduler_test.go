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
	config.Set("REPORT_TIME", time.Now().Format(storage.TIME_FORMAT))
	err = RunSingle(db, config, time.Now())
	require.NoError(t, err)
	require.FileExists(t, tmp_file)

	dat, err := os.ReadFile(tmp_file)
	require.NoError(t, err)
	require.NotEmpty(t, dat, "Report is empty")
	t.Log(string(dat))
	content := string(dat)
	require.Contains(t, content, "<html")
	// TODO: check service appears in HTML
}

func TestShouldReport(t *testing.T) {
	db := storage.NewTestDb(t)
	defer db.Close()
	config, err := monitor.ExampleConfig()
	require.NoError(t, err)

	targetStr := "2024-12-29T05:31:26-08:00"
	config.Set("REPORT_TIME", targetStr)

	// ~ minute after target, but different timezone
	queryStr := "2024-12-29T14:32:26+01:00"
	query, err := time.Parse(storage.TIME_FORMAT, queryStr)
	require.NoError(t, err)

	// empty DB and current time close to target -> should report
	doReport, err := ShouldReport(db, config, query)
	require.NoError(t, err)
	require.True(t, doReport)

	err = db.CreateTaskLog("report", "OK", query)
	require.NoError(t, err)

	// already reported -> should not report
	doReport, err = ShouldReport(db, config, query)
	require.NoError(t, err)
	require.False(t, doReport)
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
