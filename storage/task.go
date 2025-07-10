package storage

type TaskStatus string

const (
	TASK_OK       TaskStatus = "OK"
	TASK_ERROR    TaskStatus = "ERROR"
	TASK_SENTINEL TaskStatus = "SENTINEL"
)
