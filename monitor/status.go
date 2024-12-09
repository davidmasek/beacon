package monitor

type ServiceState string

const (
	STATUS_OK ServiceState = "OK"
	// e.g. unable to decide, not enough data, error in the check
	STATUS_OTHER ServiceState = "OTHER"
	STATUS_FAIL  ServiceState = "FAIL"
)
