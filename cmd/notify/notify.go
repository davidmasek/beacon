// Read DB and use handlers to create reports and send notifications
// Heavily WIP (i.e. not working / implemented yet)

package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/davidmasek/beacon/handlers"
	"github.com/davidmasek/beacon/monitor"
	"github.com/davidmasek/beacon/status"
	"github.com/davidmasek/beacon/storage"
	"github.com/spf13/viper"
)

func main() {
	sendMail := flag.Bool("send-mail", false, "Send email notifications")
	environment := flag.String("env", "dev", "Environment to run in (for identification purposes)")
	flag.Parse()
	log.Printf("Emails enabled: %v\n", *sendMail)

	viper := viper.New()

	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error reading config file: %w", err))
	}

	// TODO: this is not the best approach I think, but its quick
	viper.Set("env", *environment)

	heartbeats := make(map[string]status.HeartbeatConfig)
	err = viper.UnmarshalKey("heartbeats", &heartbeats)
	if err != nil {
		panic(fmt.Errorf("fatal error unmarshaling config file: %w", err))
	}
	websites := make(map[string]monitor.WebConfig)
	err = viper.UnmarshalKey("websites", &websites)
	if err != nil {
		panic(fmt.Errorf("fatal error unmarshaling config file: %w", err))
	}

	db, err := storage.InitDB()
	if err != nil {
		panic(err)
	}

	enabledHandlers := []handlers.StatusHandler{handlers.LogHandler{}}
	if *sendMail {
		server, err := handlers.LoadServer(viper.Sub("email"))
		if err != nil {
			panic(fmt.Errorf("failed to load SMTP server: %w", err))
		}
		enabledHandlers = append(enabledHandlers, handlers.SMTPMailer{
			Server: server,
			Target: viper.GetString("email.to"),
			Env:    viper.GetString("env"),
		})
	} else {
		enabledHandlers = append(enabledHandlers, handlers.FakeMailer{Target: viper.GetString("email.to")})
	}

	reports := make([]handlers.ServiceReport, 0)

	for service, config := range heartbeats {
		log.Println("Checking heartbeat", service)
		log.Printf("Config: %+v\n", config)
		healthCheck, err := db.LatestHealthCheck(service)
		if err != nil {
			log.Println("[ERROR]", err)
			continue
		}
		timestamps := make([]time.Time, 0)
		if healthCheck != nil {
			timestamps = append(timestamps, healthCheck.Timestamp)
		}
		serviceStatus, err := config.GetServiceStatus(timestamps)
		if err != nil {
			log.Println("[ERROR]", err)
			continue
		}

		reports = append(reports, handlers.ServiceReport{
			ServiceId: service, ServiceStatus: serviceStatus, LatestHealthCheck: healthCheck,
		})
	}

	for service, config := range websites {
		log.Println("Checking website", service)
		// TODO: this should not create new data but only retrieve existing I think
		// the data retrieval from website can be trigered here
		// but that should be separate from getting the service status
		// - that should be retrieved from the DB
		serviceStatus, err := config.GetServiceStatus()
		log.Printf("Config: %+v\n", config)
		if err != nil {
			log.Println("[ERROR]", err)
			continue
		}
		healthCheck, err := db.LatestHealthCheck(service)
		if err != nil {
			log.Println("[ERROR]", err)
			continue
		}
		details := make(map[string]string)
		details["time"] = time.Now().UTC().Format(storage.TIME_FORMAT)
		reports = append(reports, handlers.ServiceReport{
			ServiceId: service, ServiceStatus: serviceStatus, LatestHealthCheck: healthCheck,
		})
	}

	for _, handler := range enabledHandlers {
		err := handler.Handle(reports)
		if err != nil {
			log.Printf("Handler %q failed. Error: %s", handler.Name(), err)
		}
	}
}
