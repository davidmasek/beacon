package logging

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"sync"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
)

var onceGet sync.Once

var logger *zap.SugaredLogger

// Initialize a logger if it has not been initialized
// already and return the same instance for subsequent calls.
func Get() *zap.SugaredLogger {
	// if already initialized by other function
	if logger != nil {
		return logger
	}
	onceGet.Do(func() {
		stdout := zapcore.AddSync(os.Stdout)

		level := zap.InfoLevel
		levelEnv := os.Getenv("LOG_LEVEL")
		if levelEnv != "" {
			levelFromEnv, err := zapcore.ParseLevel(levelEnv)
			if err != nil {
				log.Println(
					fmt.Errorf("invalid level, defaulting to INFO: %w", err),
				)
			}

			level = levelFromEnv
		}

		logLevel := zap.NewAtomicLevelAt(level)

		productionCfg := zap.NewProductionEncoderConfig()
		productionCfg.TimeKey = "timestamp"
		productionCfg.EncodeTime = zapcore.ISO8601TimeEncoder

		developmentCfg := zap.NewDevelopmentEncoderConfig()
		developmentCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder

		devEncoder := zapcore.NewConsoleEncoder(developmentCfg)
		structuredEncoder := zapcore.NewJSONEncoder(productionCfg)

		var gitRevision string

		buildInfo, ok := debug.ReadBuildInfo()
		if ok {
			for _, v := range buildInfo.Settings {
				if v.Key == "vcs.revision" {
					gitRevision = v.Value
					break
				}
			}
		}

		// structured logging
		core := zapcore.NewCore(structuredEncoder, stdout, logLevel).
			With(
				[]zapcore.Field{
					zap.String("git_rev", gitRevision),
				},
			)

		if os.Getenv("BEACON_LOGS") == "dev" {
			// "dev human readable" logging
			core = zapcore.NewCore(devEncoder, stdout, logLevel)
		}

		options := []zap.Option{
			zap.AddCaller(),
			zap.AddStacktrace(zapcore.ErrorLevel),
		}
		logger = zap.New(core, options...).Sugar()
	})

	return logger
}

// Initialize or replace the current logger.
// Not concurrency safe.
// Does no cleanup - original logger is forever lost.
func InitTest(t *testing.T) *zap.SugaredLogger {
	encoderCfg := zap.NewDevelopmentEncoderConfig()
	encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	encoder := zapcore.NewConsoleEncoder(encoderCfg)
	// The provided writer does not use t.Helper() and the logged file+line
	// is not useful since it points to the TestWriter source.
	// Since we include zap.AddCaller() this is not that important,
	// as we can see the source anyway.
	writer := zaptest.NewTestingWriter(t)
	level := zap.NewAtomicLevelAt(zapcore.DebugLevel)

	_logger := zap.New(
		zapcore.NewCore(encoder, writer, level),
		// Send zap internal errors to the same writer and mark the test as failed if
		// that happens.
		zap.ErrorOutput(writer.WithMarkFailed(true)),
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	logger = _logger.Sugar()
	return logger
}

func logSomething() {
	localLogger := Get()
	localLogger.Infow("I am logging, look at me!", "foo", 42)
}
