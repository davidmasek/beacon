package monitor

import (
	"context"
	"errors"
	"time"

	"io"
	"net/http"
	"slices"
	"strings"

	"github.com/davidmasek/beacon/conf"
	"github.com/davidmasek/beacon/logging"
	"github.com/davidmasek/beacon/storage"
	"go.uber.org/zap"
)

type webConfig struct {
	Url         string   `mapstructure:"url"`
	HttpStatus  []int    `mapstructure:"status"`
	BodyContent []string `mapstructure:"content"`
}

const DEFAULT_TIMEOUT = 5

// Check websites and save the resulting HealthChecks to storage
func CheckWebServices(db storage.Storage, services []conf.ServiceConfig) error {
	logger := logging.Get()

	for _, service := range services {
		// skip disabled
		if !service.Enabled {
			continue
		}
		// skip non-website services
		if service.Url == "" {
			continue
		}

		logger.Debugw("Checking website", "service", service.Id, "check_config", service)

		timestamp := time.Now()
		serviceStatus, err := checkWebsite(&webConfig{
			Url:         service.Url,
			HttpStatus:  service.HttpStatus,
			BodyContent: service.BodyContent,
		})
		metadata := make(map[string]string)
		metadata["status"] = string(serviceStatus)
		if err != nil {
			logger.Error(err)
			metadata["error"] = err.Error()
		}
		healthCheck := &storage.HealthCheckInput{
			ServiceId: service.Id,
			Timestamp: timestamp,
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

func checkWebsite(config *webConfig) (ServiceStatus, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DEFAULT_TIMEOUT*time.Second)
	defer cancel()
	serviceStatus, err := config.checkWebsite(ctx)
	return serviceStatus, err
}

// Check website and return status
func (config *webConfig) checkWebsite(ctx context.Context) (status ServiceStatus, err error) {
	logger := logging.Get()
	req, err := http.NewRequestWithContext(ctx, "GET", config.Url, nil)
	if err != nil {
		// Error on side of Beacon, not the web server -> Error level logging
		logger.Errorw("Failed to create request", zap.Error(err))
		return STATUS_FAIL, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Debugw("Web check failed", zap.Error(err))
		return STATUS_FAIL, err
	}
	// When err is nil, resp always contains a non-nil resp.Body
	defer func() {
		closeErr := resp.Body.Close()
		if closeErr != nil {
			logger.Errorw("Web check failed - failed to close resp body", zap.Error(err), "config", config)
			err = errors.Join(err, closeErr)
			status = STATUS_FAIL
		}
	}()
	codeOk := slices.Contains(config.HttpStatus, resp.StatusCode)
	if !codeOk {
		logger.Debugw("Web check failed", "cause", "Unexpected status code", "expected", config.HttpStatus, "got", resp.StatusCode)
		return STATUS_FAIL, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Debugw("Web check failed", "cause", "Cannot read response", zap.Error(err))
		return STATUS_FAIL, err
	}
	fail := false
	for _, content := range config.BodyContent {
		contained := strings.Contains(string(body), content)
		if !contained {
			fail = true
			logger.Debugw("Web check failed", "cause", "missing content", "expected", content, "got", string(body))
			break
		}
	}
	if fail {
		return STATUS_FAIL, nil
	}
	return STATUS_OK, nil
}
