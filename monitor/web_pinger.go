package monitor

import (
	"fmt"
	"log"
	"time"

	"io"
	"net/http"
	"slices"
	"strings"

	"github.com/davidmasek/beacon/storage"
	"github.com/spf13/viper"
)

type WebPinger struct {
}

// Verify that HeartbeatListener implements Monitor interface
var _ Monitor = (*WebPinger)(nil)

type WebConfig struct {
	Url         string   `mapstructure:"url"`
	HttpStatus  []int    `mapstructure:"status"`
	BodyContent []string `mapstructure:"content"`
}

func (*WebPinger) Start(db storage.Storage, viper *viper.Viper) error {
	websites := make(map[string]WebConfig)
	err := viper.UnmarshalKey("websites", &websites)
	if err != nil {
		return fmt.Errorf("fatal error unmarshaling config file: %w", err)
	}
	return CheckWebsites(db, websites)
}

func CheckWebsites(db storage.Storage, websites map[string]WebConfig) error {
	for service, config := range websites {
		log.Println("Checking website", service)
		log.Printf("Config: %+v\n", config)
		timtestamp := time.Now()
		serviceStatus, err := config.GetServiceStatus()
		metadata := make(map[string]string)
		metadata["status"] = string(serviceStatus)
		if err != nil {
			log.Println("[ERROR]", err)
			metadata["error"] = err.Error()
		}
		healthCheck := &storage.HealthCheckInput{
			ServiceId: service,
			Timestamp: timtestamp,
			Metadata:  metadata,
		}
		log.Printf("Saving %+v\n", healthCheck)
		err = db.AddHealthCheck(
			healthCheck,
		)
		if err != nil {
			log.Println("[ERROR] Unable to save HealthCheck", err)
			return err
		}
	}
	return nil

}

func (config *WebConfig) GetServiceStatus() (ServiceState, error) {
	// TODO: we need to split this into two functions
	// - "get status from website to DB"
	// - "get ServiceState based on info from DB"
	resp, err := http.Get(config.Url)
	if err != nil {
		return STATUS_FAIL, err
	}
	// When err is nil, resp always contains a non-nil resp.Body
	defer resp.Body.Close()
	codeOk := slices.Contains(config.HttpStatus, resp.StatusCode)
	if !codeOk {
		log.Printf("Expected status code %v, got %v", config.HttpStatus, resp.StatusCode)
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
			log.Printf("Expected body to contain %q, but it didn't", content)
			break
		}
	}
	if fail {
		return STATUS_FAIL, nil
	}
	return STATUS_OK, nil
}
