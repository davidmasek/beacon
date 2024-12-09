package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

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
}

// https://www.sqlite.org/lang_select.html#limitoffset
const NO_LIMIT int = -1
const TIME_FORMAT = time.RFC3339

var ErrServiceNotFound = errors.New("Storage: service not found")

type SQLStorage struct {
	db *sql.DB
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

	// Create table for storing health checks
	query := `
	CREATE TABLE IF NOT EXISTS health_checks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		service_id TEXT NOT NULL,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		metadata TEXT
	);`
	_, err = db.Exec(query)
	if err != nil {
		return nil, err
	}

	return &SQLStorage{db}, nil
}

// Database setup (SQLite in this case)
func InitDB() (Storage, error) {
	db, err := NewSQLStorage("./db/heartbeats.db")
	if err != nil {
		return nil, err
	}
	return db, nil
}
