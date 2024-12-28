package scheduler

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/davidmasek/beacon/handlers"
	"github.com/davidmasek/beacon/monitor"
	"github.com/davidmasek/beacon/storage"
	"github.com/spf13/viper"
)

func CheckWebServices(db storage.Storage, services map[string]*monitor.ServiceConfig) error {
	// TODO: using "legacy" approach to get this done quickly
	// should look into monitor.CheckWebsites refactor
	// and getting rid of WebConfig struct
	websites := make(map[string]monitor.WebConfig)
	for serviceId, service := range services {
		// skip disabled
		if !service.Enabled {
			continue
		}
		// skip non-website services
		if service.Url == "" {
			continue
		}

		websites[serviceId] = monitor.WebConfig{
			Url:         service.Url,
			HttpStatus:  service.HttpStatus,
			BodyContent: service.BodyContent,
		}
	}
	return monitor.CheckWebsites(db, websites)
}

func Report(db storage.Storage, config *viper.Viper) error {
	// TODO: unify with the CLI report functionality
	reports, err := handlers.GenerateReport(db)
	if err != nil {
		return err
	}

	config.SetDefault("report-name", "report")
	filename := config.GetString("report-name")

	if !strings.HasSuffix(filename, ".html") {
		filename = fmt.Sprintf("%s.html", filename)
	}

	// proceed to send email even if writing to file fails
	// as it is better if at least one of the two succeeds
	err = handlers.WriteReportToFile(reports, filename)

	// TODO: need better way so check if should send email
	sendEmail := config.IsSet("email.smtp_password")
	if sendEmail {
		var emailErr error
		server, emailErr := handlers.LoadServer(config.Sub("email"))
		if emailErr != nil {
			emailErr := fmt.Errorf("failed to load SMTP server: %w", err)
			err = errors.Join(err, emailErr)
			return err
		}
		mailer := handlers.SMTPMailer{
			Server: server,
		}
		emailErr = mailer.Send(reports, config.Sub("email"))
		err = errors.Join(err, emailErr)
	}
	return err
}

func RunSingle(db storage.Storage, config *viper.Viper) error {
	// TODO/feature FIXME
	// TODO: should we split RunSingle (or remove it)
	// and directly schedule the "scrape web"
	// and "generate reports"
	// - it might be nice since we might want to run reports at different intervals
	// - also running it separate means we reduce work done in a single task (should be nicer to debug etc)
	// - would potentially need some locking so we do not generate reports while scraping web
	// 		- but not really, since the "web scraping" should run much more often than reports ...
	//		- so we can probably just take it as normal that we don't have the "latest" data
	log.Println("Do scheduling work")
	if config.IsSet("services") {
		services, err := monitor.ParseServicesConfig(config.Sub("services"))
		if err != nil {
			return err
		}
		err = CheckWebServices(db, services)
		// TODO: might want to continue on error here
		if err != nil {
			return err
		}
	} else {
		log.Println("No services specified in config file - not checking any websites")
	}
	err := Report(db, config)
	return err
}

// Periodically call RunSingle.
//
// Will not call RunSingle again until it returns even
// if specified interval passes.
func Start(ctx context.Context, interval time.Duration, db storage.Storage, config *viper.Viper) {
	log.Println("Starting scheduler")
	startFunction(ctx, interval, func(time.Time) error {
		return RunSingle(db, config)
	})
}

func startFunction(ctx context.Context, interval time.Duration, job func(time.Time) error) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Println("Scheduler stopped.")
			return
		case t := <-ticker.C:
			// If both ctx.Done() and ticker.C are ready
			// there is no guarantee what will be selected.
			// Hence we need to check the context was cancelled here.
			if ctx.Err() != nil {
				return
			}
			err := job(t)
			if err != nil {
				log.Println("[scheduler:error]", err)
			}
		}
	}
}
