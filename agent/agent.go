package agent

import (
	"fmt"
	"os"
	"path/filepath"
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
	Config    *Config
	PluginDir string
	Database  string
	*zap.SugaredLogger

	buildParams map[string]interface{}
	database    operator.Database
	pipeline    *pipeline.Pipeline
	running     bool
}

// Start will start the log monitoring process.
func (a *LogAgent) Start() error {
	if a.running {
		return nil
	}

	database, err := OpenDatabase(a.Database)
	if err != nil {
		a.Errorw("Failed to open database", zap.Any("error", err))
		return err
	}
	a.database = database

	registry, err := operator.NewPluginRegistry(a.PluginDir)
	if err != nil {
		a.Errorw("Failed to load plugin registry", zap.Any("error", err))
	}

	buildContext := operator.BuildContext{
		PluginRegistry: registry,
		Logger:         a.SugaredLogger,
		Database:       a.database,
		Parameters:     a.buildParams,
	}

	pipeline, err := a.Config.Pipeline.BuildPipeline(buildContext)
	if err != nil {
		return errors.Wrap(err, "build pipeline")
	}
	a.pipeline = pipeline

	err = a.pipeline.Start()
	if err != nil {
		return errors.Wrap(err, "Start pipeline")
	}

	a.running = true
	a.Info("Agent started")
	return nil
}

// Stop will stop the log monitoring process.
func (a *LogAgent) Stop() {
	if !a.running {
		return
	}

	a.pipeline.Stop()
	a.pipeline = nil

	a.database.Close()
	a.database = nil

	a.running = false
	a.Info("Agent stopped")
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
func NewLogAgent(cfg *Config, logger *zap.SugaredLogger, pluginDir, databaseFile string) *LogAgent {
	return &LogAgent{
		Config:        cfg,
		SugaredLogger: logger,
		PluginDir:     pluginDir,
		Database:      databaseFile,
		buildParams:   make(map[string]interface{}),
	}
}

func (a *LogAgent) WithBuildParameter(key string, value interface{}) *LogAgent {
	a.buildParams[key] = value
	return a
}
