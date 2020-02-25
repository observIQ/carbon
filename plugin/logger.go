package plugin

import (
	"fmt"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func init() {
	RegisterConfig("logger", &LoggerConfig{})
}

type LoggerConfig struct {
	DefaultDestinationConfig `mapstructure:",squash"`
	Level                    string
}

func (c LoggerConfig) Build(logger *zap.SugaredLogger) (Plugin, error) {
	newLogger := logger.With("plugin_type", "logger", "plugin_id", c.DefaultDestinationConfig.ID())

	if c.Level == "" {
		c.Level = "debug"
	}

	level := new(zapcore.Level)
	err := level.UnmarshalText([]byte(c.Level))
	if err != nil {
		return nil, fmt.Errorf("failed to parse level: %s", err)
	}

	var logFunc func(string, ...interface{})
	switch *level {
	case zapcore.DebugLevel:
		logFunc = newLogger.Debugw
	case zapcore.InfoLevel:
		logFunc = newLogger.Infow
	case zapcore.WarnLevel:
		logFunc = newLogger.Warnw
	case zapcore.ErrorLevel:
		logFunc = newLogger.Errorw
	default:
		return nil, fmt.Errorf("log level '%s' is unsupported", level)
	}

	plugin := &LoggerPlugin{
		DefaultDestination: c.DefaultDestinationConfig.Build(),
		config:             c,
		SugaredLogger:      newLogger,
		logFunc:            logFunc,
	}

	return plugin, nil
}

type LoggerPlugin struct {
	DefaultDestination
	config  LoggerConfig
	logFunc func(string, ...interface{})
	*zap.SugaredLogger
}

func (p *LoggerPlugin) Start(wg *sync.WaitGroup) error {
	go func() {
		defer wg.Done()

		for {
			entry, ok := <-p.DefaultDestination.Input()
			if !ok {
				// TODO flush logger?
				return
			}

			p.logFunc("Received log", "entry", entry)
		}
	}()

	return nil
}
