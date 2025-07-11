package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/davidmasek/beacon/conf"
	"github.com/davidmasek/beacon/logging"
	"github.com/davidmasek/beacon/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSentinelCreatedOnlyOnce(t *testing.T) {
	logging.InitTest(t)
	db := storage.NewTestDb(t)
	defer db.Close()

	task, err := db.LatestTaskLog("report")
	require.NoError(t, err)
	require.Nil(t, task)

	now := time.Now()
	later := now.Add(time.Hour)
	err = InitializeSentinel(db, now)
	require.NoError(t, err)
	task, err = db.LatestTaskLog("report")
	require.NoError(t, err)
	require.WithinDuration(t, now, task.Timestamp, time.Second)

	err = InitializeSentinel(db, later)
	require.NoError(t, err)
	task, err = db.LatestTaskLog("report")
	require.NoError(t, err)
	require.WithinDuration(t, now, task.Timestamp, time.Second)
}

func TestSentinelCreatedOnlyIfNeeded(t *testing.T) {
	logging.InitTest(t)
	db := storage.NewTestDb(t)
	defer db.Close()

	now := time.Now()
	later := now.Add(time.Hour)
	taskStatus := "test"

	err := db.CreateTaskLog(storage.TaskInput{
		TaskName: "report", Status: taskStatus, Timestamp: now, Details: ""})
	require.NoError(t, err)

	err = InitializeSentinel(db, later)
	require.NoError(t, err)

	task, err := db.LatestTaskLog("report")
	require.NoError(t, err)
	require.NotNil(t, task)
	require.Equal(t, taskStatus, task.Status)
}

func TestShouldReport(t *testing.T) {
	logging.InitTest(t)
	db := storage.NewTestDb(t)
	defer db.Close()
	config, err := conf.ConfigFromBytes([]byte(`
timezone: "Europe/Prague"
report_time: 10
`))
	require.NoError(t, err)

	require.Equal(t, 10, config.ReportAfter)
	timezone, err := time.LoadLocation("Europe/Prague")
	require.NoError(t, err)
	require.Equal(t, timezone, config.Timezone.Location)

	now := time.Date(2020, 5, 19, 17, 30, 0, 0, timezone)
	tSameDayLater := time.Date(2020, 5, 19, 17, 30, 0, 0, timezone)
	tNextDayEarly := time.Date(2020, 5, 20, 5, 30, 0, 0, timezone)
	tNextDayNoon := time.Date(2020, 5, 20, 12, 30, 0, 0, timezone)
	tNextMonthYearly := time.Date(2020, 6, 1, 1, 0, 0, 0, timezone)
	times := []time.Time{
		now, tSameDayLater, tNextDayEarly, tNextDayNoon, tNextMonthYearly,
	}

	assertShouldReport := func(expectedValues ...bool) {
		require.Equal(t, len(times), len(expectedValues))
		for i := range times {
			reportTime := times[i]
			expected := expectedValues[i]
			got, err := ShouldReport(db, config, reportTime)
			require.NoError(t, err)
			assert.Equal(t, expected, got, reportTime)
		}
	}

	// empty DB -> should report
	assertShouldReport(true, true, true, true, true)

	err = InitializeSentinel(db, now)
	require.NoError(t, err)
	// with sentinel -> should report next day after configured time or later
	assertShouldReport(false, false, false, true, true)

	// report created @ tSameDayLater
	err = db.CreateTaskLog(
		storage.TaskInput{TaskName: "report", Status: string(storage.TASK_OK), Timestamp: tSameDayLater, Details: ""})
	require.NoError(t, err)
	// -> should report next day after configured time or later
	assertShouldReport(false, false, false, true, true)

	// report created @ tNextDayEarly
	err = db.CreateTaskLog(
		storage.TaskInput{TaskName: "report", Status: string(storage.TASK_OK), Timestamp: tNextDayEarly, Details: ""})
	require.NoError(t, err)
	// -> should not report same day again
	assertShouldReport(false, false, false, false, true)

	// report failed @ tNextDayNoon
	err = db.CreateTaskLog(
		storage.TaskInput{TaskName: "report", Status: string(storage.TASK_ERROR), Timestamp: tNextDayNoon, Details: ""})
	require.NoError(t, err)
	// -> should retry
	got, err := ShouldReport(db, config, tNextDayNoon.Add(time.Hour))
	require.NoError(t, err)
	assert.Equal(t, true, got)
}

func TestStart(t *testing.T) {
	logging.InitTest(t)
	ctx, cancel := context.WithCancel(context.Background())
	calledCounter := 0
	t.Log("Starting...")
	StartFunction(ctx, time.Microsecond, func(time.Time) error {
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
	logging.InitTest(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	calledCounter := 0
	t.Log("Starting...")
	StartFunction(ctx, time.Microsecond, func(time.Time) error {
		t.Log("Called")
		calledCounter += 1
		return nil
	})
	time.Sleep(1 * time.Millisecond)
	require.Equal(t, 0, calledCounter)
}
