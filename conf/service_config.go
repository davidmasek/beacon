package conf

import (
	"fmt"
	"time"
)

type ServiceConfig struct {
	Id      string
	Timeout time.Duration
	Enabled bool
	// web only below
	Url         string
	HttpStatus  []int
	BodyContent []string
}

func defaultServiceConfig(id string) *ServiceConfig {
	return &ServiceConfig{
		Id:          id,
		Timeout:     24 * time.Hour,
		Enabled:     true,
		Url:         "",
		HttpStatus:  []int{200},
		BodyContent: nil,
	}
}

func NewServiceConfig(id string, input map[string]interface{}) (*ServiceConfig, error) {
	service := defaultServiceConfig(id)
	if input == nil {
		return service, nil
	}
	inputUrl := input["url"]
	if inputUrl != nil {
		if url, ok := inputUrl.(string); ok {
			service.Url = url
		} else {
			return nil, fmt.Errorf("[%s] url specified but not string, got %q", id, inputUrl)
		}
	}

	inputStatus := input["status"]
	if inputStatus != nil {
		if statuses, ok := inputStatus.([]interface{}); ok {
			// Create a new slice to store converted values
			var parsedStatuses []int
			for _, s := range statuses {
				if status, ok := s.(int); ok {
					parsedStatuses = append(parsedStatuses, status)
				} else {
					return nil, fmt.Errorf("[%s] invalid value in status, got %v", id, s)
				}
			}
			if len(parsedStatuses) > 0 {
				service.HttpStatus = parsedStatuses
			}
		} else {
			return nil, fmt.Errorf("[%s] invalid type for field status, got %q", id, inputStatus)
		}
	}

	inputContent := input["content"]
	if inputContent != nil {
		// Ensure inputContent is a slice of interface{}
		if contents, ok := inputContent.([]interface{}); ok {
			// Create a new slice to store converted values
			var parsedContents []string
			for _, c := range contents {
				if content, ok := c.(string); ok {
					parsedContents = append(parsedContents, content)
				} else {
					return nil, fmt.Errorf("[%s] invalid value in content, got %v", id, c)
				}
			}
			service.BodyContent = parsedContents
		} else {
			return nil, fmt.Errorf("[%s] invalid type for field content, got %q", id, inputContent)
		}
	}

	inputTimeout := input["timeout"]
	if inputTimeout != nil {
		if timeoutStr, ok := inputTimeout.(string); ok {
			duration, err := time.ParseDuration(timeoutStr)
			if err != nil {
				return nil, fmt.Errorf("[%s] invalid duration format for timeout, got %q", id, timeoutStr)
			}
			service.Timeout = duration
		} else {
			return nil, fmt.Errorf("[%s] invalid type for timeout, expected string, got %q", id, inputTimeout)
		}
	}

	inputEnabled := input["enabled"]
	if inputEnabled != nil {
		if enabled, ok := inputEnabled.(bool); ok {
			service.Enabled = enabled
		} else {
			return nil, fmt.Errorf("[%s] invalid type for enabled, expected bool, got %q", id, inputEnabled)
		}
	}

	return service, nil
}

func (sc *ServiceConfig) IsWebService() bool {
	return sc.Url != ""
}

func ParseServicesConfig(servicesConfig *Config) (map[string]*ServiceConfig, error) {
	inputs := servicesConfig.AllSettings()

	services := make(map[string]*ServiceConfig)
	for id, rawInput := range inputs {
		input, ok := rawInput.(map[string]interface{})
		// TODO: is this the correct check for empty values?
		if !ok {
			input = make(map[string]interface{})
		}
		serviceConfig, err := NewServiceConfig(id, input)
		if err != nil {
			return nil, err
		}
		services[id] = serviceConfig
	}
	return services, nil

}
