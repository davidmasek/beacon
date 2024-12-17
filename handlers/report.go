package handlers

import (
	"log"
	"time"

	"github.com/davidmasek/beacon/storage"
)

func GenerateReport(db storage.Storage) ([]ServiceReport, error) {
	reports := make([]ServiceReport, 0)

	services, err := db.ListServices()
	if err != nil {
		return nil, err
	}

	checkConfig := ServiceChecker{
		Timeout: 24 * time.Hour,
	}

	for _, service := range services {
		log.Println("Checking service", service)
		healthCheck, err := db.LatestHealthCheck(service)
		if err != nil {
			// TODO: should probably still include in the report with some explanation
			log.Println("[ERROR]", err)
			continue
		}
		serviceStatus, err := checkConfig.GetServiceStatus(healthCheck)
		if err != nil {
			// TODO: should probably still include in the report with some explanation
			log.Println("[ERROR]", err)
			continue
		}
		log.Println(" - Service status:", serviceStatus)

		reports = append(reports, ServiceReport{
			ServiceId: service, ServiceStatus: serviceStatus, LatestHealthCheck: healthCheck,
		})
	}

	return reports, nil
}
