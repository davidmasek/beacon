package storage

import (
	"database/sql"
	_ "embed"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed create.sql
var CREATE_TABLE_QUERY string

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
	CreateTaskLog(taskName string, status string, timestamp time.Time) error
	// Get latest task log.
	//
	// Return time.Time{}, empty status and nil error when none found.
	// You can use timestamp.IsZero() to check for time.Time{}
	LatestTaskLog(taskName string) (time.Time, string, error)
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

func (s *SQLStorage) CreateTaskLog(taskName string, status string, timestamp time.Time) error {
	timestampStr := timestamp.UTC().Format(TIME_FORMAT)
	_, err := s.db.Exec(
		"INSERT INTO task_logs (task_name, status, timestamp) VALUES (?, ?, ?)",
		taskName,
		status,
		timestampStr,
	)
	return err
}

func (s *SQLStorage) LatestTaskLog(taskName string) (time.Time, string, error) {
	var timestampStr string
	var status string

	err := s.db.QueryRow(
		`SELECT timestamp, status 
		 FROM task_logs
		 WHERE task_name = ?
		 ORDER BY timestamp DESC
		 LIMIT 1`,
		taskName,
	).Scan(&timestampStr, &status)

	if err != nil {
		if err == sql.ErrNoRows {
			return time.Time{}, "", nil // No logs found
		}
		return time.Time{}, "", err // Query error
	}

	timestamp, err := time.Parse(TIME_FORMAT, timestampStr)
	if err != nil {
		return time.Time{}, "", err
	}

	return timestamp, status, nil
}

func (s *SQLStorage) CreateUser(email string, password string) error {
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

func (s *SQLStorage) ValidateUser(email string, password string) (*User, error) {
	query := `SELECT password_hash FROM users WHERE email = ?`

	// QueryRow is used because we expect at most one result
	row := s.db.QueryRow(query, email)

	var passwordHash string
	err := row.Scan(&passwordHash)
	if err != nil {
		if err == sql.ErrNoRows {
			// no user found
			return nil, nil
		}
		return nil, err
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
		dbPath = os.Getenv("BEACON_DB")
	}
	// if still not specified use homedir/beacon.db
	if dbPath == "" {
		homedir, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		dbPath = filepath.Join(homedir, "beacon.db")
	}
	db, err := NewSQLStorage(dbPath)
	if err != nil {
		return nil, err
	}
	return db, nil
}
