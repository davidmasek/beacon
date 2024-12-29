package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/davidmasek/beacon/handlers"
	"github.com/davidmasek/beacon/monitor"
	"github.com/davidmasek/beacon/storage"
	"github.com/spf13/viper"
)

func ShouldCheckWebServices(db storage.Storage, config *viper.Viper, now time.Time) bool {
	if !config.IsSet("services") {
		log.Println("No services specified in config file - not checking any websites")
		return false
	}
	// TODO - should follow some config or smth
	return true
}

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

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// Decide if report should be generated and send at given time `query`
//
// Config: REPORT_TIME
// Acceptable format is RFC3339
// For example:
// - 2024-12-29T13:31:28Z
// - 2024-12-29T14:31:26+01:00
// - 2024-12-29T05:31:26-08:00
// The date part will be ignored.
func ShouldReport(db storage.Storage, config *viper.Viper, query time.Time) (bool, error) {
	// TODO: this function could use more work, building minimal functionality now

	// Provide default REPORT_TIME as 17h of local time.
	ts := time.Now()
	defaultHour := 17
	ref := time.Date(ts.Year(), ts.Month(), ts.Day(), defaultHour, 0, 0, 0, ts.Location())
	config.SetDefault("REPORT_TIME", ref.Format(storage.TIME_FORMAT))

	targetStr := config.GetString("REPORT_TIME")
	target, err := time.Parse(storage.TIME_FORMAT, targetStr)
	if err != nil {
		return false, err
	}

	// should report when close to REPORT_TIME
	timeDiffFromTarget :=
		abs(target.UTC().Hour()-query.UTC().Hour())*60 +
			abs(target.UTC().Minute()-query.UTC().Minute())

	// TODO: hardcoded
	if timeDiffFromTarget > 60 {
		return false, nil
	}

	lastTimestamp, lastStatus, err := db.LatestTaskLog("report")
	if err != nil {
		return false, err
	}
	// no previous report or failed to report -> should report
	if lastTimestamp.IsZero() || lastStatus != "OK" {
		return true, nil
	}
	// should not report if reported recently
	timeDiffFromLast := query.Sub(lastTimestamp).Abs()
	if timeDiffFromLast < 90*time.Minute {
		return false, nil
	}

	return true, nil
}

func RunSingle(db storage.Storage, config *viper.Viper, now time.Time) error {
	log.Println("Do scheduling work")

	var err error = nil
	if ShouldCheckWebServices(db, config, now) {
		services, err := monitor.ParseServicesConfig(config.Sub("services"))
		if err != nil {
			return err
		}
		log.Println("Checking web services...")
		err = CheckWebServices(db, services)
		// TODO: might want to continue on error here
		if err != nil {
			return err
		}
	}
	doReport, err := ShouldReport(db, config, now)
	if err != nil {
		return err
	}
	if doReport {
		log.Println("Reporting...")
		err = handlers.DoReportTask(db, config, now)
	}
	return err
}

// Run periodic jobs.
// Config: SCHEDULER_PERIOD (duration)
//
// Will not call run next job again until previous one returns, even
// if specified interval passes.
func Start(ctx context.Context, db storage.Storage, config *viper.Viper) {
	config.SetDefault("SCHEDULER_PERIOD", "15m")
	checkInterval := config.GetDuration("SCHEDULER_PERIOD")
	log.Printf("Starting scheduler: run each %s\n", checkInterval)
	startFunction(ctx, checkInterval, func(now time.Time) error {
		return RunSingle(db, config, now)
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
