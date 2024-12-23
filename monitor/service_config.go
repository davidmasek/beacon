package monitor

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
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
		HttpStatus:  nil,
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
			return nil, fmt.Errorf("url specified but not string, got %q", inputUrl)
		}
	}

	// TODO/feature: parse other fields

	return service, nil
}

func (sc *ServiceConfig) IsWebService() bool {
	return sc.Url != ""
}

func ParseServiceConfig(servicesConfig *viper.Viper) (map[string]*ServiceConfig, error) {
	inputs := make(map[string]map[string]interface{})
	err := servicesConfig.Unmarshal(&inputs)
	if err != nil {
		return nil, err
	}

	services := make(map[string]*ServiceConfig)
	for id, input := range inputs {
		serviceConfig, err := NewServiceConfig(id, input)
		if err != nil {
			return nil, err
		}
		services[id] = serviceConfig
	}
	return services, nil

}
