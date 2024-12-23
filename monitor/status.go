package monitor

type ServiceStatus string

const (
	STATUS_OK   ServiceStatus = "OK"
	STATUS_FAIL ServiceStatus = "FAIL"
	// e.g. unable to decide, not enough data, error in the check
	STATUS_OTHER ServiceStatus = "OTHER"
)
