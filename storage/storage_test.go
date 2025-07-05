package storage

import (
	"fmt"
	"maps"
	"math/rand/v2"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// DB location is settable via env variable
func TestDbPath(t *testing.T) {
	dir, err := os.MkdirTemp("", "test-beacon")
	require.Nil(t, err)
	t.Log("tmp dir:", dir)
	tmp_file := filepath.Join(dir, "test.db")
	err = os.Setenv("BEACON_DB", tmp_file)
	require.Nil(t, err)
	_, err = InitDB("")
	assert.Nil(t, err)
	assert.FileExists(t, tmp_file)
	os.Remove(tmp_file)
	os.Remove(dir)
}

// Return multiple ordered timestamps (newest first)
func setupTimestamps(t *testing.T) []time.Time {
	startTime := time.Date(2024, time.October, 10, 23, 0, 0, 0, time.UTC)
	strDurations := []string{"-1s", "-2s", "-1.5h", "-24h", "-125.5h", "-3000h"}
	timestamps := make([]time.Time, 0)
	timestamps = append(timestamps, startTime)
	for _, strDuration := range strDurations {
		duration, err := time.ParseDuration(strDuration)
		if err != nil {
			t.Fatal(err)
		}
		timestamps = append(timestamps, startTime.Add(duration))
	}
	return timestamps
}

// Basic happy path - store and get a single timestamp
func TestStoreGet(t *testing.T) {
	db := NewTestDb(t)
	defer db.Close()

	want := "2024-10-26T11:59:51Z"
	serviceID := "test-service"
	timestamp, err := time.Parse(TIME_FORMAT, want)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Record heartbeat", want)
	got, err := db.RecordHeartbeat(serviceID, timestamp)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}

	t.Log("Get last heartbeat")
	timestamps, err := db.GetLatestHeartbeats(serviceID, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(timestamps) != 1 {
		t.Fatalf("got %d, want 1", len(timestamps))
	}
	got = timestamps[0].Format(TIME_FORMAT)
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

// Test get on empty DB
func TestGetEmpty(t *testing.T) {
	db := NewTestDb(t)
	defer db.Close()

	serviceID := "test-service"
	timestamps, err := db.GetLatestHeartbeats(serviceID, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(timestamps) != 0 {
		t.Fatalf("got %d, want 0", len(timestamps))
	}
}

// Test storing multiple timestamps and getting them back
func TestStoreGetMultiple(t *testing.T) {
	db := NewTestDb(t)
	defer db.Close()

	serviceID := "test-service"

	timestamps := setupTimestamps(t)

	for _, timestamp := range timestamps {
		t.Log("Record heartbeat", timestamp)
		_, err := db.RecordHeartbeat(serviceID, timestamp)
		if err != nil {
			t.Fatal(err)
		}
	}

	gotTimestamps, err := db.GetLatestHeartbeats(serviceID, NO_LIMIT)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Checking correct number of timestamps returned")
	if len(gotTimestamps) != len(timestamps) {
		t.Fatalf("got %d, want %d timestamps", len(gotTimestamps), len(timestamps))
	}
	t.Log("Checking correct timestamps returned")
	for i, got := range gotTimestamps {
		want := timestamps[i].Format(TIME_FORMAT)
		got := got.Format(TIME_FORMAT)
		if got != want {
			t.Errorf("got %s, want %s", got, want)
		}
	}
}

// Test that the timestamps are returned in the correct order (newest first)
func TestLatestHearbeatsOrder(t *testing.T) {
	db := NewTestDb(t)
	defer db.Close()

	serviceID := "test-service"

	timestamps := setupTimestamps(t)
	// random generator with fixed seed
	r := rand.New(rand.NewPCG(12, 6))

	for _, idx := range r.Perm(len(timestamps)) {
		t.Log("Record heartbeat", timestamps[idx])
		_, err := db.RecordHeartbeat(serviceID, timestamps[idx])
		if err != nil {
			t.Fatal(err)
		}
	}

	gotTimestamps, err := db.GetLatestHeartbeats(serviceID, NO_LIMIT)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Checking correct number of timestamps returned")
	if len(gotTimestamps) != len(timestamps) {
		t.Fatalf("got %d, want %d timestamps", len(gotTimestamps), len(timestamps))
	}
	t.Log("Checking correct timestamps returned")
	for i, got := range gotTimestamps {
		want := timestamps[i].Format(TIME_FORMAT)
		got := got.Format(TIME_FORMAT)
		if got != want {
			t.Errorf("got %s, want %s", got, want)
		}
	}
}

// Return if two events have equal values. Ignores ID. Timestamps are compared as strings in TIME_FORMAT.
func eventsEqual(a *HealthCheck, b *HealthCheckInput) bool {
	check := a.ServiceId == b.ServiceId && a.Timestamp.Format(TIME_FORMAT) == b.Timestamp.Format(TIME_FORMAT)
	if !check {
		return false
	}
	return maps.Equal(a.Metadata, b.Metadata)
}

func testAddAndRetrieveEvent(t *testing.T, db Storage, event *HealthCheckInput) {
	err := db.AddHealthCheck(&HealthCheckInput{event.ServiceId, event.Timestamp, event.Metadata})
	if err != nil {
		t.Fatalf("Cannot add event %q", err)
	}
	got, err := db.LatestHealthCheck(event.ServiceId)
	if err != nil {
		t.Fatalf("Cannot retrieve event %q", err)
	}
	if !eventsEqual(got, event) {
		t.Errorf("events not equal\ngot %+v\nwant %+v", got, event)
	}
}

func TestEvents(t *testing.T) {
	db := NewTestDb(t)
	defer db.Close()
	now := time.Now().UTC()
	serviceId := "test-service"
	first := HealthCheckInput{
		ServiceId: serviceId,
		Timestamp: now,
		Metadata:  make(map[string]string),
	}
	first.Metadata["status"] = "OK"
	second := HealthCheckInput{
		ServiceId: serviceId,
		Timestamp: now.Add(time.Second),
		Metadata:  make(map[string]string),
	}
	second.Metadata["status"] = "FAIL"
	second.Metadata["cause"] = "bad thing happened"
	third := HealthCheckInput{
		ServiceId: serviceId,
		Timestamp: now.Add(time.Hour),
		Metadata:  make(map[string]string),
	}
	third.Metadata["status"] = "OK"
	third.Metadata["pond_condition"] = "ducks"

	t.Log("First event")
	testAddAndRetrieveEvent(t, db, &first)
	t.Log("Second event")
	testAddAndRetrieveEvent(t, db, &second)
	t.Log("Third event")
	testAddAndRetrieveEvent(t, db, &third)
}

// HealthCheck.Metadata should never be nil
// The goal is to simplify code and skip `.Metadata != nil` checks everywhere
func TestHealthCheckMetadataAlwaysPresent(t *testing.T) {
	db := NewTestDb(t)
	defer db.Close()
	input := HealthCheckInput{
		ServiceId: "foo",
		Timestamp: time.Now(),
		// "accidentally" leave this nil
		Metadata: nil,
	}
	err := db.AddHealthCheck(&input)
	if err != nil {
		t.Fatalf("Cannot add event %q", err)
	}
	got, err := db.LatestHealthCheck(input.ServiceId)
	if err != nil {
		t.Fatalf("Cannot retrieve event %q", err)
	}
	if got.Metadata == nil {
		t.Fatalf("got.Metadata nil, expected initialized map")
	}
}

// Helper for debugging. Dump database to a file. Assumes SQLite DB.
func DumpDatabaseToFile(db Storage, filename string) error {
	conn := db.(*SQLStorage).db
	_, err := conn.Exec(fmt.Sprintf("VACUUM INTO '%s';", filename))
	return err
}

func TestTaskLog(t *testing.T) {
	db := NewTestDb(t)
	defer db.Close()
	status := "OK"
	task := "check-web"
	otherTask := "other-task"
	timestampStr := "2024-12-26T11:22:49Z"
	timestamp, err := time.Parse(TIME_FORMAT, timestampStr)
	require.NoError(t, err)

	// log task A -> A.1
	err = db.CreateTaskLog(TaskInput{task, status, timestamp, ""})
	require.NoError(t, err)

	// get A -> A.1
	gotTask, err := db.LatestTaskLog(task)
	require.NoError(t, err)

	require.Equal(t, timestamp, gotTask.Timestamp)
	require.Equal(t, status, gotTask.Status)

	now := time.Now()
	newStatus := "FAIL"

	// log task B -> B.1
	err = db.CreateTaskLog(TaskInput{otherTask, newStatus, now, ""})
	require.NoError(t, err)

	// get task A -> A.1
	gotTask, err = db.LatestTaskLog(task)
	require.NoError(t, err)

	require.Equal(t, timestamp, gotTask.Timestamp)
	require.Equal(t, status, gotTask.Status)

	// log task A -> A.2
	err = db.CreateTaskLog(TaskInput{task, newStatus, now, ""})
	require.NoError(t, err)

	// get task A -> A.2
	gotTask, err = db.LatestTaskLog(task)
	require.NoError(t, err)

	// we do not store sub-second time info
	// even if we did time comparisons are tricky as they require
	// timezones to match even if they represent the same time
	// (i.e. two representations of the same time are not equal)
	require.WithinDuration(t, now, gotTask.Timestamp, time.Second)
	require.Equal(t, newStatus, gotTask.Status)
}

// Operations on empty database work.
func TestEmptyDb(t *testing.T) {
	db := NewTestDb(t)
	defer db.Close()

	hash, _ := GenerateFromPassword("not-relevant")
	t.Log(hash)
	t.Log("--------------")

	out, err := db.GetLatestHeartbeats("service", 10)
	require.NoError(t, err)
	require.Len(t, out, 0)

	healthCheck, err := db.LatestHealthCheck("service")
	require.NoError(t, err)
	require.Nil(t, healthCheck)

	healthChecks, err := db.LatestHealthChecks("service", 10)
	require.NoError(t, err)
	require.Len(t, healthChecks, 0)

	task, err := db.LatestTaskLog("task")
	require.NoError(t, err)
	require.Nil(t, task)

	task, err = db.LatestTaskLogWithStatus("task", "status")
	require.NoError(t, err)
	require.Nil(t, task)

	services, err := db.ListServices()
	require.NoError(t, err)
	require.Len(t, services, 0)

	user, err := db.ValidateUser("email", "password")
	require.NoError(t, err)
	require.Nil(t, user)
}

func TestUserFlow(t *testing.T) {
	db := NewTestDb(t)
	defer db.Close()

	// happy path
	err := db.CreateUser("cj@google.com", "h4xor")
	require.NoError(t, err)
	user, err := db.ValidateUser("cj@google.com", "h4xor")
	require.NoError(t, err)
	require.NotNil(t, user)
	require.Equal(t, "cj@google.com", user.email)

	// with invalid password
	user, err = db.ValidateUser("cj@google.com", "idk")
	require.NoError(t, err)
	require.Nil(t, user)

	// already existing email
	err = db.CreateUser("cj@google.com", "any")
	require.Error(t, err)

	// empty password
	err = db.CreateUser("don.john.von.lon@google.com", "")
	require.Error(t, err)
}

func TestDbSchemaIncluded(t *testing.T) {
	db := NewTestDb(t)
	defer db.Close()
	schemas, err := db.ListSchemaVersions()
	require.NoError(t, err)
	require.NotEmpty(t, schemas)
}

func TestHealthChecksSince(t *testing.T) {
	db := NewTestDb(t)
	defer db.Close()

	serviceID := "test-service"
	base := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	events := []HealthCheckInput{
		{ServiceId: serviceID, Timestamp: base.Add(-2 * time.Hour), Metadata: map[string]string{"status": "OK"}},
		{ServiceId: serviceID, Timestamp: base.Add(-1 * time.Hour), Metadata: map[string]string{"status": "FAIL"}},
		// cutoff here
		{ServiceId: serviceID, Timestamp: base, Metadata: map[string]string{"status": "OK"}},
		{ServiceId: serviceID, Timestamp: base.Add(1 * time.Hour), Metadata: map[string]string{"status": "FAIL"}},
	}

	for _, event := range events {
		err := db.AddHealthCheck(&event)
		require.NoError(t, err)
	}

	cutoff := base.Add(-5 * time.Minute)

	hc, err := db.HealthChecksSince(serviceID, cutoff)
	require.NoError(t, err)

	require.Equal(t, 2, len(hc), "expected 2 results after cutoff")

	last := hc[len(hc)-1]
	require.Equal(t, "FAIL", last.Metadata["status"])
}
