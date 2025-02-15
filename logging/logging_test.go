package logging

import (
	"testing"
)

func TestLogSomething(t *testing.T) {
	logger := InitTest(t)
	logger.Infow("This is a test", "foo", 42)
	logSomething()
}

func TestLogSomethingElse(t *testing.T) {
	logger := InitTest(t)
	t.Log("works")
	logger.Infow("This is the *other* test", "foo", 42)
	logSomething()
	logger.Error("oops")
	logger.Debug("Can you see me?")
}
