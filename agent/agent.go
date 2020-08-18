package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/observiq/carbon/errors"
	"github.com/observiq/carbon/operator"
	_ "github.com/observiq/carbon/operator/builtin" // register operators
	"github.com/observiq/carbon/pipeline"
	"go.etcd.io/bbolt"
	"go.uber.org/zap"
)

// LogAgent is an entity that handles log monitoring.
type LogAgent struct {
	database operator.Database
	pipeline *pipeline.Pipeline

	startOnce sync.Once
	stopOnce  sync.Once

	*zap.SugaredLogger
}

// Start will start the log monitoring process.
func (a *LogAgent) Start() (err error) {
	a.startOnce.Do(func() {
		err = a.pipeline.Start()
		if err != nil {
			return
		}
		a.Info("Agent started")
	})
	return
}

// Stop will stop the log monitoring process.
func (a *LogAgent) Stop() {
	a.stopOnce.Do(func() {
		a.pipeline.Stop()
		a.database.Close()
		a.Info("Agent stopped")
	})
}

// OpenDatabase will open and create a database.
func OpenDatabase(file string) (operator.Database, error) {
	if file == "" {
		return operator.NewStubDatabase(), nil
	}

	if _, err := os.Stat(filepath.Dir(file)); err != nil {
		if os.IsNotExist(err) {
			err := os.MkdirAll(filepath.Dir(file), 0755)
			if err != nil {
				return nil, fmt.Errorf("creating database directory: %s", err)
			}
		} else {
			return nil, err
		}
	}

	options := &bbolt.Options{Timeout: 1 * time.Second}
	return bbolt.Open(file, 0666, options)
}

// NewLogAgent creates a new carbon log agent.
func NewLogAgent(cfg *Config, logger *zap.SugaredLogger, pluginDir, databaseFile string, buildParams map[string]interface{}) (*LogAgent, error) {
	database, err := OpenDatabase(databaseFile)
	if err != nil {
		return nil, errors.Wrap(err, "open database")
	}

	registry, err := operator.NewPluginRegistry(pluginDir)
	if err != nil {
		return nil, errors.Wrap(err, "load plugin registry")
	}

	buildContext := operator.BuildContext{
		PluginRegistry: registry,
		Logger:         logger,
		Database:       database,
		Parameters:     buildParams,
	}

	pipeline, err := cfg.Pipeline.BuildPipeline(buildContext)
	if err != nil {
		return nil, err
	}

	return &LogAgent{
		pipeline:      pipeline,
		database:      database,
		SugaredLogger: logger,
	}, nil
}
