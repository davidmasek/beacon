package storage

import (
	"database/sql"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/davidmasek/beacon/logging"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

//go:embed create.sql
var CREATE_TABLE_QUERY string

const DUMMY_HASH = "$argon2id$v=19$m=65536,t=1,p=1$3/5E2tseeHN7AkROAl3Gvw$9a3DHFWwhpyuQp9/2t1VCESFVR/vxty/nQ1G55eeATA"

// Data needed to persist a new HealthCheck
type HealthCheckInput struct {
	ServiceId string
	Timestamp time.Time
	Metadata  map[string]string
}

// HealthCheck data, including identifier
type HealthCheck struct {
	Id        int
	ServiceId string
	Timestamp time.Time
	Metadata  map[string]string
}

type User struct {
	email string
}

type TaskInput struct {
	TaskName  string
	Status    string
	Timestamp time.Time
	Details   string
}

type Task struct {
	TaskName  string
	Status    string
	Timestamp time.Time
	Details   string
}

// DB Versions
type SchemaVersion struct {
	Version   int
	AppliedAt time.Time
}

type Storage interface {
	Close() error
	// List all distinct services
	ListServices() ([]string, error)
	// Main entrypoint for adding data
	AddHealthCheck(healthCheckInput *HealthCheckInput) error
	// List existing healthchecks
	LatestHealthChecks(serviceID string, limit int) ([]*HealthCheck, error)
	// Convenience method to get return healthcheck (possibly nil)
	LatestHealthCheck(serviceID string) (*HealthCheck, error)
	// Store heartbeat and return the stored value or error
	RecordHeartbeat(serviceID string, timestamp time.Time) (string, error)
	// Return sorted list of timestamps or error
	GetLatestHeartbeats(serviceID string, limit int) ([]time.Time, error)
	// Create new user
	CreateUser(email string, password string) error
	// Get user if email and password match
	ValidateUser(email string, password string) (*User, error)
	// Log new task run
	CreateTaskLog(taskInput TaskInput) error
	// Get latest task log.
	LatestTaskLog(taskName string) (*Task, error)
	// Get latest task log with given status.
	LatestTaskLogWithStatus(taskName string, status string) (*Task, error)
	// List all schema versions present
	ListSchemaVersions() ([]SchemaVersion, error)
}

// https://www.sqlite.org/lang_select.html#limitoffset
const NO_LIMIT int = -1
const TIME_FORMAT = time.RFC3339

var ErrEmailAlreadyUsed = errors.New("email already used")

func NewTestDb(t *testing.T) Storage {
	db, err := NewSQLStorage(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	return db
}

type SQLStorage struct {
	db *sql.DB
}

func (s *SQLStorage) CreateTaskLog(taskInput TaskInput) error {
	timestampStr := taskInput.Timestamp.UTC().Format(TIME_FORMAT)
	_, err := s.db.Exec(
		"INSERT INTO task_logs (task_name, status, timestamp, details) VALUES (?, ?, ?, ?)",
		taskInput.TaskName,
		taskInput.Status,
		timestampStr,
		taskInput.Details,
	)
	return err
}

func (s *SQLStorage) LatestTaskLogWithStatus(taskName string, status string) (*Task, error) {
	var timestampStr string
	var details string
	var taskNameDb string
	var statusDb string

	var err error
	// status should never be "" so use empty string to mean any status
	if status == "" {
		err = s.db.QueryRow(
			`SELECT timestamp, status, details, task_name
		 FROM task_logs
		 WHERE task_name = ?
		 ORDER BY timestamp DESC
		 LIMIT 1`,
			taskName,
		).Scan(&timestampStr, &statusDb, &details, &taskNameDb)
	} else {
		err = s.db.QueryRow(
			`SELECT timestamp, status, details, task_name
		 FROM task_logs
		 WHERE task_name = ? AND status = ?
		 ORDER BY timestamp DESC
		 LIMIT 1`,
			taskName, status,
		).Scan(&timestampStr, &statusDb, &details, &taskNameDb)
	}

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No logs found
		}
		return nil, err // Query error
	}

	timestamp, err := time.Parse(TIME_FORMAT, timestampStr)
	if err != nil {
		return nil, err
	}

	return &Task{TaskName: taskNameDb, Status: statusDb,
		Timestamp: timestamp, Details: details}, nil

}

func (s *SQLStorage) LatestTaskLog(taskName string) (*Task, error) {
	return s.LatestTaskLogWithStatus(taskName, "")
}

func (s *SQLStorage) CreateUser(email string, password string) error {
	if password == "" {
		return fmt.Errorf("cannot create user with empty password")
	}
	hashedPassword, err := GenerateFromPassword(password)
	if err != nil {
		return err
	}
	_, err = s.db.Exec("INSERT INTO users (email, password_hash) VALUES (?, ?)",
		email,
		hashedPassword,
	)
	return err
}

func (s *SQLStorage) ValidateUser(email string, inputPassword string) (*User, error) {
	query := `SELECT password_hash FROM users WHERE email = ?`

	// QueryRow is used because we expect at most one result
	row := s.db.QueryRow(query, email)

	var referenceHash string
	err := row.Scan(&referenceHash)
	if err != nil {
		if err == sql.ErrNoRows {
			// no user found, do a comparison to simulate same timing
			// as if user found
			_, err = ComparePasswordAndHash(inputPassword, DUMMY_HASH)
			// typically nil,nil
			return nil, err
		}
		return nil, err
	}

	match, err := ComparePasswordAndHash(inputPassword, referenceHash)
	if err != nil {
		return nil, err
	}
	if !match {
		return nil, nil
	}
	return &User{email: email}, nil
}

// List all distinct services, sorted alphabetically
func (s *SQLStorage) ListServices() ([]string, error) {
	rows, err := s.db.Query(`SELECT DISTINCT service_id FROM health_checks ORDER BY service_id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	services := make([]string, 0)
	for rows.Next() {
		var serviceID string
		if err := rows.Scan(&serviceID); err != nil {
			return nil, err
		}
		services = append(services, serviceID)
	}
	return services, nil
}

func (s *SQLStorage) Close() error {
	return s.db.Close()
}

func (s *SQLStorage) RecordHeartbeat(serviceId string, timestamp time.Time) (string, error) {
	input := HealthCheckInput{
		ServiceId: serviceId,
		Timestamp: timestamp,
		Metadata:  nil,
	}
	err := s.AddHealthCheck(&input)
	if err != nil {
		return "", err
	}
	timestampStr := timestamp.UTC().Format(TIME_FORMAT)
	return timestampStr, err
}

func (s *SQLStorage) GetLatestHeartbeats(serviceId string, limit int) ([]time.Time, error) {
	timestamps := make([]time.Time, 0)
	rows, err := s.db.Query("SELECT timestamp FROM health_checks WHERE service_id = ? ORDER BY timestamp DESC LIMIT ?", serviceId, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var timestampStr string
		err := rows.Scan(&timestampStr)
		if err != nil {
			return nil, err
		}
		timestamp, err := time.Parse(TIME_FORMAT, timestampStr)
		if err != nil {
			return nil, err
		}
		timestamps = append(timestamps, timestamp)
	}
	return timestamps, nil
}

func (s *SQLStorage) LatestHealthChecks(serviceId string, limit int) ([]*HealthCheck, error) {
	rows, err := s.db.Query(`
	SELECT
		id,
		service_id,
		timestamp,
		metadata
	FROM
		health_checks
	WHERE
		service_id = ?
	ORDER BY
		timestamp DESC
	LIMIT
		?
	`, serviceId, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	healthChecks := make([]*HealthCheck, 0)
	for rows.Next() {
		healthCheck, err := rowToHealthCheck(rows)
		if err != nil {
			return nil, err
		}
		healthChecks = append(healthChecks, healthCheck)
	}
	return healthChecks, nil
}

func (s *SQLStorage) AddHealthCheck(healthCheckInput *HealthCheckInput) error {
	timestampStr := healthCheckInput.Timestamp.UTC().Format(TIME_FORMAT)
	metadataStr, err := json.Marshal(healthCheckInput.Metadata)
	if err != nil {
		return err
	}
	_, err = s.db.Exec("INSERT INTO health_checks (service_id, timestamp, metadata) VALUES (?, ?, ?)",
		healthCheckInput.ServiceId,
		timestampStr,
		metadataStr,
	)
	return err
}

// Convert DB row to HealthCheck. rows must be non-empty (call rows.Next() before passing and check return value)
func rowToHealthCheck(rows *sql.Rows) (*HealthCheck, error) {
	event := &HealthCheck{}
	var timestampStr, metadataStr string
	err := rows.Scan(&event.Id, &event.ServiceId, &timestampStr, &metadataStr)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(metadataStr), &event.Metadata)
	if err != nil {
		return nil, err
	}
	// make sure .Metadata is always initialized so we don't have to check for nil all the time
	if event.Metadata == nil {
		event.Metadata = make(map[string]string)
	}
	event.Timestamp, err = time.Parse(TIME_FORMAT, timestampStr)
	if err != nil {
		return nil, err
	}
	return event, nil
}

// Return (last_event, nil) if event found and sucesfully retrieved.
// Return (nil, err) on error.
// Return (nil, nil) if no events found.
func (s *SQLStorage) LatestHealthCheck(serviceID string) (*HealthCheck, error) {
	healthChecks, err := s.LatestHealthChecks(serviceID, 1)
	if err != nil {
		return nil, err
	}
	if len(healthChecks) == 0 {
		return nil, nil
	}
	return healthChecks[0], nil
}

func NewSQLStorage(path string) (*SQLStorage, error) {
	// Create parent directory if needed
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(CREATE_TABLE_QUERY)
	if err != nil {
		return nil, err
	}

	return &SQLStorage{db}, nil
}

// Database setup (SQLite in this case).
//
// Pass empty dbPath for default location.
// If BEACON_DB is set that will be used.
// Otherwise homedir/beacon.db is used.
func InitDB(dbPath string) (Storage, error) {
	logger := logging.Get()
	// If not specified try loading from env variable.
	// Always falling back to env var makes the path
	// rewritable independent on how config is loaded.
	// ---
	// Alternatively we could pass config as usual with *conf.Config
	// but if that would not be set to read from env variables
	// (due to an error) than we could modify the wrong database.
	// Due to high impact of a potential mistake we use
	// the somewhat non-systematic behavior of directly accessing
	// env here instead of using Viper.
	if dbPath == "" {
		logger.Info("Trying to load DB path from ENV BEACON_DB")
		dbPath = os.Getenv("BEACON_DB")
	}
	// if still not specified use homedir/beacon.db
	if dbPath == "" {
		homedir, err := os.UserHomeDir()
		logger.Infow("Using default DB path", "homedir", homedir)
		if err != nil {
			return nil, err
		}
		dbPath = filepath.Join(homedir, "beacon.db")
	}
	logger.Infow("Initializing DB", "path", dbPath)
	db, err := NewSQLStorage(dbPath)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (s *SQLStorage) ListSchemaVersions() ([]SchemaVersion, error) {
	logger := logging.Get()
	rows, err := s.db.Query(`SELECT version, applied_at FROM schema_version ORDER BY version DESC`)
	if err != nil {
		logger.Errorw("Failed query", zap.Error(err))
		return nil, err
	}
	defer rows.Close()
	versions := make([]SchemaVersion, 0)
	for rows.Next() {
		var timestampStr string
		var version int
		if err := rows.Scan(&version, &timestampStr); err != nil {
			logger.Errorw("Failed scan", zap.Error(err))
			return nil, err
		}
		timestamp, err := time.Parse(TIME_FORMAT, timestampStr)
		if err != nil {
			logger.Errorw("Failed parsing time", "timestampStr", timestampStr, zap.Error(err))
			return nil, err
		}
		versions = append(versions, SchemaVersion{version, timestamp})
	}
	return versions, nil
}
