package monitor

import (
	"time"

	"io"
	"net/http"
	"slices"
	"strings"

	"github.com/davidmasek/beacon/logging"
	"github.com/davidmasek/beacon/storage"
	"go.uber.org/zap"
)

type WebConfig struct {
	Url         string   `mapstructure:"url"`
	HttpStatus  []int    `mapstructure:"status"`
	BodyContent []string `mapstructure:"content"`
}

func CheckWebsites(db storage.Storage, websites map[string]WebConfig) error {
	logger := logging.Get()
	for service, config := range websites {
		logger.Debugw("Checking website", "service", service, "check_config", config)
		timtestamp := time.Now()
		serviceStatus, err := config.GetServiceStatus()
		metadata := make(map[string]string)
		metadata["status"] = string(serviceStatus)
		if err != nil {
			logger.Error(err)
			metadata["error"] = err.Error()
		}
		healthCheck := &storage.HealthCheckInput{
			ServiceId: service,
			Timestamp: timtestamp,
			Metadata:  metadata,
		}
		logger.Debugw("Saving", "healthCheck", healthCheck)
		err = db.AddHealthCheck(
			healthCheck,
		)
		if err != nil {
			logger.Errorw("Unable to save HealthCheck", zap.Error(err), "healthCheck", healthCheck, "service", service)
			return err
		}
	}
	return nil

}

func (config *WebConfig) GetServiceStatus() (ServiceStatus, error) {
	logger := logging.Get()
	// TODO: we need to split this into two functions
	// - "get status from website to DB"
	// - "get ServiceStatus based on info from DB"
	resp, err := http.Get(config.Url)
	if err != nil {
		return STATUS_FAIL, err
	}
	// When err is nil, resp always contains a non-nil resp.Body
	defer resp.Body.Close()
	codeOk := slices.Contains(config.HttpStatus, resp.StatusCode)
	if !codeOk {
		logger.Debugw("Web check - Unexpected status code", "expected", config.HttpStatus, "got", resp.StatusCode)
		return STATUS_FAIL, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return STATUS_FAIL, err
	}
	fail := false
	for _, content := range config.BodyContent {
		contained := strings.Contains(string(body), content)
		if !contained {
			fail = true
			logger.Debugw("Web check - Unexpected body", "expected", contained)
			break
		}
	}
	if fail {
		return STATUS_FAIL, nil
	}
	return STATUS_OK, nil
}
